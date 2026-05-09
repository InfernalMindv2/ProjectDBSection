package main

import (
	"Cluster/Services/SServices"
	"fmt"
	"net/http"
)

func main() {

	SServices.Init_DB()

	server := http.Server{
		Addr: "127.0.0.1:9081",
	}

	http.HandleFunc("/insert", SServices.Insert_Handler)

	fmt.Println("Slave1 running on 9081")

	server.ListenAndServe()
}