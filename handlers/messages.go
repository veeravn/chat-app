package handlers

import (
	"chat-app/database"
	"chat-app/models"
	"log"

	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
)

// Save message to Cassandra
func StoreMessage(msg models.Message) {
	msg.ID = gocql.TimeUUID()
	query := "INSERT INTO messages (id, sender, recipient, content, timestamp, read) VALUES (?, ?, ?, ?, ?, ?)"
	err := database.Session.Query(query, msg.ID, msg.Sender, msg.Recipient, msg.Content, msg.Timestamp, msg.Read).Exec()
	if err != nil {
		log.Printf("❌ Failed to save message: %v", err)
	}
}

func MarkMessageAsReadInCassandra(recipient string, msgID gocql.UUID) {
	query := "UPDATE messages SET read = true WHERE recipient = ? and id = ?"
	err := database.Session.Query(query, recipient, msgID).Exec()
	if err != nil {
		log.Printf("❌ Failed to mark message as read in Cassandra: %v", err)
	} else {
		log.Printf("✅ Message marked as read in Cassandra (ID: %s)", msgID)
	}
}

// Handle incoming messages
func HandleIncomingMessage(conn *websocket.Conn, msg models.Message) {
	log.Printf("📩 Received message from %s to %s: %s", msg.Sender, msg.Recipient, msg.Content)

	// Store message
	StoreMessage(msg)

	// Forward message if recipient is online
	recipientConn, exists := clients[msg.Recipient]
	if exists {
		recipientConn.WriteJSON(msg)
	}
}
