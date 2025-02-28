package main

import (
	"chat-app/database"
	"chat-app/handlers"
	"fmt"
	"net/http"

	"github.com/rs/cors"
)

func main() {
	database.InitCassandra()
	database.InitRedis()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handlers.HandleConnections)
	mux.HandleFunc("/api/auth", handlers.AuthHandler)
	mux.HandleFunc("/api/register", handlers.RegisterHandler)

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
	}).Handler(mux)

	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", handler)
}
