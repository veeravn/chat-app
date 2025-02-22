package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	clients[conn] = true
	fmt.Println("New client connected")

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected")
			delete(clients, conn)
			break
		}
		fmt.Printf("Received: %s\n", msg)

		for client := range clients {
			if err := client.WriteMessage(messageType, msg); err != nil {
				fmt.Println("Error sending message:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	fmt.Println("WebSocket server running on ws://localhost:8080/ws")
	http.ListenAndServe(":8080", nil)
}

