package dto

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound       = errors.New("user not found in chat service")
	ErrUsernameTaken      = errors.New("username is already used by another user")
	ErrInvalidUserProfile = errors.New("name and username cannot be blank")
)

type SyncUserRequest struct {
	Username  string  `json:"username" binding:"required,min=1,max=100"`
	Name      string  `json:"name" binding:"required,min=1,max=200"`
	AvatarURL *string `json:"avatar_url" binding:"omitempty,max=2048"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}
