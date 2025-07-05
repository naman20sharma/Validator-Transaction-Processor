package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		log.Printf("\033[34m[Server] Received batch:\n%s\033[0m", string(body))
		w.WriteHeader(http.StatusOK)
	})

	log.Println("\033[36m[Server] Listening on http://localhost:2002\033[0m")
	if err := http.ListenAndServe(":2002", nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
