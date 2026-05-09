package MServices

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func Replicate_Insert(urls []string, table string, name string, age int) {

	flow := make(chan string, len(urls))

	for _, url := range urls {

		go func(u string) {

			data := map[string]interface{}{
				"table": table,
				"name":  name,
				"age":   age,
			}

			json_data, _ := json.Marshal(data)

			resp, err := http.Post(u+"/insert", "application/json", bytes.NewBuffer(json_data))

			if err != nil {
				flow <- "Failed"
				return
			}

			defer resp.Body.Close()

			flow <- "Success"

		}(url)
	}

	for i := 0; i < len(urls); i++ {
		fmt.Println(<-flow)
	}
}