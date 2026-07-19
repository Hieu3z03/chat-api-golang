package dto

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrChannelNotFound    = errors.New("channel not found")
	ErrNotChannelMember   = errors.New("user is not a member of this channel")
	ErrMembersNotFound    = errors.New("one or more users are not synchronized in chat service")
	ErrInvalidChannelName = errors.New("channel name cannot be empty")
)

type CreateChannelRequest struct {
	Name      string      `json:"name" binding:"required,max=120"`
	MemberIDs []uuid.UUID `json:"member_ids"`
}

type ChannelUserResponse struct {
	ID        uuid.UUID  `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Username  string     `json:"username"`
	AvatarID  *uuid.UUID `json:"avatar_id,omitempty"`
}

type ChannelResponse struct {
	ID        uuid.UUID             `json:"id"`
	Name      string                `json:"name"`
	CreatedBy uuid.UUID             `json:"created_by"`
	Members   []ChannelUserResponse `json:"members"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}
