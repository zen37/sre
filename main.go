package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type LoginResponse struct {
	Message string `json:"message"`
}

type FunctionResponse struct {
	Function string `json:"function"`
	Input    string `json:"input"`
	Output   string `json:"output"`
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

func handleMaskToCidr() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		inputValue := r.URL.Query().Get("value")

		outputValue := maskToCidr(inputValue)

		response := FunctionResponse{
			Function: "maskToCidr",
			Input:    inputValue,
			Output:   outputValue,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func handleCidrToMask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		inputValue := r.URL.Query().Get("value")

		outputValue := cidrToMask(inputValue)

		response := FunctionResponse{
			Function: "cidrToMask",
			Input:    inputValue,
			Output:   outputValue,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func isAuthenticated(r *http.Request) bool {
	// TODO: Implement authentication logic
	return true
}

func cidrToMask(cidr string) string {
	prefixSize, _ := strconv.Atoi(cidr[strings.IndexByte(cidr, '/')+1:])
	mask := net.CIDRMask(prefixSize, 32)
	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
}

func maskToCidr(mask string) string {
	// Parse the IP address from the mask string.
	ip := net.ParseIP(mask)
	if ip == nil {
		return ""
	}

	// Count the number of bits that are set in the IP address.
	var bits int
	for _, b := range ip.To4() {
		for i := 0; i < 8; i++ {
			if b&(1<<uint(7-i)) > 0 {
				bits++
			}
		}
	}

	// Return the CIDR notation string.
	return strconv.Itoa(bits)
}

func main() {
	http.HandleFunc("/", handleRoot())
	http.HandleFunc("/_health", handleHealth())
	http.HandleFunc("/login", handleLogin())

	http.HandleFunc("/mask-to-cidr", handleMaskToCidr())
	http.HandleFunc("/cidr-to-mask", handleCidrToMask())

	fmt.Println("server started on port 8000")

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		panic(err)
	}
}
