package middleware

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"emc_lb/src/pkg/logs"

	"github.com/gin-gonic/gin"
)

func RequestLogMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startedAt := time.Now()
		requestBody := readRequestBody(ctx)

		logs.LogOperation("handler", "request_started", map[string]any{
			"method": ctx.Request.Method,
			"path":   ctx.Request.URL.Path,
			"ip":     ctx.ClientIP(),
			"body":   requestBody,
		})

		ctx.Next()

		logs.LogOperation("handler", "request_finished", map[string]any{
			"method":      ctx.Request.Method,
			"path":        ctx.Request.URL.Path,
			"ip":          ctx.ClientIP(),
			"status_code": ctx.Writer.Status(),
			"duration_ms": time.Since(startedAt).Milliseconds(),
		})
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

	if _, exists := payload["password"]; exists {
		payload["password"] = "[REDACTED]"
	}

	sanitized, err := json.Marshal(payload)
	if err != nil {
		return string(body)
	}

	return string(sanitized)
}
