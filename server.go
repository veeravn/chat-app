package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

var session *gocql.Session
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
	Username string `json:"username"`
	Password string `json:"password"`
}

type Message struct {
	ID        gocql.UUID `json:"id"`
	Sender    string     `json:"sender"`
	Recipient string     `json:"recipient"`
	Content   string     `json:"content"`
	Timestamp time.Time  `json:"timestamp"`
	Read      bool       `json:"read"`
}

func initDB() {
	cluster := gocql.NewCluster(os.Getenv("CASSANDRA_HOST"))
	cluster.Keyspace = os.Getenv("CASSANDRA_KEYSPACE")
	cluster.Consistency = gocql.Quorum

	var err error
	session, err = cluster.CreateSession()
	if err != nil {
		fmt.Println("Failed to connect to Cassandra:", err)
		os.Exit(1)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
            username TEXT PRIMARY KEY,
            password TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS messages (
            id UUID,
            sender TEXT,
            recipient TEXT,
            content TEXT,
            timestamp TIMESTAMP,
            read BOOLEAN,
			PRIMARY KEY ((recipient), id)
        )`,
	}

	for _, query := range queries {
		err = session.Query(query).Exec()
		if err != nil {
			fmt.Println("Failed to create table:", err)
		}
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var creds ChatUser
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	err = session.Query("INSERT INTO users (username, password) VALUES (?, ?)", creds.Username, string(hashedPassword)).Exec()
	if err != nil {
		http.Error(w, "Error registering user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	var creds ChatUser
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var storedPassword string
	err = session.Query("SELECT password FROM users WHERE username = ?", creds.Username).Scan(&storedPassword)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
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

	fmt.Printf("User connected: %s\n", user.Username)
	mu.Lock()
	clients[user.Username] = conn
	mu.Unlock()

	iter := session.Query("SELECT id, sender, recipient, content, timestamp, read FROM messages WHERE recipient = ? AND read = false ALLOW FILTERING", user.Username).Iter()
	var unreadMessages []Message
	var msg Message
	for iter.Scan(&msg.ID, &msg.Sender, &msg.Recipient, &msg.Content, &msg.Timestamp, &msg.Read) {
		unreadMessages = append(unreadMessages, msg)
	}
	iter.Close()

	for _, msg := range unreadMessages {
		conn.WriteJSON(msg)
		session.Query("UPDATE messages SET read = true WHERE recipient = ? and id = ?", user.Username, msg.ID).Exec()
	}

	for {
		var newMsg Message
		err := conn.ReadJSON(&newMsg)
		if err != nil {
			mu.Lock()
			delete(clients, user.Username)
			mu.Unlock()
			break
		}

		newMsg.ID = gocql.TimeUUID()
		newMsg.Timestamp = time.Now()
		newMsg.Read = false

		err = session.Query("INSERT INTO messages (id, sender, recipient, content, timestamp, read) VALUES (?, ?, ?, ?, ?, ?)",
			newMsg.ID, newMsg.Sender, newMsg.Recipient, newMsg.Content, newMsg.Timestamp, newMsg.Read).Exec()
		if err != nil {
			fmt.Println("Error inserting message into database:", err)
		}

		mu.Lock()
		recipientConn, exists := clients[newMsg.Recipient]
		if exists {
			recipientConn.WriteJSON(newMsg)
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
