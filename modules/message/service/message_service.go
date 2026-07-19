package service

import (
	"context"
	"strings"
	"time"

	channelService "github.com/Hieu3z03/chat-api-golang/modules/channel/service"
	"github.com/Hieu3z03/chat-api-golang/modules/message/dto"
	"github.com/Hieu3z03/chat-api-golang/modules/message/model"
	"github.com/Hieu3z03/chat-api-golang/modules/message/repository"
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
	messages repository.MessageRepository
	channels channelService.ChannelService
}

func NewMessageService(
	messages repository.MessageRepository,
	channels channelService.ChannelService,
) MessageService {
	return &messageService{
		messages: messages,
		channels: channels,
	}
}

func (service *messageService) Create(
	ctx context.Context,
	channelID uuid.UUID,
	userID uuid.UUID,
	request dto.CreateMessageRequest,
) (dto.MessageResponse, error) {
	if err := service.channels.EnsureMember(ctx, channelID, userID); err != nil {
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

	return toMessageResponse(message), nil
}

func (service *messageService) List(
	ctx context.Context,
	channelID uuid.UUID,
	userID uuid.UUID,
	limit int64,
	before *time.Time,
) ([]dto.MessageResponse, error) {
	if err := service.channels.EnsureMember(ctx, channelID, userID); err != nil {
		return nil, err
	}

	messages, err := service.messages.ListByChannel(ctx, channelID.String(), limit, before)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.MessageResponse, 0, len(messages))
	for _, message := range messages {
		responses = append(responses, toMessageResponse(message))
	}

	return responses, nil
}

func toMessageResponse(message model.Message) dto.MessageResponse {
	return dto.MessageResponse{
		ID:        message.ID.Hex(),
		ChannelID: uuid.MustParse(message.ChannelID),
		UserID:    uuid.MustParse(message.UserID),
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}
}
