package dto

import (
	"errors"
	"time"

	channelDTO "github.com/Hieu3z03/chat-api-golang/modules/channel/dto"
	"github.com/Hieu3z03/chat-api-golang/modules/message/model"
	"github.com/google/uuid"
)

var ErrInvalidMessageContent = errors.New("message type and content must be valid")

type CreateMessageRequest struct {
	Type    model.MessageType `json:"type" binding:"required"`
	Content *string           `json:"content" binding:"required"`
}

type MessageResponse struct {
	ID        string                           `json:"id"`
	ChannelID uuid.UUID                        `json:"channel_id"`
	UserID    uuid.UUID                        `json:"user_id"`
	Type      model.MessageType                `json:"type"`
	Content   *string                          `json:"content,omitempty"`
	Sequence  int64                            `json:"sequence"`
	IsDeleted bool                             `json:"is_deleted"`
	CreatedAt time.Time                        `json:"created_at"`
	SeenBy    []channelDTO.ChannelUserResponse `json:"seen_by"`
}
