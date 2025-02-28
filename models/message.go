package models

import (
	"time"

	"github.com/gocql/gocql"
)

type Message struct {
	ID        gocql.UUID `json:"id"`
	Sender    string     `json:"sender"`
	Recipient string     `json:"recipient"`
	Content   string     `json:"content"`
	Timestamp time.Time  `json:"timestamp"`
	Read      bool       `json:"read"`
}
