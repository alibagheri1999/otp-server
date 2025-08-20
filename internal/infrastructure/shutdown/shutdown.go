package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"otp-server/internal/infrastructure/logger"
)

// ShutdownManager manages graceful shutdown of the application
type ShutdownManager struct {
	logger     logger.Logger
	handlers   []ShutdownHandler
	mu         sync.RWMutex
	shutdownCh chan os.Signal
	done       chan struct{}
	timeout    time.Duration
}

// ShutdownHandler defines a shutdown handler interface
type ShutdownHandler interface {
	Name() string
	Shutdown(ctx context.Context) error
	Priority() int
}

// ShutdownFunc is a function-based shutdown handler
type ShutdownFunc struct {
	name     string
	priority int
	fn       func(ctx context.Context) error
}

// NewShutdownManager creates a new shutdown manager
func NewShutdownManager(logger logger.Logger, timeout time.Duration) *ShutdownManager {
	return &ShutdownManager{
		logger:     logger,
		handlers:   make([]ShutdownHandler, 0),
		shutdownCh: make(chan os.Signal, 1),
		done:       make(chan struct{}),
		timeout:    timeout,
	}
}

// AddHandler adds a shutdown handler
func (sm *ShutdownManager) AddHandler(handler ShutdownHandler) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.handlers = append(sm.handlers, handler)
}

// AddFunc adds a function-based shutdown handler
func (sm *ShutdownManager) AddFunc(name string, priority int, fn func(ctx context.Context) error) {
	handler := &ShutdownFunc{
		name:     name,
		priority: priority,
		fn:       fn,
	}
	sm.AddHandler(handler)
}

// Start starts listening for shutdown signals
func (sm *ShutdownManager) Start() {
	signal.Notify(sm.shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sm.shutdownCh
		sm.logger.Info(context.Background(), "Received shutdown signal, starting graceful shutdown")
		sm.Shutdown()
	}()
}

// Shutdown performs graceful shutdown
func (sm *ShutdownManager) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), sm.timeout)
	defer cancel()

	sm.mu.RLock()
	handlers := make([]ShutdownHandler, len(sm.handlers))
	copy(handlers, sm.handlers)
	sm.mu.RUnlock()

	for i := 0; i < len(handlers)-1; i++ {
		for j := i + 1; j < len(handlers); j++ {
			if handlers[i].Priority() > handlers[j].Priority() {
				handlers[i], handlers[j] = handlers[j], handlers[i]
			}
		}
	}

	var wg sync.WaitGroup
	errors := make(chan error, len(handlers))

	for _, handler := range handlers {
		wg.Add(1)
		go func(h ShutdownHandler) {
			defer wg.Done()

			sm.logger.Info(ctx, "Shutting down", logger.F("component", h.Name()))

			if err := h.Shutdown(ctx); err != nil {
				sm.logger.Error(ctx, "Error during shutdown",
					logger.F("component", h.Name()),
					logger.F("error", err))
				errors <- fmt.Errorf("shutdown error for %s: %w", h.Name(), err)
			} else {
				sm.logger.Info(ctx, "Successfully shut down", logger.F("component", h.Name()))
			}
		}(handler)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		sm.logger.Info(ctx, "All shutdown handlers completed")
	case <-ctx.Done():
		sm.logger.Error(ctx, "Shutdown timeout exceeded", logger.F("timeout", sm.timeout))
	}

	close(errors)

	for err := range errors {
		sm.logger.Error(ctx, "Shutdown error", logger.F("error", err))
	}

	close(sm.done)
}

// Wait waits for shutdown to complete
func (sm *ShutdownManager) Wait() {
	<-sm.done
}

// ShutdownFunc implementation
func (sf *ShutdownFunc) Name() string {
	return sf.name
}

func (sf *ShutdownFunc) Priority() int {
	return sf.priority
}

func (sf *ShutdownFunc) Shutdown(ctx context.Context) error {
	return sf.fn(ctx)
}

// Common shutdown priorities
const (
	PriorityHighest = 0
	PriorityHigh    = 10
	PriorityNormal  = 50
	PriorityLow     = 90
	PriorityLowest  = 100
)

// Predefined shutdown handlers
type DatabaseShutdownHandler struct {
	name     string
	shutdown func(ctx context.Context) error
}

func NewDatabaseShutdownHandler(name string, shutdown func(ctx context.Context) error) *DatabaseShutdownHandler {
	return &DatabaseShutdownHandler{
		name:     name,
		shutdown: shutdown,
	}
}

func (h *DatabaseShutdownHandler) Name() string {
	return fmt.Sprintf("database-%s", h.name)
}

func (h *DatabaseShutdownHandler) Priority() int {
	return PriorityHighest
}

func (h *DatabaseShutdownHandler) Shutdown(ctx context.Context) error {
	return h.shutdown(ctx)
}

type CacheShutdownHandler struct {
	name     string
	shutdown func(ctx context.Context) error
}

func NewCacheShutdownHandler(name string, shutdown func(ctx context.Context) error) *CacheShutdownHandler {
	return &CacheShutdownHandler{
		name:     name,
		shutdown: shutdown,
	}
}

func (h *CacheShutdownHandler) Name() string {
	return fmt.Sprintf("cache-%s", h.name)
}

func (h *CacheShutdownHandler) Priority() int {
	return PriorityHigh
}

func (h *CacheShutdownHandler) Shutdown(ctx context.Context) error {
	return h.shutdown(ctx)
}

type ServerShutdownHandler struct {
	name     string
	shutdown func(ctx context.Context) error
}

func NewServerShutdownHandler(name string, shutdown func(ctx context.Context) error) *ServerShutdownHandler {
	return &ServerShutdownHandler{
		name:     name,
		shutdown: shutdown,
	}
}

func (h *ServerShutdownHandler) Name() string {
	return fmt.Sprintf("server-%s", h.name)
}

func (h *ServerShutdownHandler) Priority() int {
	return PriorityNormal
}

func (h *ServerShutdownHandler) Shutdown(ctx context.Context) error {
	return h.shutdown(ctx)
}

type BackgroundWorkerShutdownHandler struct {
	name     string
	shutdown func(ctx context.Context) error
}

func NewBackgroundWorkerShutdownHandler(name string, shutdown func(ctx context.Context) error) *BackgroundWorkerShutdownHandler {
	return &BackgroundWorkerShutdownHandler{
		name:     name,
		shutdown: shutdown,
	}
}

func (h *BackgroundWorkerShutdownHandler) Name() string {
	return fmt.Sprintf("worker-%s", h.name)
}

func (h *BackgroundWorkerShutdownHandler) Priority() int {
	return PriorityLow
}

func (h *BackgroundWorkerShutdownHandler) Shutdown(ctx context.Context) error {
	return h.shutdown(ctx)
}
