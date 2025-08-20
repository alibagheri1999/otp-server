package pool

import (
	"context"
	"sync"
	"time"

	"otp-server/internal/infrastructure/logger"
)

// Connection represents a database connection
type Connection interface {
	Ping(ctx context.Context) error
	Close() error
	IsValid() bool
}

// ConnectionFactory creates new connections
type ConnectionFactory func(ctx context.Context) (Connection, error)

// ConnectionPool manages a pool of database connections
type ConnectionPool struct {
	factory     ConnectionFactory
	maxOpen     int
	maxIdle     int
	maxLifetime time.Duration
	idleTimeout time.Duration

	mu          sync.RWMutex
	connections []Connection
	idle        []Connection
	waiting     []chan Connection
	closed      bool

	logger logger.Logger
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(
	factory ConnectionFactory,
	maxOpen, maxIdle int,
	maxLifetime, idleTimeout time.Duration,
	logger logger.Logger,
) *ConnectionPool {
	pool := &ConnectionPool{
		factory:     factory,
		maxOpen:     maxOpen,
		maxIdle:     maxIdle,
		maxLifetime: maxLifetime,
		idleTimeout: idleTimeout,
		logger:      logger,
	}

	go pool.cleanupRoutine()

	return pool
}

// Get returns a connection from the pool
func (p *ConnectionPool) Get(ctx context.Context) (Connection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, ErrPoolClosed
	}

	if len(p.idle) > 0 {
		conn := p.idle[len(p.idle)-1]
		p.idle = p.idle[:len(p.idle)-1]

		if conn.IsValid() {
			p.connections = append(p.connections, conn)
			return conn, nil
		} else {
			conn.Close()
		}
	}

	if len(p.connections) < p.maxOpen {
		conn, err := p.factory(ctx)
		if err != nil {
			return nil, err
		}
		p.connections = append(p.connections, conn)
		return conn, nil
	}

	wait := make(chan Connection, 1)
	p.waiting = append(p.waiting, wait)
	p.mu.Unlock()

	select {
	case conn := <-wait:
		p.mu.Lock()
		p.connections = append(p.connections, conn)
		return conn, nil
	case <-ctx.Done():
		p.mu.Lock()
		for i, w := range p.waiting {
			if w == wait {
				p.waiting = append(p.waiting[:i], p.waiting[i+1:]...)
				break
			}
		}
		return nil, ctx.Err()
	}
}

// Put returns a connection to the pool
func (p *ConnectionPool) Put(conn Connection) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		conn.Close()
		return
	}

	for i, c := range p.connections {
		if c == conn {
			p.connections = append(p.connections[:i], p.connections[i+1:]...)
			break
		}
	}

	if !conn.IsValid() {
		conn.Close()
		return
	}

	if len(p.idle) < p.maxIdle {
		p.idle = append(p.idle, conn)
	} else {
		conn.Close()
	}

	if len(p.waiting) > 0 {
		wait := p.waiting[0]
		p.waiting = p.waiting[1:]
		wait <- conn
	}
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	for _, conn := range p.connections {
		conn.Close()
	}
	p.connections = nil

	for _, conn := range p.idle {
		conn.Close()
	}
	p.idle = nil

	for _, wait := range p.waiting {
		close(wait)
	}
	p.waiting = nil

	return nil
}

// Stats returns pool statistics
func (p *ConnectionPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PoolStats{
		OpenConnections: len(p.connections),
		IdleConnections: len(p.idle),
		WaitingRequests: len(p.waiting),
		MaxOpen:         p.maxOpen,
		MaxIdle:         p.maxIdle,
	}
}

// cleanupRoutine periodically cleans up expired connections
func (p *ConnectionPool) cleanupRoutine() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.cleanup()
	}
}

// cleanup removes expired connections
func (p *ConnectionPool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	var validIdle []Connection
	for _, conn := range p.idle {
		if conn.IsValid() {
			validIdle = append(validIdle, conn)
		} else {
			conn.Close()
		}
	}
	p.idle = validIdle

	// Clean up active connections that have exceeded maxLifetime
	var validActive []Connection
	for _, conn := range p.connections {
		if conn.IsValid() {
			validActive = append(validActive, conn)
		} else {
			conn.Close()
		}
	}
	p.connections = validActive

	p.logger.Info(context.Background(), "Connection pool cleanup completed",
		logger.F("active_connections", len(p.connections)),
		logger.F("idle_connections", len(p.idle)))
}

// PoolStats represents pool statistics
type PoolStats struct {
	OpenConnections int
	IdleConnections int
	WaitingRequests int
	MaxOpen         int
	MaxIdle         int
}

// Errors
var (
	ErrPoolClosed = &PoolError{Message: "connection pool is closed"}
)

// PoolError represents a pool-related error
type PoolError struct {
	Message string
}

func (e *PoolError) Error() string {
	return e.Message
}
