package utils

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

// Logger is a singleton structured logger using slog.
var (
	handler slog.Handler
	logger  *slog.Logger
	once    sync.Once
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
	orgIDKey     contextKey = "org_id"
)

// InitLogger initializes the singleton logger with JSON output.
// Call this once at startup (main.go).
func InitLogger(level string) {
	once.Do(func() {
		var logLevel slog.Level
		switch level {
		case "debug":
			logLevel = slog.LevelDebug
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		default:
			logLevel = slog.LevelInfo
		}

		opts := &slog.HandlerOptions{
			Level: logLevel,
		}

		handler = slog.NewJSONHandler(os.Stdout, opts)
		logger = slog.New(handler)
		slog.SetDefault(logger)
	})
}

// GetLogger returns the singleton logger instance.
func GetLogger() *slog.Logger {
	if logger == nil {
		// Fallback if not initialized — initialize with info level.
		InitLogger("info")
	}
	return logger
}

// WithRequestID returns a new logger with request_id attached.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves request_id from context.
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(requestIDKey); v != nil {
		return v.(string)
	}
	return ""
}

// WithUserID returns a new context with user_id.
func WithUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID retrieves user_id from context.
func GetUserID(ctx context.Context) uint {
	if v := ctx.Value(userIDKey); v != nil {
		return v.(uint)
	}
	return 0
}

// WithOrgID returns a new context with org_id.
func WithOrgID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, orgIDKey, orgID)
}

// GetOrgID retrieves org_id from context.
func GetOrgID(ctx context.Context) string {
	if v := ctx.Value(orgIDKey); v != nil {
		return v.(string)
	}
	return ""
}

// Log is the package-level shorthand for GetLogger().Log.
// Usage: utils.Log(ctx, "INFO", "message", slog.String("key", "value"))
func Log(ctx context.Context, level string, msg string, args ...any) {
	l := GetLogger()
	reqID := GetRequestID(ctx)
	userID := GetUserID(ctx)
	orgID := GetOrgID(ctx)

	args = append(args,
		slog.String("request_id", reqID),
		slog.Uint64("user_id", uint64(userID)),
		slog.String("org_id", orgID),
	)

	switch level {
	case "DEBUG":
		l.Debug(msg, args...)
	case "WARN":
		l.Warn(msg, args...)
	case "ERROR":
		l.Error(msg, args...)
	default:
		l.Info(msg, args...)
	}
}

// DiscardLogger returns a logger that writes to io.Discard (useful for tests).
func DiscardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}
