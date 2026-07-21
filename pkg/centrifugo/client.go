package centrifugo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const tokenTTL = time.Hour

type Client struct {
	apiURL     string
	apiKey     string
	tokenKey   []byte
	httpClient *http.Client
}

type apiResponse struct {
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewClient(apiURL, apiKey, tokenHMACSecret string) *Client {
	return &Client{
		apiURL:   strings.TrimRight(apiURL, "/"),
		apiKey:   apiKey,
		tokenKey: []byte(tokenHMACSecret),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (client *Client) ConnectionToken(userID uuid.UUID) (string, error) {
	return client.signToken(jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(tokenTTL).Unix(),
	})
}

func (client *Client) SubscriptionToken(userID uuid.UUID, channel string) (string, error) {
	return client.signToken(jwt.MapClaims{
		"sub":     userID.String(),
		"channel": channel,
		"exp":     time.Now().Add(tokenTTL).Unix(),
	})
}

func (client *Client) signToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(client.tokenKey)
}

func (client *Client) Publish(ctx context.Context, channel string, data any) error {
	payload, err := json.Marshal(map[string]any{
		"method": "publish",
		"params": map[string]any{
			"channel": channel,
			"data":    data,
		},
	})
	if err != nil {
		return fmt.Errorf("encode Centrifugo publish request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create Centrifugo publish request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", client.apiKey)

	response, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("publish to Centrifugo: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("publish to Centrifugo: unexpected HTTP status %d", response.StatusCode)
	}

	var result apiResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode Centrifugo publish response: %w", err)
	}
	if result.Error != nil {
		return fmt.Errorf("publish to Centrifugo: code %d: %s", result.Error.Code, result.Error.Message)
	}

	return nil
}
