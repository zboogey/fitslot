package logger

import (
	"bytes"
	"log/slog"
	"os"
)

var (
	log *slog.Logger
)

func NewJSONHandler(w *bytes.Buffer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return slog.NewJSONHandler(w, opts)
}

func New(handler slog.Handler) *slog.Logger {
	return slog.New(handler)
}

func Init() {
	if log != nil {
		return // Already initialized
	}
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	log = slog.New(handler)
}

func ensureInitialized() {
	if log == nil {
		Init()
	}
}

func Info(msg string, args ...any) {
	ensureInitialized()
	log.Info(msg, args...)
}

func Infof(format string, args ...any) {
	ensureInitialized()
	log.Info(format, args...)
}

func Error(msg string, args ...any) {
	ensureInitialized()
	log.Error(msg, args...)
}

func Errorf(format string, args ...any) {
	ensureInitialized()
	log.Error(format, args...)
}

func Debug(msg string, args ...any) {
	ensureInitialized()
	log.Debug(msg, args...)
}

func Debugf(format string, args ...any) {
	ensureInitialized()
	log.Debug(format, args...)
}

func Fatal(msg string, args ...any) {
	ensureInitialized()
	log.Error(msg, args...)
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	ensureInitialized()
	log.Error(format, args...)
	os.Exit(1)
}

func WithError(err error) *slog.Logger {
	ensureInitialized()
	return log.With("error", err)
}

func WithFields(fields map[string]interface{}) *slog.Logger {
	ensureInitialized()
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return log.With(args...)
}