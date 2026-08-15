package logs

import (
	"fmt"
	"log/slog"
)

type AsynqLogger struct {
	logger *slog.Logger
}

func NewAsynqLogger() *AsynqLogger {
	return &AsynqLogger{
		logger: L(),
	}
}

func (l *AsynqLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *AsynqLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *AsynqLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *AsynqLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *AsynqLogger) Fatal(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}
