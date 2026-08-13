package logs

import (
	"fmt"
	"log/slog"
)

type asynqLogger struct {
	logger *slog.Logger
}

func NewAsynqLogger() *asynqLogger {
	return &asynqLogger{
		logger: L(),
	}
}

func (l *asynqLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *asynqLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *asynqLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *asynqLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}
