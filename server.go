package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

// Global Variables
var (
	redisClient       *redis.Client
	ctx               = context.Background()
	clients           = make(map[string]*websocket.Conn)
	serverConnections = make(map[string]*websocket.Conn)
	mu                sync.Mutex
	session           *gocql.Session
)

// WebSocket Upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Message Structure
type Message struct {
	ID        gocql.UUID `json:"id"`
	Sender    string     `json:"sender"`
	Recipient string     `json:"recipient"`
	Content   string     `json:"content"`
	Timestamp time.Time  `json:"timestamp"`
	Read      bool       `json:"read"`
}

type ChatUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Initialize Redis Connection
func initRedis() {
	redisAddr := os.Getenv("REDIS_HOST")
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")
}

// Store User Connection in Redis
func storeUserConnection(username string) {
	// Get the hostname assigned by Docker
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("❌ Error retrieving hostname: %v", err)
		return
	}

	serverAddress := fmt.Sprintf("ws://%s:8080", hostname) // Use dynamic hostname
	log.Printf("📝 Storing user connection: %s -> %s", username, serverAddress)

	err = redisClient.Set(ctx, "user:"+username, serverAddress, 0).Err()
	if err != nil {
		log.Printf("❌ Failed to store user connection: %v", err)
	} else {
		log.Printf("✅ Connection stored in Redis: %s -> %s", username, serverAddress)
	}
}

// Remove User Connection from Redis
func removeUserConnection(username string) {
	log.Printf("🗑 Removing user connection for: %s", username)

	err := redisClient.Del(ctx, "user:"+username).Err()
	if err != nil {
		log.Printf("❌ Failed to remove user connection: %v", err)
	} else {
		log.Printf("✅ User connection removed from Redis: %s", username)
	}
}

// Lookup User Connection in Redis
func getUserConnection(username string) (string, error) {
	log.Printf("🔍 Looking up WebSocket server for user: %s", username)

	server, err := redisClient.Get(ctx, "user:"+username).Result()
	if err == redis.Nil {
		log.Printf("❌ User %s is not connected to any WebSocket server", username)
		return "", fmt.Errorf("User %s not connected", username)
	} else if err != nil {
		log.Printf("❌ Redis error while retrieving user connection: %v", err)
		return "", err
	}

	log.Printf("✅ User %s is connected to %s", username, server)
	return server, nil
}

// WebSocket Connection Handler
func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("❌ Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	log.Println("🔄 New WebSocket connection received, waiting for data...")

	// Read first message and determine if it's a user login or a forwarded message
	var rawMessage json.RawMessage
	err = conn.ReadJSON(&rawMessage)
	if err != nil {
		log.Println("❌ Error reading WebSocket message:", err)
		return
	}

	// Try parsing as a user connection first
	var user struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(rawMessage, &user); err == nil && user.Username != "" {
		log.Printf("✅ User connected: %s", user.Username)

		mu.Lock()
		clients[user.Username] = conn
		mu.Unlock()

		storeUserConnection(user.Username)
		listenForUserMessages(conn, user.Username)
		return
	}

	// If it's not a user connection, process as a forwarded message
	log.Println("📩 Processing forwarded message...")
	processForwardedMessage(conn, rawMessage)
}

// Listen for User Messages
func listenForUserMessages(conn *websocket.Conn, username string) {
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("❌ User %s disconnected: %v", username, err)
			mu.Lock()
			delete(clients, username)
			mu.Unlock()
			removeUserConnection(username)
			return
		}

		log.Printf("📩 Received message from %s to %s: %s", msg.Sender, msg.Recipient, msg.Content)

		// Store message in Cassandra
		storeMessageInCassandra(msg)

		recipientServer, err := getUserConnection(msg.Recipient)
		if err != nil {
			log.Printf("❌ Recipient %s is offline.", msg.Recipient)
			continue
		}

		if recipientServer == os.Getenv("WEBSOCKET_SERVER") {
			mu.Lock()
			recipientConn, exists := clients[msg.Recipient]
			mu.Unlock()
			if exists {
				recipientConn.WriteJSON(msg)
				// ✅ Mark as read since recipient is online
				markMessageAsReadInCassandra(msg.Recipient, msg.ID)
			}
		} else {
			forwardMessageViaWebSocket(msg, recipientServer)
		}
	}
}

// Forward Message via WebSocket
func forwardMessageViaWebSocket(msg Message, recipientServer string) {
	log.Printf("🚀 Attempting to forward message from %s to %s via %s", msg.Sender, msg.Recipient, recipientServer)

	ws := connectToServer(recipientServer)
	if ws == nil {
		log.Printf("❌ Could not establish WebSocket connection to %s", recipientServer)
		return
	}

	jsonData, _ := json.Marshal(msg)
	log.Printf("📩 Sending forwarded message to %s: %s", recipientServer, string(jsonData))

	err := ws.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Printf("❌ WebSocket error while sending to %s: %v", recipientServer, err)
	} else {
		log.Printf("✅ Successfully forwarded message to %s", recipientServer)
	}
}

// Establish WebSocket Connection to Another Server
func connectToServer(serverAddr string) *websocket.Conn {
	mu.Lock()
	defer mu.Unlock()

	if conn, exists := serverConnections[serverAddr]; exists {
		return conn
	}

	ws, _, err := websocket.DefaultDialer.Dial(serverAddr+"/ws", nil)
	if err != nil {
		log.Printf("❌ Failed to connect to WebSocket server %s: %v", serverAddr, err)
		return nil
	}

	serverConnections[serverAddr] = ws
	log.Printf("✅ Connected to WebSocket server %s", serverAddr)

	return ws
}

// Process Forwarded WebSocket Messages
func processForwardedMessage(conn *websocket.Conn, rawMessage json.RawMessage) {
	var msg Message
	err := json.Unmarshal(rawMessage, &msg)
	if err != nil {
		log.Printf("❌ Error decoding forwarded message JSON: %v", err)
		return
	}

	log.Printf("📩 Forwarded message received: %s from %s", msg.Content, msg.Sender)

	// Store forwarded message in Cassandra
	storeMessageInCassandra(msg)

	mu.Lock()
	recipientConn, exists := clients[msg.Recipient]
	mu.Unlock()

	if exists {
		recipientConn.WriteJSON(msg)
		log.Printf("✅ Delivered forwarded message to %s", msg.Recipient)
		// ✅ Mark as read since recipient is online
		markMessageAsReadInCassandra(msg.Recipient, msg.ID)
	} else {
		log.Printf("❌ Recipient %s not connected, storing message in DB.", msg.Recipient)
	}
}

func markMessageAsReadInCassandra(recipient string, msgID gocql.UUID) {
	query := "UPDATE chat.messages SET read = true WHERE recipient = ? and id = ?"
	err := session.Query(query, recipient, msgID).Exec()
	if err != nil {
		log.Printf("❌ Failed to mark message as read in Cassandra: %v", err)
	} else {
		log.Printf("✅ Message marked as read in Cassandra (ID: %s)", msgID)
	}
}

func getServerAddress() string {
	hostname, _ := os.Hostname() // Get unique hostname (e.g., ws-server_1, ws-server_2)
	return "ws://" + hostname + ":8080"
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

func storeMessageInCassandra(msg Message) {
	msg.ID = gocql.TimeUUID() // Generate unique message ID

	query := "INSERT INTO chat.messages (id, sender, recipient, content, timestamp, read) VALUES (?, ?, ?, ?, ?, ?)"
	err := session.Query(query, msg.ID, msg.Sender, msg.Recipient, msg.Content, msg.Timestamp, msg.Read).Exec()

	if err != nil {
		log.Printf("❌ Failed to save message in Cassandra: %v", err)
	} else {
		log.Printf("✅ Message saved in Cassandra: %s -> %s", msg.Sender, msg.Recipient)
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

func main() {
	initDB()
	initRedis()

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
