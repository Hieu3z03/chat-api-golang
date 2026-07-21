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
	created   model.Message
	messages  []model.Message
	listCalls int
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
	repository.listCalls++
	return repository.messages, nil
}

type fakeChannelService struct {
	memberError       error
	members           []channelDTO.ChannelMemberReadState
	channel           channelDTO.ChannelResponse
	memberListCalls   int
	ensureMemberCalls int
	markedSequence    int64
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
	if service.memberError != nil {
		return channelDTO.ChannelResponse{}, service.memberError
	}
	return service.channel, nil
}

func (service *fakeChannelService) List(context.Context, uuid.UUID) ([]channelDTO.ChannelResponse, error) {
	return nil, nil
}

func (service *fakeChannelService) EnsureMember(context.Context, uuid.UUID, uuid.UUID) error {
	service.ensureMemberCalls++
	return service.memberError
}

func (service *fakeChannelService) ListMembersAndMarkRead(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	lastReadSequence int64,
) ([]channelDTO.ChannelMemberReadState, error) {
	service.memberListCalls++
	service.markedSequence = lastReadSequence
	if service.memberError != nil {
		return nil, service.memberError
	}
	return service.members, nil
}

var _ service.ChannelService = (*fakeChannelService)(nil)

type publishedEvent struct {
	channel string
	data    any
}

type fakeEventPublisher struct {
	events []publishedEvent
}

func (publisher *fakeEventPublisher) Publish(_ context.Context, channel string, data any) error {
	publisher.events = append(publisher.events, publishedEvent{channel: channel, data: data})
	return nil
}

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

func TestCreatePublishesToEveryChannelMember(t *testing.T) {
	channelID := uuid.New()
	firstUserID := uuid.New()
	secondUserID := uuid.New()
	publisher := &fakeEventPublisher{}
	messageService := NewMessageService(
		&fakeMessageRepository{},
		&fakeChannelService{channel: channelDTO.ChannelResponse{Members: []channelDTO.ChannelUserResponse{
			{ID: firstUserID},
			{ID: secondUserID},
		}}},
		publisher,
	)

	_, err := messageService.Create(context.Background(), channelID, firstUserID, dto.CreateMessageRequest{
		Content: map[string]any{"type": "text", "text": "hello"},
	})
	if err != nil {
		t.Fatalf("create message: %v", err)
	}
	if len(publisher.events) != 2 {
		t.Fatalf("expected 2 publications, got %d", len(publisher.events))
	}
	if publisher.events[0].channel != "$personal_"+firstUserID.String() {
		t.Fatalf("unexpected first channel: %s", publisher.events[0].channel)
	}
	if publisher.events[1].channel != "$personal_"+secondUserID.String() {
		t.Fatalf("unexpected second channel: %s", publisher.events[1].channel)
	}
}

func TestListBuildsSeenByFromMembersInMemory(t *testing.T) {
	channelID := uuid.New()
	readerID := uuid.New()
	otherUserID := uuid.New()
	messages := &fakeMessageRepository{messages: []model.Message{
		{ChannelID: channelID.String(), UserID: otherUserID.String(), Sequence: 3},
		{ChannelID: channelID.String(), UserID: otherUserID.String(), Sequence: 2},
	}}
	channels := &fakeChannelService{members: []channelDTO.ChannelMemberReadState{
		{User: channelDTO.ChannelUserResponse{ID: readerID}, LastReadSequence: 3},
		{User: channelDTO.ChannelUserResponse{ID: otherUserID}, LastReadSequence: 2},
	}}
	messageService := NewMessageService(messages, channels)

	responses, err := messageService.List(context.Background(), channelID, readerID, 50, nil)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if messages.listCalls != 1 || channels.memberListCalls != 1 {
		t.Fatalf("expected one bulk read per database, got MongoDB=%d PostgreSQL=%d", messages.listCalls, channels.memberListCalls)
	}
	if channels.ensureMemberCalls != 0 {
		t.Fatalf("list must not issue a separate membership query")
	}
	if channels.markedSequence != 3 {
		t.Fatalf("expected read pointer 3, got %d", channels.markedSequence)
	}
	if len(responses[0].SeenBy) != 1 || responses[0].SeenBy[0].ID != readerID {
		t.Fatalf("unexpected seen_by for sequence 3: %+v", responses[0].SeenBy)
	}
	if len(responses[1].SeenBy) != 2 {
		t.Fatalf("expected both members to have seen sequence 2")
	}
}
