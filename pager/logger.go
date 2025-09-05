package pager

import (
	"log"
	"time"
)

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type defaultLogger struct {
	level string
}

func newDefaultLogger(level string) Logger {
	return &defaultLogger{level: level}
}

func (l *defaultLogger) Debug(msg string, args ...interface{}) {
	if l.level == "debug" {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

func (l *defaultLogger) Info(msg string, args ...interface{}) {
	if l.level == "debug" || l.level == "info" {
		log.Printf("[INFO] "+msg, args...)
	}
}

func (l *defaultLogger) Warn(msg string, args ...interface{}) {
	if l.level != "error" {
		log.Printf("[WARN] "+msg, args...)
	}
}

func (l *defaultLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

func logQuery(logger Logger, mode string, limit int, order string, rowCount int, elapsed time.Duration) {
	logger.Info("pager query executed",
		"mode", mode,
		"limit", limit,
		"order", order,
		"rows", rowCount,
		"elapsed_ms", elapsed.Milliseconds(),
	)
}