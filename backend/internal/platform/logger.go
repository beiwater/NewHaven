package platform

import "log/slog"

// Logger is a thin wrapper around slog for structured logging.
// It exists so the rest of the codebase depends on this interface
// rather than directly on slog, allowing future adapter swaps.
type Logger struct {
	inner *slog.Logger
}

func NewLogger(inner *slog.Logger) *Logger {
	return &Logger{inner: inner}
}

func (l *Logger) Info(msg string, args ...any)  { l.inner.Info(msg, args...) }
func (l *Logger) Warn(msg string, args ...any)  { l.inner.Warn(msg, args...) }
func (l *Logger) Error(msg string, args ...any) { l.inner.Error(msg, args...) }
func (l *Logger) Debug(msg string, args ...any) { l.inner.Debug(msg, args...) }
