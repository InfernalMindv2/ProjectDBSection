package main

import (
	"Cluster/Services/MServices"
	"encoding/json"
	"fmt"
	"net/http"
)

var SECRET = "cluster_secret"

// ---------------- METADATA ----------------
var Metadata = map[string]string{}

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

	slaves := MServices.LoadConfig()

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
		"value": value,
		"hash":  MServices.GenerateHash(SECRET),
	}

	// metadata update
	Metadata[value] = fmt.Sprintf("stored in slave %d", target.ID)

	go func() {
		err := MServices.SendToSlave(target, "insert", data)
		if err != nil {
			fmt.Println("Retrying...")
			MServices.SendToSlave(target, "insert", data)
		}
	}()

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

	target := MServices.SelectSlave(len(req["id"]), slaves)

	if !MServices.IsAlive(target) {
		target = MServices.Failover(slaves)
	}

	data := map[string]string{
		"id":    req["id"],
		"value": req["value"],
		"hash":  MServices.GenerateHash(SECRET),
	}

	go MServices.SendToSlave(target, "update", data)

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

	target := MServices.SelectSlave(len(req["id"]), slaves)

	if !MServices.IsAlive(target) {
		target = MServices.Failover(slaves)
	}

	data := map[string]string{
		"id":   req["id"],
		"hash": MServices.GenerateHash(SECRET),
	}

	go MServices.SendToSlave(target, "delete", data)

	w.Write([]byte("DELETE OK"))
}

// ---------------- MAIN ----------------

func main() {

	http.HandleFunc("/insert", safe(InsertHandler))
	http.HandleFunc("/select", safe(SelectHandler))
	http.HandleFunc("/update", safe(UpdateHandler))
	http.HandleFunc("/delete", safe(DeleteHandler))

	fmt.Println("MASTER API GATEWAY running on 9000")

	http.ListenAndServe(":9000", nil)
}