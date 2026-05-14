package main

import (
	"Cluster/Services/SServices"
	"fmt"
	"net/http"
)

// ---------------- SAFE WRAPPER ----------------

func safe(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				fmt.Println("Recovered panic:", err)
				http.Error(w, "internal error", 500)
			}
		}()

		handler(w, r)
	}
}

// ---------------- MAIN ----------------

func main() {

	SServices.InitDB()

	http.HandleFunc("/insert", safe(SServices.InsertHandler))
	http.HandleFunc("/select", safe(SServices.SelectHandler))
	http.HandleFunc("/update", safe(SServices.UpdateHandler))
	http.HandleFunc("/delete", safe(SServices.DeleteHandler))
	http.HandleFunc("/health", safe(SServices.HealthHandler))

	fmt.Println("🚀 Slave1 running on 9083")

	err := http.ListenAndServe(":9083", nil)
	if err != nil {
		fmt.Println("❌ Server failed:", err)
	}
}