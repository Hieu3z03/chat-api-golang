package service

import (
	"context"
	"errors"

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

type realtimeService struct {
	tokens TokenIssuer
}

func NewRealtimeService(tokens TokenIssuer) RealtimeService {
	return &realtimeService{tokens: tokens}
}

func PersonalChannel(userID uuid.UUID) string {
	return personalChannelPrefix + userID.String()
}

func (service *realtimeService) ConnectionToken(ctx context.Context, userID uuid.UUID) (string, error) {
	defer appLogger.Measure(ctx, "RealtimeService.ConnectionToken")()

	return service.tokens.ConnectionToken(userID)
}

func (service *realtimeService) SubscriptionToken(ctx context.Context, userID uuid.UUID, channel string) (string, error) {
	defer appLogger.Measure(ctx, "RealtimeService.SubscriptionToken")()

	if channel != PersonalChannel(userID) {
		return "", ErrChannelAccessDenied
	}

	return service.tokens.SubscriptionToken(userID, channel)
}
