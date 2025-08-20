package logger

import (
	"context"
	"sync"
	"time"
)

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Fields    map[string]interface{}
	Context   context.Context
}

// AsyncLogger provides asynchronous logging capabilities
type AsyncLogger struct {
	baseLogger  Logger
	logChan     chan LogEntry
	workerCount int
	wg          sync.WaitGroup
	stopChan    chan struct{}
}

// NewAsyncLogger creates a new async logger
func NewAsyncLogger(baseLogger Logger, bufferSize, workerCount int) *AsyncLogger {
	al := &AsyncLogger{
		baseLogger:  baseLogger,
		logChan:     make(chan LogEntry, bufferSize),
		workerCount: workerCount,
		stopChan:    make(chan struct{}),
	}

	for i := 0; i < workerCount; i++ {
		al.wg.Add(1)
		go al.worker()
	}

	return al
}

// worker processes log entries from the channel
func (al *AsyncLogger) worker() {
	defer al.wg.Done()

	for {
		select {
		case entry := <-al.logChan:
			al.processLogEntry(entry)
		case <-al.stopChan:
			return
		}
	}
}

// processLogEntry processes a single log entry
func (al *AsyncLogger) processLogEntry(entry LogEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	var fields []Field
	for k, v := range entry.Fields {
		fields = append(fields, F(k, v))
	}

	switch entry.Level {
	case "debug":
		al.baseLogger.Debug(entry.Context, entry.Message, fields...)
	case "info":
		al.baseLogger.Info(entry.Context, entry.Message, fields...)
	case "warn":
		al.baseLogger.Warn(entry.Context, entry.Message, fields...)
	case "error":
		al.baseLogger.Error(entry.Context, entry.Message, fields...)
	}
}

// asyncLog sends a log entry to the channel
func (al *AsyncLogger) asyncLog(level, message string, ctx context.Context, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    fields,
		Context:   ctx,
	}

	select {
	case al.logChan <- entry:
	default:
		al.baseLogger.Warn(ctx, "Async logger buffer full, logging synchronously")
		al.processLogEntry(entry)
	}
}

// Debug logs a debug message asynchronously
func (al *AsyncLogger) Debug(ctx context.Context, message string, fields ...Field) {
	fieldMap := make(map[string]interface{})
	for _, field := range fields {
		fieldMap[field.Key] = field.Value
	}
	al.asyncLog("debug", message, ctx, fieldMap)
}

// Info logs an info message asynchronously
func (al *AsyncLogger) Info(ctx context.Context, message string, fields ...Field) {
	fieldMap := make(map[string]interface{})
	for _, field := range fields {
		fieldMap[field.Key] = field.Value
	}
	al.asyncLog("info", message, ctx, fieldMap)
}

// Warn logs a warning message asynchronously
func (al *AsyncLogger) Warn(ctx context.Context, message string, fields ...Field) {
	fieldMap := make(map[string]interface{})
	for _, field := range fields {
		fieldMap[field.Key] = field.Value
	}
	al.asyncLog("warn", message, ctx, fieldMap)
}

// Error logs an error message asynchronously
func (al *AsyncLogger) Error(ctx context.Context, message string, fields ...Field) {
	fieldMap := make(map[string]interface{})
	for _, field := range fields {
		fieldMap[field.Key] = field.Value
	}
	al.asyncLog("error", message, ctx, fieldMap)
}

// F creates a field for logging
func (al *AsyncLogger) F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Shutdown gracefully shuts down the async logger
func (al *AsyncLogger) Shutdown() {
	close(al.stopChan)
	al.wg.Wait()
	close(al.logChan)
}

// Flush waits for all pending log entries to be processed
func (al *AsyncLogger) Flush() {
	for len(al.logChan) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
	al.wg.Wait()
}
