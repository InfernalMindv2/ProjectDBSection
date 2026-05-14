package main

import (
	"Cluster/Services/MServices"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

var SECRET = "cluster_secret"
var NextID int


func saveID() {
	data, _ := json.MarshalIndent(NextID, "", "  ")
	os.WriteFile(idFile, data, 0644)
}

func loadID() {
	file, err := os.ReadFile(idFile)
	if err == nil {
		json.Unmarshal(file, &NextID)
	}
}


// ---------------- METADATA ----------------
var Metadata = map[string]string{}
var metaFile = "./metadata.json"
var idFile = "./id.json"
var mu sync.Mutex




func findSlaveByID(slaves []MServices.Slave, id string) MServices.Slave {
	for _, s := range slaves {
		if fmt.Sprint(s.ID) == id {
			return s
		}
	}
	return MServices.Slave{}
}

func saveMetadata() {
	data, _ := json.MarshalIndent(Metadata, "", "  ")
	os.WriteFile(metaFile, data, 0644)
}

// load metadata on startup
func loadMetadata() {
	file, err := os.ReadFile(metaFile)
	if err == nil {
		json.Unmarshal(file, &Metadata)
		fmt.Println("📦 Metadata loaded")
	}
}

// ---------------- SAFE WRAPPER ----------------

func safe(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				fmt.Println("Recovered panic:", err)
				http.Error(w, "internal server error", 500)
			}
		}()

		handler(w, r)
	}
}

// ---------------- INSERT ----------------

func InsertHandler(w http.ResponseWriter, r *http.Request) {

	var req map[string]string

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	value := req["value"]
	if value == "" {
		http.Error(w, "missing value", 400)
		return
	}

	idx :=req["id"]
	if idx == "" {
		http.Error(w, "missing id", 400)
		return
	}
	slaves := MServices.LoadConfig()

	// 1. CHECK DUPLICATE FIRST
	mu.Lock()
	if _, exists := Metadata[idx]; exists {
		mu.Unlock()
		http.Error(w, "id already exists", 409)
		return
	}
	mu.Unlock()


	if len(slaves) == 0 {
		http.Error(w, "no slaves available", 500)
		return
	}

	// shard selection
	target := MServices.SelectSlave(len(value), slaves)

	// failover if needed
	if !MServices.IsAlive(target) {
		fmt.Println("Primary slave down, switching...")
		target = MServices.Failover(slaves)
	}

	if (target == MServices.Slave{}) {
		http.Error(w, "all slaves down", 500)
		return
	}


	data := map[string]string{
		"id":    idx,
		"value": value,
		"hash":  MServices.GenerateHash(SECRET),
	}

		err := MServices.SendToSlave(target, "insert", data)
		if err != nil {
			http.Error(w, "insert failed on slave", 500)
			return
		}
	
	// 4. ONLY AFTER SUCCESS → update metadata
	mu.Lock()
	Metadata[idx] = fmt.Sprint(target.ID)
	saveMetadata()
	mu.Unlock()

	w.Write([]byte("INSERT OK"))
}

// ---------------- SELECT ----------------

func SelectHandler(w http.ResponseWriter, r *http.Request) {

	slaves := MServices.LoadConfig()

	if len(slaves) == 0 {
		http.Error(w, "no slaves available", 500)
		return
	}

	results := make(chan string, len(slaves))

	for _, s := range slaves {

		go func(slave MServices.Slave) {

			if !MServices.IsAlive(slave) {
				results <- ""
				return
			}

			resp, err := http.Get(
				fmt.Sprintf("http://%s:%s/select", slave.IP, slave.Port),
			)

			if err != nil {
				results <- ""
				return
			}
			defer resp.Body.Close()

			buf := make([]byte, 2048)
			n, _ := resp.Body.Read(buf)

			results <- string(buf[:n])

		}(s)
	}

	final := ""

	for i := 0; i < len(slaves); i++ {
		final += <-results
	}

	w.Write([]byte(final))
}

// ---------------- UPDATE ----------------

func UpdateHandler(w http.ResponseWriter, r *http.Request) {

	var req map[string]string

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	slaves := MServices.LoadConfig()

	if len(slaves) == 0 {
		http.Error(w, "no slaves", 500)
		return
	}

	slaveID, ok := Metadata[req["id"]]
	if !ok {
		http.Error(w, "record not found", 404)
		return
	}

	target := findSlaveByID(slaves, slaveID)


	if !MServices.IsAlive(target) {
		target = MServices.Failover(slaves)
	}

	data := map[string]string{
		"id":    req["id"],
		"value": req["value"],
		"hash":  MServices.GenerateHash(SECRET),
	}

	err := MServices.SendToSlave(target, "update", data)
	if err != nil {
		http.Error(w, "update failed", 500)
		return
	}

	w.Write([]byte("UPDATE OK"))
}

// ---------------- DELETE ----------------

func DeleteHandler(w http.ResponseWriter, r *http.Request) {

	var req map[string]string

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	slaves := MServices.LoadConfig()

	if len(slaves) == 0 {
		http.Error(w, "no slaves", 500)
		return
	}

	slaveID, ok := Metadata[req["id"]]
	if !ok {
		http.Error(w, "record not found", 404)
		return
	}

	target := findSlaveByID(slaves, slaveID)

	if !MServices.IsAlive(target) {
		target = MServices.Failover(slaves)
	}

	data := map[string]string{
		"id":   req["id"],
		"hash": MServices.GenerateHash(SECRET),
	}

	err := MServices.SendToSlave(target, "delete", data)
	if err != nil {
		http.Error(w, "delete failed", 500)
		return
	}

	w.Write([]byte("DELETE OK"))
}

// ---------------- MAIN ----------------

func main() {
	loadMetadata()
	loadID()

	http.HandleFunc("/insert", safe(InsertHandler))
	http.HandleFunc("/select", safe(SelectHandler))
	http.HandleFunc("/update", safe(UpdateHandler))
	http.HandleFunc("/delete", safe(DeleteHandler))

	fmt.Println("MASTER API GATEWAY running on 9000")

	http.ListenAndServe(":9000", nil)
}