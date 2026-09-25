package middleware

import (
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"time"

	"emc_lb/src/pkg/logs"

	"github.com/gin-gonic/gin"
)

func RequestLogMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startedAt := time.Now()
		requestBody := readRequestBody(ctx)
		requestID := GetRequestID(ctx)

		logs.L().Info("request_started",
			slog.String("request_id", requestID),
			slog.String("method", ctx.Request.Method),
			slog.String("path", ctx.Request.URL.Path),
			slog.String("ip", ctx.ClientIP()),
			slog.String("body", requestBody),
		)

		ctx.Next()

		userID, _ := ctx.Get(ContextUserIDKey)
		userIDStr, _ := userID.(string)

		logs.L().Info("request_finished",
			slog.String("request_id", requestID),
			slog.String("method", ctx.Request.Method),
			slog.String("path", ctx.Request.URL.Path),
			slog.String("ip", ctx.ClientIP()),
			slog.String("user_id", userIDStr),
			slog.Int("status_code", ctx.Writer.Status()),
			slog.Int64("duration_ms", time.Since(startedAt).Milliseconds()),
		)
	}
}

func readRequestBody(ctx *gin.Context) string {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return ""
	}

	ctx.Request.Body = io.NopCloser(strings.NewReader(string(body)))
	return sanitizeRequestBody(body)
}

func sanitizeRequestBody(body []byte) string {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return string(body)
	}

	// Redact any field that could contain credentials, tokens, or secrets.
	sensitiveFields := []string{
		"password",
		"old_password",
		"new_password",
		"refresh_token",
		"access_token",
		"otp",
		"secret_key",
		"card_number",
		"card_holder_name",
		"card_expiry",
		"cvv",
	}

	for _, field := range sensitiveFields {
		if _, exists := payload[field]; exists {
			payload[field] = "[REDACTED]"
		}
	}

	sanitized, err := json.Marshal(payload)
	if err != nil {
		return string(body)
	}

	return string(sanitized)
}
