package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Message struct {
	ID        bson.ObjectID  `bson:"_id"`
	ChannelID string         `bson:"channel_id"`
	UserID    string         `bson:"user_id"`
	Sequence  int64          `bson:"sequence"`
	Content   map[string]any `bson:"content"`
	CreatedAt time.Time      `bson:"created_at"`
}
