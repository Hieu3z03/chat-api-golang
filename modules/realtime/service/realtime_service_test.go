package service

import (
	"testing"

	"github.com/google/uuid"
)

type fakeTokenIssuer struct {
	subscriptionChannel string
}

func (issuer *fakeTokenIssuer) ConnectionToken(uuid.UUID) (string, error) {
	return "connection-token", nil
}

func (issuer *fakeTokenIssuer) SubscriptionToken(_ uuid.UUID, channel string) (string, error) {
	issuer.subscriptionChannel = channel
	return "subscription-token", nil
}

func TestSubscriptionTokenOnlyAllowsCurrentUsersPersonalChannel(t *testing.T) {
	userID := uuid.New()
	issuer := &fakeTokenIssuer{}
	service := NewRealtimeService(issuer)
	personalChannel := PersonalChannel(userID)

	token, err := service.SubscriptionToken(userID, personalChannel)
	if err != nil {
		t.Fatalf("issue subscription token: %v", err)
	}
	if token != "subscription-token" || issuer.subscriptionChannel != personalChannel {
		t.Fatalf("unexpected token result: token=%q channel=%q", token, issuer.subscriptionChannel)
	}

	if _, err := service.SubscriptionToken(userID, "$personal_"+uuid.NewString()); err != ErrChannelAccessDenied {
		t.Fatalf("expected ErrChannelAccessDenied, got %v", err)
	}
}
