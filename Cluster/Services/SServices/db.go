package SServices

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Init_DB() {

	// 1- Connect to MySQL (slave DB)
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/TestDB")
	if err != nil {
		fmt.Println("Error connecting slave DB:", err)
		return
	}

	DB = db

	// 2- Check connection
	err = DB.Ping()
	if err != nil {
		fmt.Println("Slave DB not reachable:", err)
		return
	}

	fmt.Println("Slave connected to DB")

	// 3- Ensure table exists (VERY IMPORTANT for your error)
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(50),
			age INT
		)
	`)

	if err != nil {
		fmt.Println("Error creating table in slave:", err)
		return
	}
}