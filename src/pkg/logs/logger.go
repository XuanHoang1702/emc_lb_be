// Package logs provides a structured, context-aware logger backed by the
// standard library slog. It writes JSON in production and human-readable
// text in development, automatically including request_id when available.
package logs

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	contextKeyRequestID = "request_id"
	contextKeyUserID    = "user_id"
)

var (
	globalLogger     *slog.Logger
	globalLoggerOnce sync.Once
)

// contextKey is an unexported type to avoid key collisions in context.
type contextKey string

const (
	CtxRequestID contextKey = "request_id"
	CtxUserID    contextKey = "user_id"
)

// Init initializes the global structured logger.
//
//   - level: "debug" | "info" | "warn" | "error"
//   - env:   "release" for JSON output, anything else for text output
//
// Should be called once at application startup (in main or app.New).
func Init(level, env string) {
	globalLoggerOnce.Do(func() {
		globalLogger = newLogger(level, env, os.Stdout)
	})
}

// L returns the global logger. Falls back to a default info-level logger if
// Init has not been called yet (safe for tests).
func L() *slog.Logger {
	if globalLogger == nil {
		return slog.Default()
	}

	return globalLogger
}

// WithContext returns a logger enriched with context values (request_id, user_id).
func WithContext(ctx context.Context) *slog.Logger {
	logger := L()

	if ctx == nil {
		return logger
	}

	if id, ok := ctx.Value(CtxRequestID).(string); ok && id != "" {
		logger = logger.With(contextKeyRequestID, id)
	}

	if uid, ok := ctx.Value(CtxUserID).(string); ok && uid != "" {
		logger = logger.With(contextKeyUserID, uid)
	}

	return logger
}

// WithRequestID returns a new context with the request ID set.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, CtxRequestID, requestID)
}

// WithUserID returns a new context with the user ID set.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, CtxUserID, userID)
}

// newLogger constructs an slog.Logger writing to w.
func newLogger(levelStr, env string, w io.Writer) *slog.Logger {
	level := parseLevel(levelStr)

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if strings.EqualFold(env, "release") {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	return slog.New(handler)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ensureLogDir creates the logs directory if it does not exist.
func ensureLogDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// NewFileLogger creates an slog.Logger that writes to a file inside dir.
// Used by components that need domain-specific log files (e.g. DB queries).
func NewFileLogger(dir, filename, levelStr, env string) (*slog.Logger, error) {
	if err := ensureLogDir(dir); err != nil {
		return nil, err
	}

	path := filepath.Join(dir, filename)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	return newLogger(levelStr, env, f), nil
}
