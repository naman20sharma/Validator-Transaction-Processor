package main

import (
	"encoding/json"
	"log"
	"net"
)

func main() {
	// Sample transaction: alice loses 10, bob gets 2, carol gets 8 → Net change = 0
	tx := map[string]interface{}{
		"fee": map[string]interface{}{
			"payer":  "alice",
			"amount": 1,
		},
		"instructions": []map[string]interface{}{
			{
				"account": "alice",
				"change":  -10,
			},
			{
				"account": "bob",
				"change":  2,
			},
			{
				"account": "carol",
				"change":  8,
			},
		},
	}

	data, err := json.Marshal(tx)
	if err != nil {
		log.Fatalf("Failed to encode transaction: %v", err)
	}

	conn, err := net.Dial("udp", "127.0.0.1:2001")
	if err != nil {
		log.Fatalf("Failed to dial UDP: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

	log.Println("Sample transaction sent to 127.0.0.1:2001")
}
