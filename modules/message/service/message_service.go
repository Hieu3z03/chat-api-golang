package service

import (
	"context"
	"log"
	"strings"
	"time"

	channelDTO "github.com/Hieu3z03/chat-api-golang/modules/channel/dto"
	channelService "github.com/Hieu3z03/chat-api-golang/modules/channel/service"
	"github.com/Hieu3z03/chat-api-golang/modules/message/dto"
	"github.com/Hieu3z03/chat-api-golang/modules/message/model"
	"github.com/Hieu3z03/chat-api-golang/modules/message/repository"
	realtimeService "github.com/Hieu3z03/chat-api-golang/modules/realtime/service"
	"github.com/google/uuid"
)

type MessageService interface {
	Create(
		ctx context.Context,
		channelID uuid.UUID,
		userID uuid.UUID,
		request dto.CreateMessageRequest,
	) (dto.MessageResponse, error)
	List(
		ctx context.Context,
		channelID uuid.UUID,
		userID uuid.UUID,
		limit int64,
		before *time.Time,
	) ([]dto.MessageResponse, error)
}

type messageService struct {
	messages  repository.MessageRepository
	channels  channelService.ChannelService
	publisher EventPublisher
}

type EventPublisher interface {
	Publish(ctx context.Context, channel string, data any) error
}

func NewMessageService(
	messages repository.MessageRepository,
	channels channelService.ChannelService,
	publishers ...EventPublisher,
) MessageService {
	var publisher EventPublisher
	if len(publishers) > 0 {
		publisher = publishers[0]
	}

	return &messageService{
		messages:  messages,
		channels:  channels,
		publisher: publisher,
	}
}

func (service *messageService) Create(
	ctx context.Context,
	channelID uuid.UUID,
	userID uuid.UUID,
	request dto.CreateMessageRequest,
) (dto.MessageResponse, error) {
	channel, err := service.channels.Get(ctx, userID, channelID)
	if err != nil {
		return dto.MessageResponse{}, err
	}

	contentType, ok := request.Content["type"].(string)
	if !ok || strings.TrimSpace(contentType) == "" {
		return dto.MessageResponse{}, dto.ErrInvalidMessageContent
	}

	message, err := service.messages.Create(ctx, model.Message{
		ChannelID: channelID.String(),
		UserID:    userID.String(),
		Content:   request.Content,
	})
	if err != nil {
		return dto.MessageResponse{}, err
	}

	response := toMessageResponse(message, []channelDTO.ChannelUserResponse{})
	service.publishMessageCreated(ctx, channel.Members, response)

	return response, nil
}

func (service *messageService) publishMessageCreated(
	ctx context.Context,
	members []channelDTO.ChannelUserResponse,
	message dto.MessageResponse,
) {
	if service.publisher == nil {
		return
	}

	event := map[string]any{
		"type": "message_added",
		"body": message,
	}
	for _, member := range members {
		channel := realtimeService.PersonalChannel(member.ID)
		if err := service.publisher.Publish(ctx, channel, event); err != nil {
			log.Printf("publish message %s to %s: %v", message.ID, channel, err)
		}
	}
}

func (service *messageService) List(
	ctx context.Context,
	channelID uuid.UUID,
	userID uuid.UUID,
	limit int64,
	before *time.Time,
) ([]dto.MessageResponse, error) {
	messages, err := service.messages.ListByChannel(ctx, channelID.String(), limit, before)
	if err != nil {
		return nil, err
	}

	lastReadSequence := int64(0)
	for _, message := range messages {
		if message.Sequence > lastReadSequence {
			lastReadSequence = message.Sequence
		}
	}

	members, err := service.channels.ListMembersAndMarkRead(ctx, channelID, userID, lastReadSequence)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.MessageResponse, 0, len(messages))
	for _, message := range messages {
		seenBy := make([]channelDTO.ChannelUserResponse, 0, len(members))
		for _, member := range members {
			if member.LastReadSequence >= message.Sequence {
				seenBy = append(seenBy, member.User)
			}
		}

		responses = append(responses, toMessageResponse(message, seenBy))
	}

	return responses, nil
}

func toMessageResponse(
	message model.Message,
	seenBy []channelDTO.ChannelUserResponse,
) dto.MessageResponse {
	return dto.MessageResponse{
		ID:        message.ID.Hex(),
		ChannelID: uuid.MustParse(message.ChannelID),
		UserID:    uuid.MustParse(message.UserID),
		Sequence:  message.Sequence,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
		SeenBy:    seenBy,
	}
}
