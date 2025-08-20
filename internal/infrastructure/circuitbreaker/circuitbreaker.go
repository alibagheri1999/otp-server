package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"

	"otp-server/internal/infrastructure/logger"
	"otp-server/internal/infrastructure/redis"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}

// Config holds circuit breaker configuration
type Config struct {
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
	MaxConcurrent    int
	WindowSize       time.Duration
	MinRequestCount  int
}

// DefaultConfig returns default circuit breaker configuration
func DefaultConfig() Config {
	return Config{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          30 * time.Second,
		MaxConcurrent:    2,
		WindowSize:       1 * time.Minute,
		MinRequestCount:  10,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config Config
	state  State
	mu     sync.RWMutex

	// State tracking
	failures        int
	successes       int
	lastFailure     time.Time
	lastStateChange time.Time

	semaphore chan struct{}

	redisClient *redis.Client
	keyPrefix   string

	logger logger.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config Config, redisClient *redis.Client, keyPrefix string, logger logger.Logger) *CircuitBreaker {
	cb := &CircuitBreaker{
		config:      config,
		state:       StateClosed,
		semaphore:   make(chan struct{}, config.MaxConcurrent),
		redisClient: redisClient,
		keyPrefix:   keyPrefix,
		logger:      logger,
	}

	go cb.stateManager()

	return cb
}

// Execute runs a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if !cb.canExecute() {
		return ErrCircuitOpen
	}

	if cb.getState() == StateHalfOpen {
		select {
		case cb.semaphore <- struct{}{}:
			defer func() { <-cb.semaphore }()
		default:
			return ErrCircuitOpen
		}
	}

	err := fn()

	cb.recordResult(err)

	return err
}

// ExecuteAsync runs a function asynchronously with circuit breaker protection
func (cb *CircuitBreaker) ExecuteAsync(ctx context.Context, fn func() error) <-chan error {
	result := make(chan error, 1)

	go func() {
		defer close(result)
		result <- cb.Execute(ctx, fn)
	}()

	return result
}

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has passed
		return time.Since(cb.lastStateChange) >= cb.config.Timeout
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of an operation
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	if err != nil {
		cb.failures++
		cb.lastFailure = now
		cb.logger.Warn(context.Background(), "Circuit breaker failure recorded",
			logger.F("failures", cb.failures),
			logger.F("threshold", cb.config.FailureThreshold))
	} else {
		cb.successes++
		cb.logger.Info(context.Background(), "Circuit breaker success recorded",
			logger.F("successes", cb.successes),
			logger.F("threshold", cb.config.SuccessThreshold))
	}

	cb.updateState()
}

// updateState updates the circuit breaker state based on current conditions
func (cb *CircuitBreaker) updateState() {
	now := time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.FailureThreshold {
			cb.transitionTo(StateOpen, now)
		}
	case StateOpen:
		if time.Since(cb.lastStateChange) >= cb.config.Timeout {
			cb.transitionTo(StateHalfOpen, now)
		}
	case StateHalfOpen:
		if cb.successes >= cb.config.SuccessThreshold {
			cb.transitionTo(StateClosed, now)
		} else if cb.failures >= cb.config.FailureThreshold {
			cb.transitionTo(StateOpen, now)
		}
	}
}

// transitionTo transitions to a new state
func (cb *CircuitBreaker) transitionTo(newState State, timestamp time.Time) {
	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = timestamp

	if newState == StateClosed {
		cb.failures = 0
		cb.successes = 0
	} else if newState == StateHalfOpen {
		cb.failures = 0
		cb.successes = 0
	}

	// Log state transition
	cb.logger.Info(context.Background(), "Circuit breaker state transition",
		logger.F("old_state", oldState.String()),
		logger.F("new_state", newState.String()),
		logger.F("failures", cb.failures),
		logger.F("successes", cb.successes))

	if cb.redisClient != nil {
		cb.updateRedisState(newState, timestamp)
	}
}

// updateRedisState updates the circuit breaker state in Redis
func (cb *CircuitBreaker) updateRedisState(state State, timestamp time.Time) {
	ctx := context.Background()
	key := cb.keyPrefix + ":state"

	stateData := map[string]interface{}{
		"state":             state.String(),
		"timestamp":         timestamp.Unix(),
		"failures":          cb.failures,
		"successes":         cb.successes,
		"last_failure":      cb.lastFailure.Unix(),
		"last_state_change": timestamp.Unix(),
	}

	pipe := cb.redisClient.GetClient().Pipeline()
	pipe.HMSet(ctx, key, stateData)
	pipe.Expire(ctx, key, 24*time.Hour)

	_, err := pipe.Exec(ctx)
	if err != nil {
		cb.logger.Error(ctx, "Failed to update Redis state", logger.F("error", err))
	}
}

// stateManager manages circuit breaker state in the background
func (cb *CircuitBreaker) stateManager() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		cb.mu.Lock()

		cb.updateState()

		if cb.state == StateClosed && cb.failures > 0 {
			if time.Since(cb.lastFailure) > cb.config.WindowSize {
				cb.failures = 0
				cb.logger.Info(context.Background(), "Circuit breaker failure window reset")
			}
		}

		cb.mu.Unlock()
	}
}

// getState returns the current state
func (cb *CircuitBreaker) getState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return Stats{
		State:           cb.state.String(),
		Failures:        cb.failures,
		Successes:       cb.successes,
		LastFailure:     cb.lastFailure,
		LastStateChange: cb.lastStateChange,
		Config:          cb.config,
	}
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.transitionTo(StateOpen, time.Now())
}

// ForceClose forces the circuit breaker to closed state
func (cb *CircuitBreaker) ForceClose() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.transitionTo(StateClosed, time.Now())
}

// Reset resets the circuit breaker to initial state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.successes = 0
	cb.lastFailure = time.Time{}
	cb.lastStateChange = time.Time{}
	cb.transitionTo(StateClosed, time.Now())
}

// Stats represents circuit breaker statistics
type Stats struct {
	State           string    `json:"state"`
	Failures        int       `json:"failures"`
	Successes       int       `json:"successes"`
	LastFailure     time.Time `json:"last_failure"`
	LastStateChange time.Time `json:"last_state_change"`
	Config          Config    `json:"config"`
}

// Errors
var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	circuitBreakers map[string]*CircuitBreaker
	mu              sync.RWMutex
	logger          logger.Logger
}

// NewManager creates a new circuit breaker manager
func NewManager(logger logger.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		circuitBreakers: make(map[string]*CircuitBreaker),
		logger:          logger,
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (m *CircuitBreakerManager) GetOrCreate(name string, config Config) *CircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cb, exists := m.circuitBreakers[name]; exists {
		return cb
	}

	// Create new circuit breaker with default Redis client (nil for now)
	cb := NewCircuitBreaker(config, nil, name, m.logger)
	m.circuitBreakers[name] = cb
	return cb
}

// Get returns a circuit breaker by name
func (m *CircuitBreakerManager) Get(name string) (*CircuitBreaker, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cb, exists := m.circuitBreakers[name]
	return cb, exists
}

// Remove removes a circuit breaker by name
func (m *CircuitBreakerManager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.circuitBreakers, name)
}

// GetAll returns all circuit breakers
func (m *CircuitBreakerManager) GetAll() map[string]*CircuitBreaker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*CircuitBreaker)
	for k, v := range m.circuitBreakers {
		result[k] = v
	}
	return result
}
