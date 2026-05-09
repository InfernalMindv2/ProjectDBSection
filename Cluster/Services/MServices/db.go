package MServices

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Init_DB() {

	db, err := sql.Open("mysql", "root:123456789@tcp(127.0.0.1:3306)/test")
	if err != nil {
		fmt.Println(err)
		return
	}

	DB = db

	err = DB.Ping()
	if err != nil {
		fmt.Println("DB not reachable:", err)
		return
	}

	fmt.Println("Connected to MySQL...")
}

func Create_DB(db_name string) {

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", db_name)

	_, err := DB.Exec(query)
	if err != nil {
		fmt.Println("Create DB error:", err)
		return
	}

	fmt.Println("DB Created")
}

func Use_DB(db_name string) {

	dsn := fmt.Sprintf("root:123456789@tcp(127.0.0.1:3306)/%s", db_name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("Use DB error:", err)
		return
	}

	DB = db

	err = DB.Ping()
	if err != nil {
		fmt.Println("DB switch failed:", err)
		return
	}
}

func Create_Table(table string) {

	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(50),
		age INT
	)`, table)

	_, err := DB.Exec(query)
	if err != nil {
		fmt.Println("Error creating table")
		return
	}

	fmt.Println("Table created")
}

func Insert_Record(table, name string, age int) {

	query := fmt.Sprintf("INSERT INTO %s (name, age) VALUES (?, ?)", table)

	_, err := DB.Exec(query, name, age)
	if err != nil {
		fmt.Println("Insert error")
		return
	}

	fmt.Println("Inserted")
}