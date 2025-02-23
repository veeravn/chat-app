package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

var serverList = []string{
	"ws://localhost:8081/ws",
	"ws://localhost:8082/ws",
	"ws://localhost:8083/ws",
}

var currentIndex uint64

func getNextServer() string {
	index := atomic.AddUint64(&currentIndex, 1)
	return serverList[index%uint64(len(serverList))]
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	backendServer := getNextServer()
	fmt.Println("Forwarding request to:", backendServer)

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
	fmt.Println("WebSocket Load Balancer running on ws://localhost:8080/ws")
	http.ListenAndServe(":8080", nil)
}
