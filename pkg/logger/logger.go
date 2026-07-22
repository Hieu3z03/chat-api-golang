package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const defaultSlowQueryThreshold = 200 * time.Millisecond

type Config struct {
	Level              zerolog.Level
	SQL                bool
	HTTP               bool
	SlowQueryThreshold time.Duration
	Service            string
}

var (
	configMu sync.RWMutex
	config   = Config{
		Level:              zerolog.InfoLevel,
		SQL:                true,
		HTTP:               true,
		SlowQueryThreshold: defaultSlowQueryThreshold,
		Service:            "chat-api",
	}
	baseLogger = newLogger(config, os.Stdout)
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
	endpointKey  contextKey = "endpoint"
)

func ConfigureFromEnv() error {
	level, err := zerolog.ParseLevel(getEnv("LOG_LEVEL", "info"))
	if err != nil {
		return fmt.Errorf("parse LOG_LEVEL: %w", err)
	}

	sqlLogging, err := getBoolEnv("LOG_SQL", true)
	if err != nil {
		return err
	}
	httpLogging, err := getBoolEnv("LOG_HTTP", true)
	if err != nil {
		return err
	}

	Configure(Config{
		Level:              level,
		SQL:                sqlLogging,
		HTTP:               httpLogging,
		SlowQueryThreshold: defaultSlowQueryThreshold,
		Service:            getEnv("APP_NAME", "chat-api"),
	})
	return nil
}

func Configure(next Config) {
	configure(next, os.Stdout)
}

func ConfigureWriter(next Config, writer io.Writer) {
	configure(next, writer)
}

func configure(next Config, writer io.Writer) {
	if next.SlowQueryThreshold <= 0 {
		next.SlowQueryThreshold = defaultSlowQueryThreshold
	}
	if next.Service == "" {
		next.Service = "chat-api"
	}

	configMu.Lock()
	config = next
	baseLogger = newLogger(next, writer)
	configMu.Unlock()
}

func newLogger(settings Config, writer io.Writer) zerolog.Logger {
	return zerolog.New(zerolog.SyncWriter(writer)).
		Level(settings.Level).
		With().
		Timestamp().
		Str("service", settings.Service).
		Logger()
}

func Configured() Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return config
}

func Global() zerolog.Logger {
	configMu.RLock()
	defer configMu.RUnlock()
	return baseLogger
}

func Log(ctx context.Context) *zerolog.Logger {
	if ctx != nil {
		if contextual := zerolog.Ctx(ctx); contextual.GetLevel() != zerolog.Disabled {
			return contextual
		}
	}

	global := Global()
	return &global
}

func WithRequest(ctx context.Context, requestID, endpoint string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	contextual := Global()
	contextual = contextual.With().
		Str("request_id", requestID).
		Logger()
	if requestID != "" {
		ctx = context.WithValue(ctx, requestIDKey, requestID)
	}
	if endpoint != "" {
		ctx = context.WithValue(ctx, endpointKey, endpoint)
	}
	return contextual.WithContext(ctx)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	contextual := *Log(ctx)
	contextual = contextual.With().Str("user_id", userID).Logger()
	return context.WithValue(contextual.WithContext(ctx), userIDKey, userID)
}

func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}

func UserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(userIDKey).(string)
	return value
}

func Endpoint(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(endpointKey).(string)
	return value
}

func Measure(ctx context.Context, operation string) func() {
	started := time.Now()
	return func() {
		elapsed := time.Since(started)
		Log(ctx).Info().
			Str("component", "timing").
			Str("operation", operation).
			Int64("duration_ms", elapsed.Milliseconds()).
			Msg(operation + " took " + elapsed.Round(time.Microsecond).String())
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) (bool, error) {
	value := getEnv(key, strconv.FormatBool(fallback))
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}
