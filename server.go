package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients = make(map[string]*websocket.Conn) // 🔥 Map usernames to WebSocket connections
var mu sync.Mutex

type ChatUser struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex"`
	Password string
}

type Message struct {
	ID        uint      `gorm:"primaryKey"`
	Sender    string    `json:"sender"`
	Recipient string    `json:"recipient"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP"`
	Read      bool      `json:"read" gorm:"default:false"`
	Type      string    `json:"type"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func initDB() {
	var err error
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		fmt.Println("DATABASE_URL is not set")
		os.Exit(1)
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		os.Exit(1)
	}

	fmt.Println("Database connected successfully")

	// AutoMigrate will create/update tables as needed
	db.AutoMigrate(&ChatUser{}, &Message{})
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil || creds.Username == "" || creds.Password == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user := ChatUser{Username: creds.Username, Password: creds.Password}
	result := db.Create(&user)
	if result.Error != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	fmt.Printf("New user registered: %s\n", creds.Username)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil || creds.Username == "" || creds.Password == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var user ChatUser
	result := db.Where("username = ? AND password = ?", creds.Username, creds.Password).First(&user)
	if result.Error != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	var user struct {
		Username string `json:"username"`
	}
	err = conn.ReadJSON(&user)
	if err != nil {
		fmt.Println("Error reading username JSON:", err)
		return
	}

	mu.Lock()
	clients[user.Username] = conn // 🔥 Store the WebSocket connection by username
	mu.Unlock()

	fmt.Printf("%s connected\n", user.Username)

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Printf("User %s disconnected\n", user.Username)
			mu.Lock()
			delete(clients, user.Username)
			mu.Unlock()
			break
		}

		fmt.Printf("Message from %s to %s: %s\n", msg.Sender, msg.Recipient, msg.Content)

		msg.Timestamp = time.Now()
		msg.Read = false
		db.Create(&msg) // 🔥 Store message in the database

		mu.Lock()
		recipientConn, exists := clients[msg.Recipient] // 🔥 Retrieve recipient's connection
		if exists {
			fmt.Printf("Recipient %s is online", msg.Recipient)
			recipientConn.WriteJSON(msg) // 🔥 Send message to recipient if online
		} else {
			fmt.Printf("Recipient %s is offline", msg.Recipient)
		}
		mu.Unlock()
	}
}

func main() {
	initDB()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handleConnections)
	mux.HandleFunc("/api/auth", authHandler)
	mux.HandleFunc("/api/register", registerHandler)

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
	}).Handler(mux)

	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", handler)
}
