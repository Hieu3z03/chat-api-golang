package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/Hieu3z03/chat-api-golang/database/entities"
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

	creator := entities.User{ID: uuid.New(), FirstName: "Channel", LastName: "Creator", Username: "creator"}
	member := entities.User{ID: uuid.New(), FirstName: "Channel", LastName: "Member", Username: "member"}
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
