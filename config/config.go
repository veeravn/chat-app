package config

import (
	"os"
)

// Load environment variables
func GetRedisHost() string {
	return os.Getenv("REDIS_HOST")
}

func GetCassandraHost() string {
	return os.Getenv("CASSANDRA_HOST")
}

func GetCassandraKeyspace() string {
	return os.Getenv("CASSANDRA_KEYSPACE")
}

func GetWebsocketServer() string {
	return os.Getenv("WEBSOCKET_SERVER")
}
