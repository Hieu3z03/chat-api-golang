package dto

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidMessageContent = errors.New("message content must contain a non-empty string type")

type CreateMessageRequest struct {
	Content map[string]any `json:"content" binding:"required"`
}

type MessageResponse struct {
	ID        string         `json:"id"`
	ChannelID uuid.UUID      `json:"channel_id"`
	UserID    uuid.UUID      `json:"user_id"`
	Content   map[string]any `json:"content"`
	CreatedAt time.Time      `json:"created_at"`
}
