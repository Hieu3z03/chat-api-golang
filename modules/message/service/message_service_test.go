package service

import (
	"context"
	"errors"
	"testing"
	"time"

	channelDTO "github.com/Hieu3z03/chat-api-golang/modules/channel/dto"
	"github.com/Hieu3z03/chat-api-golang/modules/channel/service"
	"github.com/Hieu3z03/chat-api-golang/modules/message/dto"
	"github.com/Hieu3z03/chat-api-golang/modules/message/model"
	"github.com/google/uuid"
)

type fakeMessageRepository struct {
	created model.Message
}

func (repository *fakeMessageRepository) EnsureIndexes(context.Context) error {
	return nil
}

func (repository *fakeMessageRepository) Create(_ context.Context, message model.Message) (model.Message, error) {
	repository.created = message
	message.CreatedAt = time.Now()
	return message, nil
}

func (repository *fakeMessageRepository) ListByChannel(
	context.Context,
	string,
	int64,
	*time.Time,
) ([]model.Message, error) {
	return nil, nil
}

type fakeChannelService struct {
	memberError error
}

func (service *fakeChannelService) Create(
	context.Context,
	uuid.UUID,
	channelDTO.CreateChannelRequest,
) (channelDTO.ChannelResponse, error) {
	return channelDTO.ChannelResponse{}, nil
}

func (service *fakeChannelService) Get(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) (channelDTO.ChannelResponse, error) {
	return channelDTO.ChannelResponse{}, nil
}

func (service *fakeChannelService) List(context.Context, uuid.UUID) ([]channelDTO.ChannelResponse, error) {
	return nil, nil
}

func (service *fakeChannelService) EnsureMember(context.Context, uuid.UUID, uuid.UUID) error {
	return service.memberError
}

var _ service.ChannelService = (*fakeChannelService)(nil)

func TestCreateRequiresMessageType(t *testing.T) {
	messageService := NewMessageService(&fakeMessageRepository{}, &fakeChannelService{})

	_, err := messageService.Create(context.Background(), uuid.New(), uuid.New(), dto.CreateMessageRequest{
		Content: map[string]any{"text": "hello"},
	})
	if !errors.Is(err, dto.ErrInvalidMessageContent) {
		t.Fatalf("expected ErrInvalidMessageContent, got %v", err)
	}
}

func TestCreateRequiresChannelMembership(t *testing.T) {
	messageService := NewMessageService(&fakeMessageRepository{}, &fakeChannelService{
		memberError: channelDTO.ErrNotChannelMember,
	})

	_, err := messageService.Create(context.Background(), uuid.New(), uuid.New(), dto.CreateMessageRequest{
		Content: map[string]any{"type": "text", "text": "hello"},
	})
	if !errors.Is(err, channelDTO.ErrNotChannelMember) {
		t.Fatalf("expected ErrNotChannelMember, got %v", err)
	}
}
