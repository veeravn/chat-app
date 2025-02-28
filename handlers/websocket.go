package handlers

import (
	"chat-app/database"
	"chat-app/models"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

var clients = make(map[string]*websocket.Conn)
var mu sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handle WebSocket connections
func HandleConnections(w http.ResponseWriter, r *http.Request) {
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

		database.StoreUserConnection(user.Username)
		listenForUserMessages(conn, user.Username)
		return
	}

	// If it's not a user connection, process as a forwarded message
	log.Println("📩 Processing forwarded message...")
	processForwardedMessage(rawMessage)
}

// Listen for User Messages
func listenForUserMessages(conn *websocket.Conn, username string) {
	for {
		var msg models.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("❌ User %s disconnected: %v", username, err)
			mu.Lock()
			delete(clients, username)
			mu.Unlock()
			database.StoreUserConnection(username)
			return
		}

		log.Printf("📩 Received message from %s to %s: %s", msg.Sender, msg.Recipient, msg.Content)

		// Store message in Cassandra
		StoreMessage(msg)

		recipientServer, err := database.GetUserConnection(msg.Recipient)
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
				MarkMessageAsReadInCassandra(msg.Recipient, msg.ID)
			}
		} else {
			forwardMessageViaWebSocket(msg, recipientServer)
		}
	}
}

// Forward models.Message via WebSocket
func forwardMessageViaWebSocket(msg models.Message, recipientServer string) {
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

	if conn, exists := clients[serverAddr]; exists {
		return conn
	}

	ws, _, err := websocket.DefaultDialer.Dial(serverAddr+"/ws", nil)
	if err != nil {
		log.Printf("❌ Failed to connect to WebSocket server %s: %v", serverAddr, err)
		return nil
	}

	clients[serverAddr] = ws
	log.Printf("✅ Connected to WebSocket server %s", serverAddr)

	return ws
}

func processForwardedMessage(rawMessage json.RawMessage) {
	var msg models.Message
	err := json.Unmarshal(rawMessage, &msg)
	if err != nil {
		log.Printf("❌ Error decoding forwarded message JSON: %v", err)
		return
	}

	log.Printf("📩 Forwarded message received: %s from %s", msg.Content, msg.Sender)

	// Store forwarded message in Cassandra
	StoreMessage(msg)

	mu.Lock()
	recipientConn, exists := clients[msg.Recipient]
	mu.Unlock()

	if exists {
		recipientConn.WriteJSON(msg)
		log.Printf("✅ Delivered forwarded message to %s", msg.Recipient)
		// ✅ Mark as read since recipient is online
		MarkMessageAsReadInCassandra(msg.Recipient, msg.ID)
	} else {
		log.Printf("❌ Recipient %s not connected, storing message in DB.", msg.Recipient)
	}
}
