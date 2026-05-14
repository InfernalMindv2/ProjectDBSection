package SServices

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var SECRET = "cluster_secret"

// ---------------- INIT DB (FIXED FOR DISTRIBUTED SYSTEM) ----------------

func InitDB() {

	// each slave must define its own DB name
	// example: SLAVE_ID=1 → test_slave1
	// set SLAVE_ID=1
	slaveID := os.Getenv("SLAVE_ID")

	if slaveID == "" {
		slaveID = "1" // fallback
	}

	dbName := fmt.Sprintf("test_slave%s", slaveID)

	dsn := fmt.Sprintf(
		"root:123456789@tcp(127.0.0.1:3306)/%s",
		dbName,
	)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	// 🔥 AUTO CREATE TABLE
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT  PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		)
	`)

	if err != nil {
		panic(err)
	}

	fmt.Println("✅ Slave connected to DB:", dbName)
}

// ---------------- AUTH ----------------

func VerifyHash(hash string) bool {
	expected := fmt.Sprintf("%x", sha256.Sum256([]byte(SECRET)))
	return hash == expected
}

// ---------------- HEALTH ----------------

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// ---------------- INSERT ----------------

func InsertHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
	http.Error(w, "POST only", 405)
	return
}
	var data map[string]string



	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

		if data["id"] == "" {
	http.Error(w, "missing id", 400)
	return
}

	if !VerifyHash(data["hash"]) {
		http.Error(w, "unauthorized", 403)
		return
	}

	_, err := db.Exec(
	"INSERT INTO users(id, name) VALUES(?, ?)",
	data["id"],
	data["value"],
)

	if err != nil {
		http.Error(w, "insert failed", 500)
		return
	}

	fmt.Println("📥 Insert:", data["value"])

	w.Write([]byte("inserted"))
}

// ---------------- SELECT ----------------

func SelectHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
	http.Error(w, "GET only", 405)
	return
}

	rows, err := db.Query("SELECT id, name FROM users")
	if err != nil {
		http.Error(w, "db error", 500)
		return
	}
	defer rows.Close()

	result := ""

	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		result += fmt.Sprintf("[%d:%s] ", id, name)
	}

	w.Write([]byte(result))
}

// ---------------- UPDATE ----------------

func UpdateHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
	http.Error(w, "POST only", 405)
	return
}
	var data map[string]string
	

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

		if data["id"] == "" {
	http.Error(w, "missing id", 400)
	return
}

	// 🔥 SECURITY FIX
	if !VerifyHash(data["hash"]) {
		http.Error(w, "unauthorized", 403)
		return
	}

	_, err := db.Exec(
		"UPDATE users SET name=? WHERE id=?",
		data["value"],
		data["id"],
	)

	if err != nil {
		http.Error(w, "update failed", 500)
		return
	}

	fmt.Println("✏️ Update:", data["id"])

	w.Write([]byte("updated"))
}

// ---------------- DELETE ----------------

func DeleteHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
	http.Error(w, "POST only", 405)
	return
}
	var data map[string]string



	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

		if data["id"] == "" {
	http.Error(w, "missing id", 400)
	return
}

	// 🔥 SECURITY FIX
	if !VerifyHash(data["hash"]) {
		http.Error(w, "unauthorized", 403)
		return
	}

	_, err := db.Exec(
		"DELETE FROM users WHERE id=?",
		data["id"],
	)

	if err != nil {
		http.Error(w, "delete failed", 500)
		return
	}

	fmt.Println("🗑 Delete:", data["id"])

	w.Write([]byte("deleted"))
}