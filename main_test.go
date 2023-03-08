package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dgrijalva/jwt-go"
)

func TestRootEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRoot())

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestHealthEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/_health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleHealth())

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "OK"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
func TestHandleLogin(t *testing.T) {
	// Define a test table
	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
		expectedToken  string
	}{
		{"Valid credentials", "bob", "thisIsNotAPasswordBob", http.StatusOK, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoidmlld2VyIn0.l7pxJXYHlJdtI9RME2UesMzuVjqf-RtzQeLTHomo_Ic"},
		{"Invalid username", "invaliduser", "testpassword", http.StatusUnauthorized, ""},
		{"Invalid password", "testuser", "invalidpassword", http.StatusUnauthorized, ""},
	}

	// connection string should be retrieved from AWS Secrets Manager or Azure Key Vault or any other suitable service
	db, err := sql.Open("mysql", "secret:jOdznoyH6swQB9sTGdLUeeSrtejWkcw@tcp(sre-bootcamp.czdpg2eovfhn.us-west-1.rds.amazonaws.com:3306)/bootcamp_tht")
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a new request with a JSON request body
			reqBody := strings.NewReader(fmt.Sprintf(`{"username":"%s","password":"%s"}`, test.username, test.password))
			req, err := http.NewRequest("POST", "/login", reqBody)
			if err != nil {
				t.Fatal(err)
			}
			// Set the content type to JSON
			req.Header.Set("Content-Type", "application/json")

			// Create a new ResponseRecorder (which satisfies http.ResponseWriter) to capture the response
			rr := httptest.NewRecorder()

			// Call the handler function, passing in the mock request and response recorder
			handleLogin(db)(rr, req)

			// Check that the response status code is the expected status code
			if status := rr.Code; status != test.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, test.expectedStatus)
			}

			// Check that the response body contains a non-empty token (if the expected status code is OK)
			if test.expectedStatus == http.StatusOK {
				body := rr.Body.String()
				var response map[string]string
				err := json.Unmarshal([]byte(body), &response)
				if err != nil {
					t.Errorf("failed to unmarshal response body: %v", err)
				}
				actualToken := response["token"]
				if actualToken != test.expectedToken {
					t.Errorf("HandleLogin did not return the expected token for test case %q: got %q, want %q", test.name, actualToken, test.expectedToken)
				}
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {

	// secret should be retrieved from AWS Secrets Manager or Azure Key Vault or any other suitable service
	secret := "my2w7wjd7yXF64FIADfJxNs1oupTGAuW"

	role := "admin"

	tokenMap, err := generateToken(role)
	if err != nil {
		t.Fatalf("Error generating token: %v", err)
	}

	tokenString, ok := tokenMap["token"]
	if !ok {
		t.Fatalf("Token map doesn't contain token field")
	}

	// Parse the token to get the claims
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("Error parsing token: %v", err)
	}

	// Verify the claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("Token claims are not of type MapClaims")
	}

	if role, ok := claims["role"].(string); !ok || role != "admin" {
		t.Errorf("Token has unexpected role: %v", role)
	}

}

func TestHandleMaskToCidr(t *testing.T) {
	req, err := http.NewRequest("GET", "/mask-to-cidr?value=255.255.0.0", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := handleMaskToCidr()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"function":"maskToCidr","input":"255.255.0.0","output":"16"}`
	actual := strings.TrimSpace(rr.Body.String())

	if actual != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestHandleCidrToMask(t *testing.T) {
	req, err := http.NewRequest("GET", "/cidr-to-mask?value=24", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := handleCidrToMask()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := strings.TrimSpace(`{"function":"cidrToMask","input":"24","output":"255.255.255.0"}`)
	actual := strings.TrimSpace(rr.Body.String())

	if actual != expected {
		t.Errorf("handler returned unexpected body:\nactual: %q\nexpected: %q", actual, expected)
	}
}
