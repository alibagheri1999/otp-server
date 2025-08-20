package database

import (
	"context"
	"database/sql"
	"fmt"

	"otp-server/internal/infrastructure/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB wraps the PostgreSQL connection pool
type PostgresDB struct {
	pool *pgxpool.Pool
}

// NewPostgresConnection creates a new PostgreSQL connection pool
func NewPostgresConnection(cfg config.DatabaseConfig) (*PostgresDB, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{pool: pool}, nil
}

// Close closes the connection pool
func (db *PostgresDB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// GetPool returns the underlying connection pool
func (db *PostgresDB) GetPool() *pgxpool.Pool {
	return db.pool
}

// Exec executes a query without returning rows
func (db *PostgresDB) Exec(ctx context.Context, sql string, arguments ...interface{}) error {
	_, err := db.pool.Exec(ctx, sql, arguments...)
	return err
}

// Query executes a query that returns rows
func (db *PostgresDB) Query(ctx context.Context, sql string, arguments ...interface{}) (*sql.Rows, error) {
	return nil, fmt.Errorf("not implemented: pgx.Rows is not compatible with sql.Rows")
}

// QueryRow executes a query that returns a single row
func (db *PostgresDB) QueryRow(ctx context.Context, sql string, arguments ...interface{}) *sql.Row {
	return nil
}

// Begin starts a new transaction
func (db *PostgresDB) Begin(ctx context.Context) (*sql.Tx, error) {
	return nil, fmt.Errorf("not implemented: pgx.Tx is not compatible with sql.Tx")
}

// Ping pings the database
func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}
