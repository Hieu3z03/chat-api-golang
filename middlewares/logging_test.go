package middlewares

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	appLogger "github.com/Hieu3z03/chat-api-golang/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func TestHTTPLoggerIncludesRequestAndUserMetadata(t *testing.T) {
	previous := appLogger.Configured()
	defer appLogger.Configure(previous)

	var output bytes.Buffer
	appLogger.ConfigureWriter(appLogger.Config{
		Level:   zerolog.InfoLevel,
		SQL:     true,
		HTTP:    true,
		Service: "test-api",
	}, &output)

	gin.SetMode(gin.TestMode)
	router := loggingTestRouter()
	router.GET("/messages/:id", RequireIdentityHeaders(), func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/messages/42", nil)
	request.Header.Set("X-Request-ID", "request-123")
	request.Header.Set("X-User-ID", "11111111-1111-4111-8111-111111111111")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Header().Get("X-Request-ID") != "request-123" {
		t.Fatalf("unexpected response request ID: %q", response.Header().Get("X-Request-ID"))
	}
	entries := decodeLogEntries(t, output.Bytes())
	if len(entries) != 1 {
		t.Fatalf("expected one HTTP log, got %d", len(entries))
	}
	entry := entries[0]
	assertLogField(t, entry, "request_id", "request-123")
	assertLogField(t, entry, "user_id", "11111111-1111-4111-8111-111111111111")
	assertLogField(t, entry, "method", http.MethodGet)
	assertLogField(t, entry, "path", "/messages/42")
	assertLogField(t, entry, "endpoint", "/messages/:id")
	assertLogField(t, entry, "status_code", float64(http.StatusNoContent))
	if _, ok := entry["response_time_ms"]; !ok {
		t.Fatal("response_time_ms field was not logged")
	}
	if _, ok := entry["client_ip"]; !ok {
		t.Fatal("client_ip field was not logged")
	}
}

func TestErrorLoggerIncludesEndpointAndIdentity(t *testing.T) {
	previous := appLogger.Configured()
	defer appLogger.Configure(previous)

	var output bytes.Buffer
	appLogger.ConfigureWriter(appLogger.Config{
		Level:   zerolog.InfoLevel,
		SQL:     true,
		HTTP:    false,
		Service: "test-api",
	}, &output)

	gin.SetMode(gin.TestMode)
	router := loggingTestRouter()
	router.GET("/messages/:id", RequireIdentityHeaders(), func(ctx *gin.Context) {
		RecordError(ctx, errors.New("database unavailable"))
		ctx.Status(http.StatusInternalServerError)
	})

	request := httptest.NewRequest(http.MethodGet, "/messages/42", nil)
	request.Header.Set("X-Request-ID", "request-456")
	request.Header.Set("X-User-ID", "22222222-2222-4222-8222-222222222222")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	entries := decodeLogEntries(t, output.Bytes())
	if len(entries) != 1 {
		t.Fatalf("expected one error log, got %d", len(entries))
	}
	entry := entries[0]
	assertLogField(t, entry, "error", "database unavailable")
	assertLogField(t, entry, "request_id", "request-456")
	assertLogField(t, entry, "user_id", "22222222-2222-4222-8222-222222222222")
	assertLogField(t, entry, "endpoint", "/messages/:id")
}

func loggingTestRouter() *gin.Engine {
	router := gin.New()
	router.Use(RequestID())
	router.Use(HTTPLogger())
	router.Use(ErrorLogger())
	router.Use(Recovery())
	return router
}

func decodeLogEntries(t *testing.T, data []byte) []map[string]any {
	t.Helper()
	var entries []map[string]any
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var entry map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("decode log entry: %v\n%s", err, scanner.Bytes())
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan logs: %v", err)
	}
	return entries
}

func assertLogField(t *testing.T, entry map[string]any, field string, expected any) {
	t.Helper()
	if entry[field] != expected {
		t.Fatalf("expected %s=%#v, got %#v", field, expected, entry[field])
	}
}
