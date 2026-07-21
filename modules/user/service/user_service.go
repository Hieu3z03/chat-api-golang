package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Hieu3z03/chat-api-golang/database/entities"
	"github.com/Hieu3z03/chat-api-golang/modules/user/dto"
	"github.com/Hieu3z03/chat-api-golang/modules/user/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	Sync(ctx context.Context, userID uuid.UUID, request dto.SyncUserRequest) (dto.UserResponse, error)
	GetByID(ctx context.Context, userID uuid.UUID) (dto.UserResponse, error)
	Search(ctx context.Context, search string, limit int) ([]dto.UserResponse, error)
}

type userService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) UserService {
	return &userService{users: users}
}

func (service *userService) Sync(
	ctx context.Context,
	userID uuid.UUID,
	request dto.SyncUserRequest,
) (dto.UserResponse, error) {
	username := strings.TrimSpace(request.Username)
	name := strings.TrimSpace(request.Name)
	if name == "" || username == "" {
		return dto.UserResponse{}, dto.ErrInvalidUserProfile
	}

	var avatarURL *string
	if request.AvatarURL != nil {
		trimmed := strings.TrimSpace(*request.AvatarURL)
		if trimmed != "" {
			avatarURL = &trimmed
		}
	}

	existing, err := service.users.FindByUsername(ctx, username)
	if err == nil && existing.ID != userID {
		return dto.UserResponse{}, dto.ErrUsernameTaken
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.UserResponse{}, err
	}

	user, err := service.users.Upsert(ctx, entities.User{
		ID:        userID,
		Username:  username,
		Name:      name,
		AvatarURL: avatarURL,
	})
	if err != nil {
		return dto.UserResponse{}, err
	}

	return toUserResponse(user), nil
}

func (service *userService) GetByID(ctx context.Context, userID uuid.UUID) (dto.UserResponse, error) {
	user, err := service.users.FindByID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.UserResponse{}, dto.ErrUserNotFound
	}
	if err != nil {
		return dto.UserResponse{}, err
	}

	return toUserResponse(user), nil
}

func (service *userService) Search(ctx context.Context, search string, limit int) ([]dto.UserResponse, error) {
	users, err := service.users.Search(ctx, strings.TrimSpace(search), limit)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, toUserResponse(user))
	}

	return responses, nil
}

func toUserResponse(user entities.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
	}
}
