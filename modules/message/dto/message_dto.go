package dto

import (
	"errors"
	"time"

	channelDTO "github.com/Hieu3z03/chat-api-golang/modules/channel/dto"
	"github.com/google/uuid"
)

var ErrInvalidMessageContent = errors.New("message content must contain a non-empty string type")

type CreateMessageRequest struct {
	Content map[string]any `json:"content" binding:"required"`
}

type MessageResponse struct {
	ID        string                           `json:"id"`
	ChannelID uuid.UUID                        `json:"channel_id"`
	UserID    uuid.UUID                        `json:"user_id"`
	Sequence  int64                            `json:"sequence"`
	Content   map[string]any                   `json:"content"`
	CreatedAt time.Time                        `json:"created_at"`
	SeenBy    []channelDTO.ChannelUserResponse `json:"seen_by"`
}
