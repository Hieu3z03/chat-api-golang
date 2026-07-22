package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	appLogger "github.com/Hieu3z03/chat-api-golang/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := strings.TrimSpace(ctx.GetHeader(requestIDHeader))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx.Header(requestIDHeader, requestID)
		requestContext := appLogger.WithRequest(
			ctx.Request.Context(),
			requestID,
			ctx.Request.URL.Path,
		)
		ctx.Request = ctx.Request.WithContext(requestContext)
		ctx.Next()
	}
}

func HTTPLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !appLogger.Configured().HTTP {
			ctx.Next()
			return
		}

		started := time.Now()
		ctx.Next()

		endpoint := ctx.FullPath()
		if endpoint == "" {
			endpoint = ctx.Request.URL.Path
		}
		elapsed := time.Since(started)
		statusCode := ctx.Writer.Status()

		event := appLogger.Log(ctx.Request.Context()).Info()
		if statusCode >= http.StatusBadRequest {
			event = appLogger.Log(ctx.Request.Context()).Warn()
		}

		event.
			Str("component", "http").
			Str("method", ctx.Request.Method).
			Str("path", ctx.Request.URL.Path).
			Str("endpoint", endpoint).
			Int("status_code", statusCode).
			Int64("response_time_ms", elapsed.Milliseconds()).
			Str("client_ip", ctx.ClientIP()).
			Msg("http request")
	}
}

func ErrorLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		statusCode := ctx.Writer.Status()
		if len(ctx.Errors) == 0 && statusCode < http.StatusBadRequest {
			return
		}

		endpoint := ctx.FullPath()
		if endpoint == "" {
			endpoint = ctx.Request.URL.Path
		}

		requestError := errors.New(http.StatusText(statusCode))
		var metadata any
		if lastError := ctx.Errors.Last(); lastError != nil {
			requestError = lastError.Err
			metadata = lastError.Meta
		}

		event := appLogger.Log(ctx.Request.Context()).Error().
			Err(requestError).
			Str("component", "http").
			Str("endpoint", endpoint).
			Int("status_code", statusCode)
		if metadata != nil {
			event = event.Interface("error_meta", metadata)
		}
		event.Msg("request failed")
	}
}

func Recovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				captured := ctx.Error(fmt.Errorf("panic: %v", recovered))
				captured.Meta = string(debug.Stack())
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		ctx.Next()
	}
}

func RecordError(ctx *gin.Context, err error) {
	if err != nil {
		_ = ctx.Error(err)
	}
}
