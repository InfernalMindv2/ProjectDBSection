package SServices

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Insert_Handler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)

	var data map[string]interface{}
	json.Unmarshal(body, &data)

	table := data["table"].(string)
	name := data["name"].(string)
	age := int(data["age"].(float64))

	query := fmt.Sprintf("INSERT INTO %s (name, age) VALUES (?, ?)", table)

	_, err := DB.Exec(query, name, age)
	if err != nil {
		fmt.Println("Insert error on slave")
		return
	}

	fmt.Println("Replicated Insert Done")

	w.Write([]byte("OK"))
}