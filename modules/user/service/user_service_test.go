package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Hieu3z03/chat-api-golang/database/entities"
	"github.com/Hieu3z03/chat-api-golang/modules/user/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fakeUserRepository struct {
	upserted entities.User
}

func (repository *fakeUserRepository) Upsert(_ context.Context, user entities.User) (entities.User, error) {
	repository.upserted = user
	return user, nil
}

func (repository *fakeUserRepository) FindByID(context.Context, uuid.UUID) (entities.User, error) {
	return entities.User{}, gorm.ErrRecordNotFound
}

func (repository *fakeUserRepository) FindByUsername(context.Context, string) (entities.User, error) {
	return entities.User{}, gorm.ErrRecordNotFound
}

func (repository *fakeUserRepository) FindByIDs(context.Context, []uuid.UUID) ([]entities.User, error) {
	return nil, nil
}

func (repository *fakeUserRepository) Search(context.Context, string, int) ([]entities.User, error) {
	return nil, nil
}

func TestSyncStoresMinimalChatProfile(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewUserService(repository)
	userID := uuid.New()

	response, err := service.Sync(context.Background(), userID, dto.SyncUserRequest{
		Name:     "  Hieu Nguyen  ",
		Username: " hieu ",
	})
	if err != nil {
		t.Fatalf("sync user: %v", err)
	}
	if repository.upserted.Name != "Hieu Nguyen" {
		t.Fatalf("name was not normalized before persistence")
	}
	if response.ID != userID || response.Username != "hieu" {
		t.Fatalf("unexpected sync response: %+v", response)
	}
}

func TestSyncRejectsBlankName(t *testing.T) {
	service := NewUserService(&fakeUserRepository{})

	_, err := service.Sync(context.Background(), uuid.New(), dto.SyncUserRequest{
		Name:     "   ",
		Username: "hieu",
	})
	if !errors.Is(err, dto.ErrInvalidUserProfile) {
		t.Fatalf("expected ErrInvalidUserProfile, got %v", err)
	}
}
