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
		return true // Allow connections from any origin
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients = make(map[*websocket.Conn]string) // Mapping WebSocket connections to usernames
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

	var username string
	err = json.NewDecoder(r.Body).Decode(&username)
	if err != nil || username == "" {
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	mu.Lock()
	clients[conn] = username
	mu.Unlock()

	fmt.Printf("%s connected\n", username)

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Printf("User %s disconnected\n", username)
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			break
		}

		if msg.Type == "typing" {
			broadcastTypingIndicator(msg.Sender, msg.Recipient)
			continue
		}

		msg.Timestamp = time.Now()
		msg.Read = false

		fmt.Printf("Message from %s to %s: %s\n", msg.Sender, msg.Recipient, msg.Content)

		db.Create(&msg)
		sendPushNotification(msg.Recipient, msg.Content)

		mu.Lock()
		for client, user := range clients {
			if user == msg.Recipient {
				client.WriteJSON(msg)
				db.Model(&Message{}).Where("sender = ? AND recipient = ?", msg.Sender, msg.Recipient).Update("read", true)
			}
		}
		mu.Unlock()
	}
}

func broadcastTypingIndicator(sender string, recipient string) {
	msg := Message{
		Sender:    sender,
		Recipient: recipient,
		Type:      "typing",
	}

	mu.Lock()
	for client, user := range clients {
		if user == recipient {
			client.WriteJSON(msg)
		}
	}
	mu.Unlock()
}

func sendPushNotification(recipient string, content string) {
	fmt.Printf("Push notification sent to %s: %s\n", recipient, content)
}

func main() {
	initDB()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handleConnections)
	mux.HandleFunc("/api/auth", authHandler)
	mux.HandleFunc("/api/register", registerHandler)

	handler := cors.Default().Handler(mux)

	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", handler)
}
