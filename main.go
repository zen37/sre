package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type LoginResponse struct {
	Message string `json:"message"`
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/_health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Health OK"))
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := LoginResponse{
			Message: "Welcome to the login page",
		}

		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/notfound", NotFoundHandler)

	fmt.Println("server started on port 8000")

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err)
	}
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
}
