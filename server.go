package main

import (
    "fmt"
    "net/http"
    "github.com/gorilla/websocket"
    "sync"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var clients = make(map[*websocket.Conn]string)
var mu sync.Mutex

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

        fmt.Printf("Received: %s
", msg)

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
    fmt.Println("WebSocket server running on port 8080")
    http.ListenAndServe(":8080", nil)
}