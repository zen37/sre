package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type LoginResponse struct {
	Message string `json:"message"`
}

func handleRoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := LoginResponse{
			Message: "Welcome to the login page",
		}

		json.NewEncoder(w).Encode(response)
	}
}

func main() {
	http.HandleFunc("/", handleRoot())
	http.HandleFunc("/_health", handleHealth())
	http.HandleFunc("/login", handleLogin())

	fmt.Println("server started on port 8000")

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err)
	}
}
