package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var serverList = []string{
	"http://websocket-server:8080",
	"http://websocket-server:8080",
	"http://websocket-server:8080",
}

var currentIndex uint64

func getNextServer() string {
	index := atomic.AddUint64(&currentIndex, 1)
	return serverList[index%uint64(len(serverList))]
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	backendServer := getNextServer() + "/ws"
	fmt.Println("Forwarding WebSocket request to:", backendServer)

	// Parse the WebSocket backend server URL
	u, err := url.Parse(backendServer)
	if err != nil {
		http.Error(w, "Invalid WebSocket backend URL", http.StatusInternalServerError)
		return
	}

	// Upgrade the request to a WebSocket connection
	backendConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		http.Error(w, "Error connecting to WebSocket backend", http.StatusBadGateway)
		return
	}
	defer backendConn.Close()

	// Upgrade the client connection
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Error upgrading to WebSocket", http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Proxy data between the client and backend WebSocket
	go func() {
		for {
			messageType, msg, err := backendConn.ReadMessage()
			if err != nil {
				return
			}
			clientConn.WriteMessage(messageType, msg)
		}
	}()

	for {
		messageType, msg, err := clientConn.ReadMessage()
		if err != nil {
			return
		}
		backendConn.WriteMessage(messageType, msg)
	}
}

func handleAPIRequests(w http.ResponseWriter, r *http.Request) {
	backendServer := getNextServer() + r.URL.Path
	fmt.Println("Forwarding API request to:", backendServer)

	u, err := url.Parse(backendServer)
	if err != nil {
		http.Error(w, "Invalid backend URL", http.StatusInternalServerError)
		return
	}

	proxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = u
			req.Host = u.Host
		},
	}

	proxy.ServeHTTP(w, r)
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/auth", handleAPIRequests)
	fmt.Println("Load Balancer running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
