package log

import (
	"log/slog"
	"os"
	"strings"
)

var defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// Init initializes the global logger (slog.Default) to log JSON to stdout.
func Init(logLevel string) {
	level := parseLevel(logLevel)
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
	}
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(defaultLogger)
}

func Debug(msg string, args ...any) { defaultLogger.Debug(msg, args...) }
func Info(msg string, args ...any)  { defaultLogger.Info(msg, args...) }
func Warn(msg string, args ...any)  { defaultLogger.Warn(msg, args...) }
func Error(msg string, args ...any) { defaultLogger.Error(msg, args...) }

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
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
