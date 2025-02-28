package database

import (
	"chat-app/config"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var ctx = context.Background()

// Initialize Redis connection
func InitRedis() {
	redisAddr := config.GetRedisHost()
	RedisClient = redis.NewClient(&redis.Options{Addr: redisAddr})

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")
}

// Get WebSocket server for a user
func GetUserConnection(username string) (string, error) {
	return RedisClient.Get(ctx, "user:"+username).Result()
}

// Store WebSocket server for a user
func StoreUserConnection(username string) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("❌ Error retrieving hostname: %v", err)
		return
	}

	serverAddress := fmt.Sprintf("ws://%s:8080", hostname) // Use dynamic hostname
	RedisClient.Set(ctx, "user:"+username, serverAddress, 0)
}

func RemoveUserConnection(username string) {
	log.Printf("🗑 Removing user connection for: %s", username)

	err := RedisClient.Del(ctx, "user:"+username).Err()
	if err != nil {
		log.Printf("❌ Failed to remove user connection: %v", err)
	} else {
		log.Printf("✅ User connection removed from Redis: %s", username)
	}
}
