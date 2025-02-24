package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
)

var serverList []string

func init() {
	serverEnv := os.Getenv("WEBSOCKET_SERVERS")
	if serverEnv == "" {
		serverList = []string{
			"ws://websocket-server:8080",
			"ws://websocket-server:8080",
			"ws://websocket-server:8080",
		}
	} else {
		serverList = strings.Split(serverEnv, ",")
	}
}

var currentIndex uint64

func getNextServer() string {
	index := atomic.AddUint64(&currentIndex, 1)
	return serverList[index%uint64(len(serverList))]
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	backendServer := getNextServer() + "/ws"
	fmt.Println("Forwarding WebSocket request to:", backendServer)

	u, err := url.Parse(backendServer)
	if err != nil {
		http.Error(w, "Invalid WebSocket backend URL", http.StatusInternalServerError)
		return
	}

	proxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = u
			req.Host = u.Host
			req.Header.Set("Connection", "Upgrade")
			req.Header.Set("Upgrade", "websocket")
		},
		ModifyResponse: func(res *http.Response) error {
			res.Header.Set("Connection", "Upgrade")
			res.Header.Set("Upgrade", "websocket")
			return nil
		},
	}

	proxy.ServeHTTP(w, r)
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
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handleWebSocket)
	mux.HandleFunc("/api/", handleAPIRequests)

	fmt.Println("Load Balancer running on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
