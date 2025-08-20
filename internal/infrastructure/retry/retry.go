package retry

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"

	"otp-server/internal/infrastructure/logger"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	Jitter          bool
	RetryableErrors []error
	OnRetry         func(attempt int, err error)
}

// DefaultConfig returns default retry configuration
func DefaultConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
		OnRetry:       func(attempt int, err error) {},
	}
}

// Retry executes a function with retry logic
func Retry(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if !isRetryableError(err, config.RetryableErrors) {
			return err
		}

		if attempt == config.MaxAttempts {
			break
		}

		config.OnRetry(attempt, err)

		delay := calculateDelay(attempt, config)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return lastErr
}

// RetryWithResult executes a function with retry logic and returns a result
func RetryWithResult[T any](ctx context.Context, config RetryConfig, fn func() (T, error)) (T, error) {
	var lastErr error
	var zero T

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !isRetryableError(err, config.RetryableErrors) {
			return zero, err
		}

		if attempt == config.MaxAttempts {
			break
		}

		config.OnRetry(attempt, err)

		delay := calculateDelay(attempt, config)

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
		}
	}

	return zero, lastErr
}

// calculateDelay calculates the delay for the next retry
func calculateDelay(attempt int, config RetryConfig) time.Duration {
	delay := float64(config.InitialDelay) * math.Pow(config.BackoffFactor, float64(attempt-1))

	if config.Jitter {
		jitter := rand.Float64() * 0.1 * delay // 10% jitter
		delay += jitter
	}

	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	return time.Duration(delay)
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error, retryableErrors []error) bool {
	if len(retryableErrors) == 0 {
		return true
	}

	for _, retryableErr := range retryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	return false
}

// RetryableError represents a retryable error
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NewRetryableError creates a new retryable error
func NewRetryableError(err error) *RetryableError {
	return &RetryableError{Err: err}
}

// RetryWithLogger creates a retry configuration with logging
func RetryWithLogger(log logger.Logger, operation string) RetryConfig {
	config := DefaultConfig()
	config.OnRetry = func(attempt int, err error) {
		log.Warn(context.Background(), "Retrying operation",
			logger.F("operation", operation),
			logger.F("attempt", attempt),
			logger.F("error", err))
	}
	return config
}

// RetryWithExponentialBackoff creates a retry configuration with exponential backoff
func RetryWithExponentialBackoff(maxAttempts int, initialDelay time.Duration) RetryConfig {
	config := DefaultConfig()
	config.MaxAttempts = maxAttempts
	config.InitialDelay = initialDelay
	return config
}

// RetryWithFixedDelay creates a retry configuration with fixed delay
func RetryWithFixedDelay(maxAttempts int, delay time.Duration) RetryConfig {
	config := DefaultConfig()
	config.MaxAttempts = maxAttempts
	config.InitialDelay = delay
	config.BackoffFactor = 1.0
	config.Jitter = false
	return config
}

// RetryWithCustomErrors creates a retry configuration that only retries specific errors
func RetryWithCustomErrors(maxAttempts int, retryableErrors ...error) RetryConfig {
	config := DefaultConfig()
	config.MaxAttempts = maxAttempts
	config.RetryableErrors = retryableErrors
	return config
}
