package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"otp-server/internal/infrastructure/config"
	"otp-server/internal/infrastructure/logger"

	_ "github.com/lib/pq"
)

// PostgresPool manages PostgreSQL connections
type PostgresPool struct {
	config *config.DatabaseConfig
	logger logger.Logger
	db     *sql.DB
	mu     sync.RWMutex
	stats  PoolStats
	closed bool
}

// PoolStats holds pool statistics
type PoolStats struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDuration       time.Duration
	MaxIdleClosed      int64
	MaxLifetimeClosed  int64
}

// NewPostgresPool creates a new PostgreSQL connection pool
func NewPostgresPool(cfg *config.DatabaseConfig, logger logger.Logger) (*PostgresPool, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pool := &PostgresPool{
		config: cfg,
		logger: logger,
		db:     db,
		stats:  PoolStats{},
	}

	go pool.collectStats()

	logger.Info(context.Background(), "PostgreSQL connection pool created")

	return pool, nil
}

// GetConnection gets a connection from the pool
func (p *PostgresPool) GetConnection(ctx context.Context) (*sql.DB, error) {
	if p.isClosed() {
		return nil, fmt.Errorf("connection pool is closed")
	}

	if err := p.db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return p.db, nil
}

// Close closes the connection pool
func (p *PostgresPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	err := p.db.Close()

	p.logger.Info(context.Background(), "PostgreSQL connection pool closed")
	return err
}

// Stats returns pool statistics
func (p *PostgresPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := p.db.Stats()
	p.stats.MaxOpenConnections = stats.MaxOpenConnections
	p.stats.OpenConnections = stats.OpenConnections
	p.stats.InUse = stats.InUse
	p.stats.Idle = stats.Idle
	p.stats.WaitCount = stats.WaitCount
	p.stats.WaitDuration = stats.WaitDuration
	p.stats.MaxIdleClosed = stats.MaxIdleClosed
	p.stats.MaxLifetimeClosed = stats.MaxLifetimeClosed

	return p.stats
}

// isClosed checks if the pool is closed
func (p *PostgresPool) isClosed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.closed
}

// collectStats periodically collects pool statistics
func (p *PostgresPool) collectStats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if p.isClosed() {
				return
			}
			stats := p.Stats()
			p.logger.Debug(context.Background(), "PostgreSQL pool stats",
				logger.F("open_connections", stats.OpenConnections),
				logger.F("in_use", stats.InUse),
				logger.F("idle", stats.Idle),
				logger.F("wait_count", stats.WaitCount))
		}
	}
}

// HealthCheck performs a health check on the database
func (p *PostgresPool) HealthCheck(ctx context.Context) error {
	if p.isClosed() {
		return fmt.Errorf("connection pool is closed")
	}

	return p.db.PingContext(ctx)
}

// BeginTransaction begins a new transaction
func (p *PostgresPool) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	if p.isClosed() {
		return nil, fmt.Errorf("connection pool is closed")
	}

	return p.db.BeginTx(ctx, nil)
}

// Exec executes a query without returning rows
func (p *PostgresPool) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if p.isClosed() {
		return nil, fmt.Errorf("connection pool is closed")
	}

	return p.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (p *PostgresPool) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if p.isClosed() {
		return nil, fmt.Errorf("connection pool is closed")
	}

	return p.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (p *PostgresPool) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if p.isClosed() {
		return nil
	}

	return p.db.QueryRowContext(ctx, query, args...)
}
