package logger

import (
	"context"
	"fmt"
	"os"
	"otp-server/internal/infrastructure/config"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

// Logger interface for application logging with context support
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)

	Debugf(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Fatalf(ctx context.Context, format string, args ...interface{})

	WithContext(ctx context.Context) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// F creates a new field
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// zerologLogger implements Logger interface using zerolog
type zerologLogger struct {
	logger zerolog.Logger
}

// New creates a new logger instance
func New(cfg config.LogConfig) Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	var output zerolog.ConsoleWriter
	if cfg.Output == "file" {
		if err := os.MkdirAll("./logs", 0755); err != nil {
			log.Error().Err(err).Msg("Failed to create logs directory")
		}

		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Error().Err(err).Msg("Failed to open log file")
			output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		} else {
			output = zerolog.ConsoleWriter{Out: file, TimeFormat: time.RFC3339}
		}
	} else {
		output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339, NoColor: false}
		output.FormatLevel = func(i interface{}) string {
			lvl := fmt.Sprintf("%s", i)
			switch lvl {
			case "debug":
				return " DBG"
			case "info":
				return "ℹ️  INF"
			case "warn":
				return "⚠️  WRN"
			case "error":
				return "❌ ERR"
			case "fatal":
				return "💀 FTL"
			default:
				return lvl
			}
		}
		output.FormatFieldName = func(i interface{}) string { return fmt.Sprintf("%s=", i) }
		output.FormatFieldValue = func(i interface{}) string { return fmt.Sprintf("%v", i) }
	}

	if cfg.Format == "json" {
		zerolog.TimeFieldFormat = time.RFC3339
		logger := zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()
		return &zerologLogger{logger: logger}
	} else {
		logger := zerolog.New(output).Level(level).With().Timestamp().Logger()
		return &zerologLogger{logger: logger}
	}
}

// getTraceInfo extracts trace information from context
func getTraceInfo(ctx context.Context) map[string]interface{} {
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	fields := make(map[string]interface{})

	if spanContext.IsValid() {
		fields["trace_id"] = spanContext.TraceID().String()
		fields["span_id"] = spanContext.SpanID().String()
	}

	if requestID, ok := ctx.Value("request_id").(string); ok {
		fields["request_id"] = requestID
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		fields["user_id"] = userID
	}

	return fields
}

// Debug logs debug level message with context
func (l *zerologLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	event := l.logger.Debug()

	for key, value := range getTraceInfo(ctx) {
		event = event.Interface(key, value)
	}

	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}

	event.Msg(msg)
}

// Info logs info level message with context
func (l *zerologLogger) Info(ctx context.Context, msg string, fields ...Field) {
	event := l.logger.Info()

	for key, value := range getTraceInfo(ctx) {
		event = event.Interface(key, value)
	}

	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}

	event.Msg(msg)
}

// Warn logs warning level message with context
func (l *zerologLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	event := l.logger.Warn()

	for key, value := range getTraceInfo(ctx) {
		event = event.Interface(key, value)
	}

	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}

	event.Msg(msg)
}

// Error logs error level message with context
func (l *zerologLogger) Error(ctx context.Context, msg string, fields ...Field) {
	event := l.logger.Error()

	for key, value := range getTraceInfo(ctx) {
		event = event.Interface(key, value)
	}

	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}

	event.Msg(msg)
}

// Fatal logs fatal level message with context and exits
func (l *zerologLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	event := l.logger.Fatal()

	for key, value := range getTraceInfo(ctx) {
		event = event.Interface(key, value)
	}

	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}

	event.Msg(msg)
}

// Debugf logs formatted debug level message with context
func (l *zerologLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	event := l.logger.Debug()

	for key, value := range getTraceInfo(ctx) {
		event = event.Interface(key, value)
	}

	event.Msgf(format, args...)
}

// Infof logs formatted info level message with context
func (l *zerologLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	event := l.logger.Info()

	for key, value := range getTraceInfo(ctx) {
		event = event.Interface(key, value)
	}

	event.Msgf(format, args...)
}

// Warnf logs formatted warning level message with context
func (l *zerologLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	event := l.logger.Warn()

	for key, value := range getTraceInfo(ctx) {
		event = event.Interface(key, value)
	}

	event.Msgf(format, args...)
}

// Errorf logs formatted error level message with context
func (l *zerologLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	event := l.logger.Error()

	for key, value := range getTraceInfo(ctx) {
		event = event.Interface(key, value)
	}

	event.Msgf(format, args...)
}

// Fatalf logs formatted fatal level message with context and exits
func (l *zerologLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	event := l.logger.Fatal()

	for key, value := range getTraceInfo(ctx) {
		event = event.Interface(key, value)
	}

	event.Msgf(format, args...)
}

// WithContext creates a new logger with context
func (l *zerologLogger) WithContext(ctx context.Context) Logger {
	logger := l.logger.With()

	for key, value := range getTraceInfo(ctx) {
		logger = logger.Interface(key, value)
	}

	return &zerologLogger{logger: logger.Logger()}
}

// WithField adds a field to the logger
func (l *zerologLogger) WithField(key string, value interface{}) Logger {
	return &zerologLogger{logger: l.logger.With().Interface(key, value).Logger()}
}

// WithFields adds multiple fields to the logger
func (l *zerologLogger) WithFields(fields map[string]interface{}) Logger {
	logger := l.logger.With()
	for key, value := range fields {
		logger = logger.Interface(key, value)
	}
	return &zerologLogger{logger: logger.Logger()}
}

// WithError adds an error to the logger
func (l *zerologLogger) WithError(err error) Logger {
	return &zerologLogger{logger: l.logger.With().Err(err).Logger()}
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context) context.Context {
	requestID := uuid.New().String()
	return context.WithValue(ctx, "request_id", requestID)
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, "user_id", userID)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}
