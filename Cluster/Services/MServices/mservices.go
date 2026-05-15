package MServices

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Slave struct {
	IP   string `json:"IP"`
	Port string `json:"Port"`
	ID   int    `json:"ID"`
}

// ---------------- LOAD CONFIG ----------------

func LoadConfig() []Slave {

	file, err := os.ReadFile("C:/Project DB/Cluster/Config/config.json")
	if err != nil {
		fmt.Println("❌ Config read error:", err)
		return nil
	}

	var slaves []Slave

	err = json.Unmarshal(file, &slaves)
	if err != nil {
		fmt.Println("❌ JSON parse error:", err)
		return nil
	}

	fmt.Println("✅ Loaded slaves:", slaves)

	return slaves
}

// ---------------- HEALTH ----------------

func IsAlive(s Slave) bool {

	client := http.Client{Timeout: 2 * time.Second}

	url := fmt.Sprintf("http://%s:%s/health", s.IP, s.Port)

	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("❌ Slave", s.ID, "down")
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ---------------- HASH ----------------

func GenerateHash(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return fmt.Sprintf("%x", sum)
}

// ---------------- SHARDING ----------------

func SelectSlave(key int, slaves []Slave) Slave {

	if len(slaves) == 0 {
		fmt.Println("❌ No slaves available for sharding")
		return Slave{}
	}

	index := key % len(slaves)

	fmt.Println("📌 Selected slave:", slaves[index].ID)

	return slaves[index]
}

// ---------------- FAILOVER ----------------

func Failover(slaves []Slave) Slave {

	fmt.Println("⚠️ Running failover...")

	for _, s := range slaves {
		if IsAlive(s) {
			fmt.Println("✅ Failover selected slave:", s.ID)
			return s
		}
	}

	fmt.Println("❌ All slaves are down")
	return Slave{}
}

// ---------------- SEND REQUEST ----------------

func SendToSlave(s Slave, endpoint string, data map[string]string) error {

	body, _ := json.Marshal(data)

	url := fmt.Sprintf("http://%s:%s/%s", s.IP, s.Port, endpoint)

	client := http.Client{Timeout: 3 * time.Second}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("❌ Request creation failed:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ HTTP error to slave", s.ID, ":", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		respBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf(
			"slave returned %d: %s",
			resp.StatusCode,
			string(respBody),
		)
	}


	// read response (VERY IMPORTANT for debugging)
	respBody, _ := io.ReadAll(resp.Body)

	fmt.Println("📩 Slave", s.ID, "response:", string(respBody))

	return nil
}

func SendWithFailover(
	slaves []Slave,
	endpoint string,
	data map[string]string,
) (Slave, error) {

	for _, s := range slaves {

		if !IsAlive(s) {
			fmt.Println("❌ Slave", s.ID, "down")
			continue
		}

		err := SendToSlave(s, endpoint, data)

		if err == nil {
			fmt.Println("✅ Request handled by slave", s.ID)
			return s, nil
		}
	}

	return Slave{}, fmt.Errorf("all slaves failed")
}