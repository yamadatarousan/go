package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Todo struct {
	ID   int    `json:"id"`
	Task string `json:"task"`
	Done bool   `json:"done"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	mux.HandleFunc("GET /todos", func(w http.ResponseWriter, r *http.Request) {
		todos := []Todo{
			{ID: 1, Task: "Goの学習", Done: false},
			{ID: 2, Task: "標準ライブラリでAPI作成", Done: true},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todos)
	})

	port := ":8080"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
