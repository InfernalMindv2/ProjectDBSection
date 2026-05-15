package main

import (
	"Cluster/Services/MServices"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

var SECRET = "cluster_secret"


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
	if r.Method != http.MethodPost {
	http.Error(w, "POST only", 405)
	return
}
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



	data := map[string]string{
		"id":    idx,
		"value": value,
		"hash":  MServices.GenerateHash(SECRET),
	}

	// 1. choose shard normally
	target := MServices.SelectSlave(len(value), slaves)

	// 2. try selected shard first
	if MServices.IsAlive(target) {

	err := MServices.SendToSlave(target, "insert", data)

	if err == nil {

		fmt.Println("✅ Inserted into selected slave", target.ID)

	} else {

		// selected slave failed during request
		fmt.Println("⚠️ Selected slave failed, running failover...")

		target, err = MServices.SendWithFailover(
			slaves,
			"insert",
			data,
		)

		if err != nil {
			http.Error(w, "all slaves failed", 500)
			return
		}
	}

} else {

	// selected slave already down
	fmt.Println("⚠️ Selected slave down, running failover...")

	var err error

	target, err = MServices.SendWithFailover(
		slaves,
		"insert",
		data,
	)

	if err != nil {
		http.Error(w, "all slaves failed", 500)
		return
	}
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

	if r.Method != http.MethodGet {
	http.Error(w, "GET only", 405)
	return
}
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

			url := fmt.Sprintf(
		"http://%s:%s/select",
		slave.IP,
		slave.Port,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		results <- ""
		return
	}

	req.Header.Set(
		"X-Auth-Hash",
		MServices.GenerateHash(SECRET),
	)

	client := http.Client{}

	resp, err := client.Do(req)

			if err != nil {
				results <- ""
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
			results <- ""
			return
		}

			body, err := io.ReadAll(resp.Body)

			if err != nil {
				results <- ""
				return
			}

			results <- string(body)

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

	if r.Method != http.MethodPost {
	http.Error(w, "POST only", 405)
	return
}
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
	if req["value"] == ""{
		http.Error(w, "missing value!", 400)
		return
	}

	slaveID, ok := Metadata[req["id"]]
	if !ok {
		http.Error(w, "record not found", 404)
		return
	}

	target := findSlaveByID(slaves, slaveID)


	if !MServices.IsAlive(target) {
		http.Error(w, "The slave that has the record is under maintainance!", 503)
		return
		//target = MServices.Failover(slaves)
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
	// UPDATE METADATA
	//We dont want to edit metadata.json
		// mu.Lock()
		// Metadata[req["id"]] = fmt.Sprint(target.ID)
		// saveMetadata()
		// mu.Unlock()

	w.Write([]byte("UPDATE OK"))
}

// ---------------- DELETE ----------------

func DeleteHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
	http.Error(w, "POST only", 405)
	return
}
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
		http.Error(w, "The slave that has the record is under maintainance!", 503)
		return
		//target = MServices.Failover(slaves)
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
	// UPDATE METADATA
	mu.Lock()
	delete(Metadata, req["id"])
	saveMetadata()
	mu.Unlock()

	w.Write([]byte("DELETE OK"))
}

// ---------------- MAIN ----------------

func main() {
	loadMetadata()

	http.HandleFunc("/insert", safe(InsertHandler))
	http.HandleFunc("/select", safe(SelectHandler))
	http.HandleFunc("/update", safe(UpdateHandler))
	http.HandleFunc("/delete", safe(DeleteHandler))

	fmt.Println("MASTER API GATEWAY running on 9000")
	fs := http.FileServer(http.Dir("C:\\Project DB\\website"))
	http.Handle("/", fs)
	http.ListenAndServe(":9000", nil)
}