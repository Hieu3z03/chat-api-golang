package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Hieu3z03/chat-api-golang/database/entities"
	"github.com/Hieu3z03/chat-api-golang/modules/channel/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fakeUserRepository struct {
	users []entities.User
}

func (repository *fakeUserRepository) Upsert(context.Context, entities.User) (entities.User, error) {
	return entities.User{}, nil
}

func (repository *fakeUserRepository) FindByID(context.Context, uuid.UUID) (entities.User, error) {
	return entities.User{}, gorm.ErrRecordNotFound
}

func (repository *fakeUserRepository) FindByUsername(context.Context, string) (entities.User, error) {
	return entities.User{}, gorm.ErrRecordNotFound
}

func (repository *fakeUserRepository) FindByIDs(context.Context, []uuid.UUID) ([]entities.User, error) {
	return repository.users, nil
}

func (repository *fakeUserRepository) Search(context.Context, string, int) ([]entities.User, error) {
	return repository.users, nil
}

type fakeChannelRepository struct {
	memberIDs []uuid.UUID
}

func (repository *fakeChannelRepository) Create(
	_ context.Context,
	channel entities.Channel,
	memberIDs []uuid.UUID,
) (entities.Channel, error) {
	repository.memberIDs = memberIDs
	channel.ID = uuid.New()
	return channel, nil
}

func (repository *fakeChannelRepository) FindByID(context.Context, uuid.UUID) (entities.Channel, error) {
	return entities.Channel{}, gorm.ErrRecordNotFound
}

func (repository *fakeChannelRepository) ListByUser(context.Context, uuid.UUID) ([]entities.Channel, error) {
	return nil, nil
}

func (repository *fakeChannelRepository) IsMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}

func TestCreateAddsCreatorAndDeduplicatesMembers(t *testing.T) {
	creatorID := uuid.New()
	memberID := uuid.New()
	users := &fakeUserRepository{users: []entities.User{{ID: creatorID}, {ID: memberID}}}
	channels := &fakeChannelRepository{}
	service := NewChannelService(channels, users)

	_, err := service.Create(context.Background(), creatorID, dto.CreateChannelRequest{
		Name:      "Engineering",
		MemberIDs: []uuid.UUID{memberID, creatorID, memberID},
	})
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	if len(channels.memberIDs) != 2 {
		t.Fatalf("expected 2 unique members, got %d", len(channels.memberIDs))
	}
	if channels.memberIDs[0] != creatorID {
		t.Fatalf("creator must be the first channel member")
	}
}

func TestCreateRejectsUsersNotSynchronizedInChat(t *testing.T) {
	creatorID := uuid.New()
	users := &fakeUserRepository{users: []entities.User{{ID: creatorID}}}
	service := NewChannelService(&fakeChannelRepository{}, users)

	_, err := service.Create(context.Background(), creatorID, dto.CreateChannelRequest{
		Name:      "Engineering",
		MemberIDs: []uuid.UUID{uuid.New()},
	})
	if !errors.Is(err, dto.ErrMembersNotFound) {
		t.Fatalf("expected ErrMembersNotFound, got %v", err)
	}
}
