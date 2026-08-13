package service

import (
	"context"
	"errors"
	"log"
	"strings"

	appLogger "github.com/Hieu3z03/chat-api-golang/pkg/logger"
	"github.com/google/uuid"
)

const personalChannelPrefix = "$personal_"

var ErrChannelAccessDenied = errors.New("realtime channel access denied")

type TokenIssuer interface {
	ConnectionToken(userID uuid.UUID) (string, error)
	SubscriptionToken(userID uuid.UUID, channel string) (string, error)
}

type RealtimeService interface {
	ConnectionToken(ctx context.Context, userID uuid.UUID) (string, error)
	SubscriptionToken(ctx context.Context, userID uuid.UUID, channel string) (string, error)
}

type ChannelRepository interface {
	IsMember(
		ctx context.Context,
		channelID uuid.UUID,
		userID uuid.UUID,
	) (bool, error)
}

type realtimeService struct {
	tokens   TokenIssuer
	channels ChannelRepository
}

func NewRealtimeService(
	tokens TokenIssuer,
	channels ChannelRepository,
) RealtimeService {
	return &realtimeService{
		tokens:   tokens,
		channels: channels,
	}
}

func PersonalChannel(userID uuid.UUID) string {
	return personalChannelPrefix + userID.String()
}

func (service *realtimeService) ConnectionToken(ctx context.Context, userID uuid.UUID) (string, error) {
	defer appLogger.Measure(ctx, "RealtimeService.ConnectionToken")()

	return service.tokens.ConnectionToken(userID)
}

func (service *realtimeService) SubscriptionToken(
	ctx context.Context,
	userID uuid.UUID,
	channel string,
) (string, error) {

	log.Println("USER:", userID)
	log.Println("CHANNEL:", channel)

	channelID := strings.TrimPrefix(channel, "chat:")

	log.Println("CHANNEL ID:", channelID)

	id, err := uuid.Parse(channelID)
	if err != nil {
		log.Println("UUID ERROR:", err)
		return "", ErrChannelAccessDenied
	}

	ok, err := service.channels.IsMember(ctx, id, userID)

	log.Println("IS MEMBER:", ok)
	log.Println("DB ERROR:", err)

	if err != nil {
		return "", err
	}

	if !ok {
		return "", ErrChannelAccessDenied
	}

	return service.tokens.SubscriptionToken(userID, channel)
}
