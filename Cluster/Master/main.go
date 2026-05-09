package main

import (
	"Cluster/Services/MServices"
)

func main() {

	// 1- Initialize DB connection
	MServices.Init_DB()

	// 2- Create DB
	MServices.Create_DB("TestDB")

	// 3- Use DB
	MServices.Use_DB("TestDB")

	// 4- Create Table
	MServices.Create_Table("users")

	// 5- Insert Record
	MServices.Insert_Record("users", "Ali", 25)

	// 6- Replicate to slaves
	slaves := []string{
		"http://127.0.0.1:9081",
		"http://127.0.0.1:9082",
	}

	MServices.Replicate_Insert(slaves, "users", "Ali", 25)
}