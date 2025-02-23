package main

import (
    "encoding/json"
    "net/http"
)

type Credentials struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func authHandler(w http.ResponseWriter, r *http.Request) {
    var creds Credentials
    err := json.NewDecoder(r.Body).Decode(&creds)
    if err != nil || creds.Username == "" || creds.Password == "" {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    if creds.Password == "password123" {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]bool{"success": true})
    } else {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
    }
}

func main() {
    http.HandleFunc("/api/auth", authHandler)
    http.ListenAndServe(":5000", nil)
}