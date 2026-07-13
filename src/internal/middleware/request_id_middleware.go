package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// HeaderRequestID is the HTTP header key for request tracing.
	HeaderRequestID = "X-Request-ID"

	// ContextRequestIDKey is the gin.Context key for the request ID.
	ContextRequestIDKey = "request_id"
)

// RequestIDMiddleware injects a unique request ID into every request.
//
// If the incoming request already carries an X-Request-ID header (e.g. set
// by the upstream load balancer or the frontend), that value is reused so the
// ID propagates end-to-end across services. Otherwise a new UUID v4 is generated.
//
// The ID is:
//   - stored in gin.Context under ContextRequestIDKey for handlers/services to read
//   - echoed back in the response header X-Request-ID for client-side correlation
func RequestIDMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.GetHeader(HeaderRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx.Set(ContextRequestIDKey, requestID)
		ctx.Header(HeaderRequestID, requestID)
		ctx.Next()
	}
}

// GetRequestID retrieves the request ID from gin.Context.
// Returns an empty string if not set.
func GetRequestID(ctx *gin.Context) string {
	if id, exists := ctx.Get(ContextRequestIDKey); exists {
		if s, ok := id.(string); ok {
			return s
		}
	}

	return ""
}
