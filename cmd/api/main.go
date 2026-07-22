package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Health check available at http://localhost:8080/health")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
