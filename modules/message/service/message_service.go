package service

import (
	"context"
	"strings"
	"time"

	channelDTO "github.com/Hieu3z03/chat-api-golang/modules/channel/dto"
	channelService "github.com/Hieu3z03/chat-api-golang/modules/channel/service"
	"github.com/Hieu3z03/chat-api-golang/modules/message/dto"
	"github.com/Hieu3z03/chat-api-golang/modules/message/model"
	"github.com/Hieu3z03/chat-api-golang/modules/message/repository"
	appLogger "github.com/Hieu3z03/chat-api-golang/pkg/logger"
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
	defer appLogger.Measure(ctx, "MessageService.Create")()

	_, err := service.channels.Get(ctx, userID, channelID)
	if err != nil {
		return dto.MessageResponse{}, err
	}

	if strings.TrimSpace(string(request.Type)) == "" || request.Content == nil || strings.TrimSpace(*request.Content) == "" {
		return dto.MessageResponse{}, dto.ErrInvalidMessageContent
	}

	message, err := service.messages.Create(ctx, model.Message{
		ChannelID: channelID.String(),
		UserID:    userID.String(),
		Type:      request.Type,
		Content:   request.Content,
	})
	if err != nil {
		return dto.MessageResponse{}, err
	}

	response := toMessageResponse(message, []channelDTO.ChannelUserResponse{})
	service.publishMessageCreated(ctx, response)

	return response, nil
}

func (service *messageService) publishMessageCreated(
	ctx context.Context,
	message dto.MessageResponse,
) {
	if service.publisher == nil {
		return
	}

	event := map[string]any{
		"type": "message_added",
		"body": message,
	}
	channel := "chat:" + message.ChannelID.String()

	if err := service.publisher.Publish(ctx, channel, event); err != nil {
		appLogger.Log(ctx).
			Error().
			Err(err).
			Str("component", "service").
			Str("message_id", message.ID).
			Str("channel", channel).
			Msg("publish message")
	}
}

func (service *messageService) List(
	ctx context.Context,
	channelID uuid.UUID,
	userID uuid.UUID,
	limit int64,
	before *time.Time,
) ([]dto.MessageResponse, error) {
	defer appLogger.Measure(ctx, "MessageService.List")()

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
		Type:      message.Type,
		Content:   message.Content,
		Sequence:  message.Sequence,
		IsDeleted: message.IsDeleted,
		CreatedAt: message.CreatedAt,
		SeenBy:    seenBy,
	}
}
