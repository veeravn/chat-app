package database

import (
	"fmt"
	"log"
	"os"

	"github.com/gocql/gocql"
)

var Session *gocql.Session

// Initialize Cassandra connection
func InitCassandra() {
	cluster := gocql.NewCluster(os.Getenv("CASSANDRA_HOST"))
	cluster.Keyspace = os.Getenv("CASSANDRA_KEYSPACE")
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		fmt.Println("Failed to connect to Cassandra:", err)
		os.Exit(1)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
            username TEXT PRIMARY KEY,
            password TEXT
        )`,
		`CREATE TABLE IF NOT EXISTS messages (
            id UUID,
            sender TEXT,
            recipient TEXT,
            content TEXT,
            timestamp TIMESTAMP,
            read BOOLEAN,
			PRIMARY KEY ((recipient), id)
        )`,
	}

	for _, query := range queries {
		err = session.Query(query).Exec()
		if err != nil {
			fmt.Println("Failed to create table:", err)
		}
	}
	Session = session
}

func markMessageAsReadInCassandra(recipient string, msgID gocql.UUID) {
	query := "UPDATE messages SET read = true WHERE recipient = ? and id = ?"
	err := Session.Query(query, recipient, msgID).Exec()
	if err != nil {
		log.Printf("❌ Failed to mark message as read in Cassandra: %v", err)
	} else {
		log.Printf("✅ Message marked as read in Cassandra (ID: %s)", msgID)
	}
}
