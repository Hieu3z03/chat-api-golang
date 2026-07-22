package logger

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type GORMLogger struct {
	level         gormLogger.LogLevel
	slowThreshold time.Duration
	enabled       bool
}

func NewGORM() gormLogger.Interface {
	settings := Configured()
	return &GORMLogger{
		level:         gormLogLevel(settings.Level),
		slowThreshold: settings.SlowQueryThreshold,
		enabled:       settings.SQL,
	}
}

func (logger *GORMLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	clone := *logger
	clone.level = level
	return &clone
}

func (logger *GORMLogger) Info(ctx context.Context, message string, args ...interface{}) {
	if !logger.enabled || logger.level < gormLogger.Info {
		return
	}
	Log(ctx).Info().Str("component", "gorm").Msgf(message, args...)
}

func (logger *GORMLogger) Warn(ctx context.Context, message string, args ...interface{}) {
	if !logger.enabled || logger.level < gormLogger.Warn {
		return
	}
	Log(ctx).Warn().Str("component", "gorm").Msgf(message, args...)
}

func (logger *GORMLogger) Error(ctx context.Context, message string, args ...interface{}) {
	if !logger.enabled || logger.level < gormLogger.Error {
		return
	}
	Log(ctx).Error().Str("component", "gorm").Msgf(message, args...)
}

func (logger *GORMLogger) Trace(
	ctx context.Context,
	begin time.Time,
	query func() (string, int64),
	err error,
) {
	if !logger.enabled || logger.level == gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	slow := elapsed > logger.slowThreshold
	if logger.level < gormLogger.Info && !slow && err == nil {
		return
	}

	sql, rows := query()
	event := Log(ctx).Info()
	if slow {
		event = Log(ctx).Warn()
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		event = Log(ctx).Error().Err(err)
	}

	event.
		Str("component", "gorm").
		Str("sql", sql).
		Int64("rows_affected", rows).
		Int64("duration_ms", elapsed.Milliseconds()).
		Bool("slow_query", slow).
		Msg("sql query")
}

func gormLogLevel(level zerolog.Level) gormLogger.LogLevel {
	switch level {
	case zerolog.TraceLevel, zerolog.DebugLevel, zerolog.InfoLevel:
		return gormLogger.Info
	case zerolog.WarnLevel:
		return gormLogger.Warn
	case zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel:
		return gormLogger.Error
	default:
		return gormLogger.Silent
	}
}

var _ gormLogger.Interface = (*GORMLogger)(nil)
