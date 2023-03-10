package main

import (
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
)

const addr = ":8000"

type LoginResponse struct {
	Message string `json:"message"`
}

type FunctionResponse struct {
	Function string `json:"function"`
	Input    string `json:"input"`
	Output   string `json:"output"`
}

func main() {

	// Open a connection to the database
	// connection string should be retrieved from AWS Secrets Manager or Azure Key Vault or any other suitable service
	db, err := sql.Open("mysql", "secret:jOdznoyH6swQB9sTGdLUeeSrtejWkcw@tcp(sre-bootcamp.czdpg2eovfhn.us-west-1.rds.amazonaws.com:3306)/bootcamp_tht")
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}
	defer db.Close()

	// Check that the database connection is valid
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// Define the HTTP handlers
	http.HandleFunc("/", handleRoot())
	http.HandleFunc("/_health", handleHealth())
	http.HandleFunc("/login", handleLogin(db))
	http.HandleFunc("/mask-to-cidr", handleMaskToCidr())
	http.HandleFunc("/cidr-to-mask", handleCidrToMask())

	// Start the HTTP server
	srv := &http.Server{
		Addr: addr,
	}
	log.Printf("starting server on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start server: %v", err)
	}
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

func handleLogin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the JSON request body
		var req struct {
			User     string `json:"username"`
			Password string `json:"password"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Verify the user's credentials against the database
		var hashedPassword, salt, role string
		var queryErr error
		err = db.QueryRow("SELECT password, salt, role FROM users WHERE username = ?", req.User).Scan(&hashedPassword, &salt, &role)
		if err != nil {
			queryErr = err
		}

		if queryErr == sql.ErrNoRows {
			log.Println("User not found")
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		} else if sha512Hash(req.Password+salt) == hashedPassword {
			// the username and password combination is valid
		} else {
			log.Println("Invalid combination of username or password")
			http.Error(w, "Invalid combination of username or password", http.StatusUnauthorized)
			return
		}

		// Generate a JWT token for the user
		token, err := generateToken(role)
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		// Send the token in the response body
		resp, err := json.Marshal(token)
		if err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	}
}

func handleMaskToCidr() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			w.WriteHeader(http.StatusUnauthorized)
			//http.Redirect(w, r, "/login", http.StatusSeeOther)
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
			//http.Redirect(w, r, "/login", http.StatusSeeOther)
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
	// Extract the JWT token from the Authorization header
	tokenString := extractToken(r)
	if tokenString == "" {
		return false
	}

	// Parse and verify the JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check that the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			log.Println(err)
			return nil, err
		}
		// secret should be retrieved from AWS Secrets Manager or Azure Key Vault or any other suitable service
		secret := "my2w7wjd7yXF64FIADfJxNs1oupTGAuW"
		// Return the secret key used to sign the token
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return false
	}

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

func generateToken(role string) (map[string]string, error) {
	// secret should be retrieved from AWS Secrets Manager or Azure Key Vault or any other suitable service
	secret := "my2w7wjd7yXF64FIADfJxNs1oupTGAuW"
	// Create a new JWT token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set the claims for the token
	claims := token.Claims.(jwt.MapClaims)
	claims["role"] = role
	//claims["exp"] = jwt.TimeFunc().Add(time.Hour * 24).Unix() // Token expires in 24 hours

	// Sign the token with the secret key
	secretKey := []byte(secret)
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return nil, err
	}

	// Return a map with the role and token fields
	return map[string]string{"role": role, "token": signedToken}, nil
}

func sha512Hash(str string) string {
	// Convert the string to bytes
	data := []byte(str)

	// Create a new SHA-512 hash object
	hash := sha512.New()

	// Write the data to the hash object
	hash.Write(data)

	// Get the raw hashed bytes
	hashed := hash.Sum(nil)

	// Convert the hashed bytes to a hex string
	return hex.EncodeToString(hashed)
}

func TestConnection(db *sql.DB) {

	// Test the connection by fetching data from a table
	rows, err := db.Query("SELECT username, role, salt, password FROM users")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	// Iterate over the rows and print the data
	for rows.Next() {
		var username, role, salt, pass string
		err = rows.Scan(&username, &role, &salt, &pass)
		if err != nil {
			panic(err.Error())
		}
		log.Printf("username: %s role: %s salt: %s pass: %s\n", username, role, salt, pass)
	}

	log.Println("Query successful!")
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	token := strings.Replace(authHeader, "Bearer ", "", 1)
	return token
}
