package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MessageType string

const (
	MessageTypeText    MessageType = "text"
	MessageTypeImage   MessageType = "image"
	MessageTypeFile    MessageType = "file"
	MessageTypeSticker MessageType = "sticker"
	MessageTypeSystem  MessageType = "system"
)

type Message struct {
	ID        bson.ObjectID `bson:"_id"`
	ChannelID string        `bson:"channel_id"`
	UserID    string        `bson:"user_id"`
	Type      MessageType   `bson:"type"`
	Content   *string       `bson:"content,omitempty"`
	Sequence  int64         `bson:"sequence"`
	IsDeleted bool          `bson:"is_deleted"`
	CreatedAt time.Time     `bson:"created_at"`
}
