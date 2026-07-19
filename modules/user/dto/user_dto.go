package dto

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound       = errors.New("user not found in chat service")
	ErrUsernameTaken      = errors.New("username is already used by another user")
	ErrInvalidUserProfile = errors.New("first_name, last_name and username cannot be blank")
)

type SyncUserRequest struct {
	FirstName string     `json:"first_name" binding:"required,min=1,max=100"`
	LastName  string     `json:"last_name" binding:"required,min=1,max=100"`
	Username  string     `json:"username" binding:"required,min=1,max=100"`
	AvatarID  *uuid.UUID `json:"avatar_id"`
}

type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Username  string     `json:"username"`
	AvatarID  *uuid.UUID `json:"avatar_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
