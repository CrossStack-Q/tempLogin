package main

import (
        "encoding/json"
        "fmt"
        "log"
        "math/rand"
        "net/http"
        "sync"
        "time"
)

// Credentials represents incoming login data
type Credentials struct {
        Email    string `json:"email"`
        Password string `json:"password"`
}

// TokenResponse holds the response with token
type TokenResponse struct {
        Message string `json:"message"`
        Token   string `json:"token"`
        Expiry  string `json:"expiry"`
}

const (
        validEmail    = "Shubham.baheti@brevo.com"
        validPassword = "12345678"
)

var (
        tokenStore = make(map[string]time.Time)
        mu         sync.Mutex
)

// generateToken returns a pseudo-random token and sets it to expire after 4 hours
func generateToken() (string, time.Time) {
        const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
        b := make([]byte, 32)
        for i := range b {
                b[i] = letters[rand.Intn(len(letters))]
        }
        token := string(b)
        expiry := time.Now().Add(4 * time.Hour)

        mu.Lock()
        tokenStore[token] = expiry
        mu.Unlock()

        return token, expiry
}

func main() {
        rand.Seed(time.Now().UnixNano())
        http.HandleFunc("/login", corsMiddleware(loginHandler))

        fmt.Println("Server running on http://localhost:8080")
        log.Fatal(http.ListenAndServe(":8900", nil))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusOK)
                return
        }

        if r.Method != http.MethodPost {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
        }

        var creds Credentials
        err := json.NewDecoder(r.Body).Decode(&creds)
        if err != nil {
                http.Error(w, "Invalid JSON", http.StatusBadRequest)
                return
        }

        if creds.Email == validEmail && creds.Password == validPassword {
                token, expiry := generateToken()
                resp := TokenResponse{
                        Message: "Login successful",
                        Token:   token,
                        Expiry:  expiry.Format(time.RFC3339),
                }
                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(resp)
        } else {
                http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        }
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Access-Control-Allow-Origin", "*")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
                w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
                if r.Method == http.MethodOptions {
                        w.WriteHeader(http.StatusOK)
                        return
                }
                next.ServeHTTP(w, r)
        }
}
