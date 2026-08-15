package logs

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"emc_lb/src/internal/db/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	loggerMu      sync.Mutex
	systemLogger  *log.Logger
	sourceLoggers map[string]*log.Logger
)

type logEntry map[string]any

func LogOperation(source string, action string, fields map[string]any) {
	entry := logEntry{
		"timestamp": time.Now().Format(time.RFC3339),
		"source":    source,
		"action":    action,
	}

	for key, value := range fields {
		entry[key] = value
	}

	writeLog(getSystemLogger(), entry)
	writeLog(getSourceLogger(source), entry)
}

func LogError(source string, action string, err error, fields map[string]any) {
	entry := logEntry{
		"timestamp": time.Now().Format(time.RFC3339),
		"source":    source,
		"action":    action,
		"error":     err.Error(),
	}

	for key, value := range fields {
		entry[key] = value
	}

	writeLog(getSystemLogger(), entry)
	writeLog(getSourceLogger(source), entry)
}

func getSystemLogger() *log.Logger {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	if systemLogger == nil {
		systemLogger = newFileLogger("system.log")
	}

	return systemLogger
}

func getSourceLogger(source string) *log.Logger {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	if sourceLoggers == nil {
		sourceLoggers = make(map[string]*log.Logger)
	}

	if logger, exists := sourceLoggers[source]; exists {
		return logger
	}

	logger := newFileLogger(source + ".log")
	sourceLoggers[source] = logger
	return logger
}

func WrapDBTX(next sqlc.DBTX) sqlc.DBTX {
	return &loggedDBTX{next: next}
}

type loggedDBTX struct {
	next sqlc.DBTX
}

func (l *loggedDBTX) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	startedAt := time.Now()
	tag, err := l.next.Exec(ctx, query, args...)
	fields := map[string]any{
		"query":       query,
		"args":        args,
		"duration_ms": time.Since(startedAt).Milliseconds(),
		"command_tag": tag.String(),
	}
	if err != nil {
		LogError("database", "exec", err, fields)
		return tag, err
	}

	LogOperation("database", "exec", fields)
	return tag, nil
}

func (l *loggedDBTX) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	startedAt := time.Now()
	rows, err := l.next.Query(ctx, query, args...)
	fields := map[string]any{
		"query":       query,
		"args":        args,
		"duration_ms": time.Since(startedAt).Milliseconds(),
	}
	if err != nil {
		LogError("database", "query", err, fields)
		return rows, err
	}

	LogOperation("database", "query", fields)
	return rows, nil
}

func (l *loggedDBTX) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return &loggedRow{
		next:      l.next.QueryRow(ctx, query, args...),
		query:     query,
		args:      args,
		startedAt: time.Now(),
	}
}

type loggedRow struct {
	next      pgx.Row
	query     string
	args      []interface{}
	startedAt time.Time
}

func (l *loggedRow) Scan(dest ...interface{}) error {
	err := l.next.Scan(dest...)
	fields := map[string]any{
		"query":       l.query,
		"args":        l.args,
		"duration_ms": time.Since(l.startedAt).Milliseconds(),
	}
	if err != nil {
		LogError("database", "query_row", err, fields)
		return err
	}

	LogOperation("database", "query_row", fields)
	return nil
}

func newFileLogger(fileName string) *log.Logger {
	logDirectory := filepath.Join(".", "src", "logs")
	if err := os.MkdirAll(logDirectory, 0o750); err != nil {
		return log.New(os.Stderr, "", 0)
	}

	filePath := filepath.Clean(filepath.Join(logDirectory, fileName))
	// #nosec G304
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return log.New(os.Stderr, "", 0)
	}

	return log.New(file, "", 0)
}

func writeLog(logger *log.Logger, entry logEntry) {
	payload, err := json.Marshal(entry)
	if err != nil {
		logger.Printf(`{"timestamp":"%s","layer":"logger","error":"marshal log entry failed"}`, time.Now().Format(time.RFC3339))
		return
	}

	logger.Println(string(payload))
}
