package service

import (
	"errors"

	"github.com/google/uuid"
)

const personalChannelPrefix = "$personal_"

var ErrChannelAccessDenied = errors.New("realtime channel access denied")

type TokenIssuer interface {
	ConnectionToken(userID uuid.UUID) (string, error)
	SubscriptionToken(userID uuid.UUID, channel string) (string, error)
}

type RealtimeService interface {
	ConnectionToken(userID uuid.UUID) (string, error)
	SubscriptionToken(userID uuid.UUID, channel string) (string, error)
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

func (service *realtimeService) ConnectionToken(userID uuid.UUID) (string, error) {
	return service.tokens.ConnectionToken(userID)
}

func (service *realtimeService) SubscriptionToken(userID uuid.UUID, channel string) (string, error) {
	if channel != PersonalChannel(userID) {
		return "", ErrChannelAccessDenied
	}

	return service.tokens.SubscriptionToken(userID, channel)
}
