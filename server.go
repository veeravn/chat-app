package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]string)
var mu sync.Mutex

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

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	clients[conn] = ""
	mu.Unlock()

	fmt.Println("New client connected")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected")
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			break
		}

		fmt.Printf("Received: %s\n", msg)

		mu.Lock()
		for client := range clients {
			if client != conn {
				client.WriteMessage(websocket.TextMessage, msg)
			}
		}
		mu.Unlock()
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/api/auth", authHandler)
	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)
}
