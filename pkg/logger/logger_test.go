package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestMeasureCarriesRequestMetadata(t *testing.T) {
	previous := Configured()
	defer Configure(previous)

	var output bytes.Buffer
	ConfigureWriter(Config{
		Level:   zerolog.InfoLevel,
		SQL:     true,
		HTTP:    true,
		Service: "test-api",
	}, &output)

	ctx := WithRequest(context.Background(), "request-123", "/messages")
	ctx = WithUserID(ctx, "user-456")
	Measure(ctx, "MessageRepository.Create")()

	entry := decodeEntry(t, output.Bytes())
	assertField(t, entry, "request_id", "request-123")
	assertField(t, entry, "user_id", "user-456")
	assertField(t, entry, "operation", "MessageRepository.Create")
	if _, ok := entry["duration_ms"]; !ok {
		t.Fatal("duration_ms field was not logged")
	}
}

func TestGORMLoggerMarksSlowQueries(t *testing.T) {
	previous := Configured()
	defer Configure(previous)

	var output bytes.Buffer
	ConfigureWriter(Config{
		Level:              zerolog.InfoLevel,
		SQL:                true,
		HTTP:               true,
		SlowQueryThreshold: 200 * time.Millisecond,
		Service:            "test-api",
	}, &output)

	NewGORM().Trace(
		context.Background(),
		time.Now().Add(-250*time.Millisecond),
		func() (string, int64) { return "SELECT * FROM messages", 3 },
		nil,
	)

	entry := decodeEntry(t, output.Bytes())
	assertField(t, entry, "sql", "SELECT * FROM messages")
	assertField(t, entry, "rows_affected", float64(3))
	assertField(t, entry, "slow_query", true)
	assertField(t, entry, "level", "warn")
	if duration, ok := entry["duration_ms"].(float64); !ok || duration < 200 {
		t.Fatalf("unexpected duration_ms: %#v", entry["duration_ms"])
	}
}

func TestGORMLoggerHonorsSQLFlag(t *testing.T) {
	previous := Configured()
	defer Configure(previous)

	var output bytes.Buffer
	ConfigureWriter(Config{
		Level:   zerolog.InfoLevel,
		SQL:     false,
		HTTP:    true,
		Service: "test-api",
	}, &output)

	NewGORM().Trace(
		context.Background(),
		time.Now(),
		func() (string, int64) { return "SELECT 1", 1 },
		nil,
	)
	if output.Len() != 0 {
		t.Fatalf("expected no SQL log, got %s", output.String())
	}
}

func TestConfigureFromEnv(t *testing.T) {
	previous := Configured()
	defer Configure(previous)

	t.Setenv("APP_NAME", "env-api")
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("LOG_SQL", "false")
	t.Setenv("LOG_HTTP", "false")

	if err := ConfigureFromEnv(); err != nil {
		t.Fatalf("configure from env: %v", err)
	}
	settings := Configured()
	if settings.Service != "env-api" || settings.Level != zerolog.WarnLevel {
		t.Fatalf("unexpected logger identity settings: %+v", settings)
	}
	if settings.SQL || settings.HTTP {
		t.Fatalf("expected SQL and HTTP logging disabled: %+v", settings)
	}
}

func decodeEntry(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(data), &entry); err != nil {
		t.Fatalf("decode log entry: %v\n%s", err, data)
	}
	return entry
}

func assertField(t *testing.T, entry map[string]any, field string, expected any) {
	t.Helper()
	if entry[field] != expected {
		t.Fatalf("expected %s=%#v, got %#v", field, expected, entry[field])
	}
}
