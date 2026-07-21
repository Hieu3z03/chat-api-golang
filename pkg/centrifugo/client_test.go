package centrifugo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestConnectionAndSubscriptionTokens(t *testing.T) {
	const secret = "test-secret"
	client := NewClient("http://localhost/api", "api-key", secret)
	userID := uuid.New()
	channel := "$personal_" + userID.String()

	connectionToken, err := client.ConnectionToken(userID)
	if err != nil {
		t.Fatalf("connection token: %v", err)
	}
	connectionClaims := parseTokenClaims(t, connectionToken, secret)
	if connectionClaims["sub"] != userID.String() {
		t.Fatalf("unexpected connection subject: %v", connectionClaims["sub"])
	}

	subscriptionToken, err := client.SubscriptionToken(userID, channel)
	if err != nil {
		t.Fatalf("subscription token: %v", err)
	}
	subscriptionClaims := parseTokenClaims(t, subscriptionToken, secret)
	if subscriptionClaims["channel"] != channel {
		t.Fatalf("unexpected subscription channel: %v", subscriptionClaims["channel"])
	}
}

func TestPublishUsesCentrifugoHTTPAPI(t *testing.T) {
	const apiKey = "test-api-key"
	var requestBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-API-Key") != apiKey {
			t.Fatalf("unexpected API key: %q", request.Header.Get("X-API-Key"))
		}
		if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"result":{"offset":1,"epoch":"test"}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, apiKey, "secret")
	if err := client.Publish(context.Background(), "$personal_user", map[string]any{"type": "message_added"}); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if requestBody["method"] != "publish" {
		t.Fatalf("unexpected method: %v", requestBody["method"])
	}
	params, ok := requestBody["params"].(map[string]any)
	if !ok || params["channel"] != "$personal_user" {
		t.Fatalf("unexpected params: %#v", requestBody["params"])
	}
}

func parseTokenClaims(t *testing.T, tokenString, secret string) jwt.MapClaims {
	t.Helper()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("parse token: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("unexpected claims type %T", token.Claims)
	}
	return claims
}
