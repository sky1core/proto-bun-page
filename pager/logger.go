package pager

import (
    "log/slog"
    "os"
    "time"
)

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type defaultLogger struct {
    l *slog.Logger
}

func newDefaultLogger(level string) Logger {
    var lv slog.Level
    switch level {
    case "debug":
        lv = slog.LevelDebug
    case "info":
        lv = slog.LevelInfo
    case "error":
        lv = slog.LevelError
    case "warn":
        fallthrough
    default:
        lv = slog.LevelWarn
    }
    h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lv})
    return &defaultLogger{l: slog.New(h)}
}

func (l *defaultLogger) Debug(msg string, args ...interface{}) {
    l.l.Debug(msg, args...)
}

func (l *defaultLogger) Info(msg string, args ...interface{}) {
    l.l.Info(msg, args...)
}

func (l *defaultLogger) Warn(msg string, args ...interface{}) {
    l.l.Warn(msg, args...)
}

func (l *defaultLogger) Error(msg string, args ...interface{}) {
    l.l.Error(msg, args...)
}

// NewSlogLoggerAdapter wraps a slog.Logger to satisfy this package's Logger interface.
func NewSlogLoggerAdapter(l *slog.Logger) Logger { return &defaultLogger{l: l} }

func logQuery(logger Logger, mode string, limit int, order string, rowCount int, elapsed time.Duration) {
	logger.Info("pager query executed",
		"mode", mode,
		"limit", limit,
		"order", order,
		"rows", rowCount,
		"elapsed_ms", elapsed.Milliseconds(),
	)
}
