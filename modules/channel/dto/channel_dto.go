package dto

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrChannelNotFound      = errors.New("channel not found")
	ErrNotChannelMember     = errors.New("user is not a member of this channel")
	ErrMembersNotFound      = errors.New("one or more users are not synchronized in chat service")
	ErrInvalidChannelName   = errors.New("channel name cannot be empty")
	ErrChannelAlreadyExists = errors.New("channel already exists for these users")
)

type CreateChannelRequest struct {
	Name      string      `json:"name" binding:"required,max=120"`
	MemberIDs []uuid.UUID `json:"member_ids"`
}

type ChannelUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

type ChannelMemberReadState struct {
	User             ChannelUserResponse
	LastReadSequence int64
	LastReadAt       *time.Time
}

type ChannelResponse struct {
	ID        uuid.UUID             `json:"id"`
	Name      string                `json:"name"`
	CreatedBy uuid.UUID             `json:"created_by"`
	Members   []ChannelUserResponse `json:"members"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}
