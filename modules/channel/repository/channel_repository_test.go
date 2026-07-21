package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Hieu3z03/chat-api-golang/database/entities"
	"github.com/Hieu3z03/chat-api-golang/modules/channel/dto"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateAndListChannelMembers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "requires cgo") {
			t.Skip("SQLite integration test requires CGO")
		}
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&entities.User{}, &entities.Channel{}, &entities.ChannelMember{}); err != nil {
		t.Fatalf("migrate chat schema: %v", err)
	}

	creator := entities.User{ID: uuid.New(), Name: "Channel Creator", Username: "creator"}
	member := entities.User{ID: uuid.New(), Name: "Channel Member", Username: "member"}
	if err := db.Create(&[]entities.User{creator, member}).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}

	repository := NewChannelRepository(db)
	channel, err := repository.Create(context.Background(), entities.Channel{
		Name:      "Engineering",
		CreatedBy: creator.ID,
	}, []uuid.UUID{creator.ID, member.ID})
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}
	if len(channel.Members) != 2 {
		t.Fatalf("expected 2 preloaded members, got %d", len(channel.Members))
	}

	channels, err := repository.ListByUser(context.Background(), member.ID)
	if err != nil {
		t.Fatalf("list channels: %v", err)
	}
	if len(channels) != 1 || channels[0].ID != channel.ID {
		t.Fatalf("expected member to see the created channel")
	}
}

func TestCreateReturnsExistingTwoPersonChannel(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "requires cgo") {
			t.Skip("SQLite integration test requires CGO")
		}
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&entities.User{}, &entities.Channel{}, &entities.ChannelMember{}); err != nil {
		t.Fatalf("migrate chat schema: %v", err)
	}

	creator := entities.User{ID: uuid.New(), Name: "Channel Creator", Username: "creator"}
	member := entities.User{ID: uuid.New(), Name: "Channel Member", Username: "member"}
	if err := db.Create(&[]entities.User{creator, member}).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}

	repository := NewChannelRepository(db)
	_, err = repository.Create(context.Background(), entities.Channel{
		Name:      "Engineering",
		CreatedBy: creator.ID,
	}, []uuid.UUID{creator.ID, member.ID})
	if err != nil {
		t.Fatalf("create first channel: %v", err)
	}

	_, err = repository.Create(context.Background(), entities.Channel{
		Name:      "Engineering",
		CreatedBy: creator.ID,
	}, []uuid.UUID{creator.ID, member.ID})
	if !errors.Is(err, dto.ErrChannelAlreadyExists) {
		t.Fatalf("expected ErrChannelAlreadyExists, got %v", err)
	}

	var count int64
	if err := db.Model(&entities.Channel{}).Count(&count).Error; err != nil {
		t.Fatalf("count channels: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected only one channel for the same two users, got %d", count)
	}
}
