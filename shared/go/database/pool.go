// Package database provides shared database utilities for QSGW services.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PoolConfig holds connection pool configuration.
type PoolConfig struct {
	// DatabaseURL is the PostgreSQL connection string. Required.
	DatabaseURL string

	// MaxConns is the maximum number of connections in the pool.
	// Default: 25. For most services, 10-50 is appropriate.
	MaxConns int32

	// MinConns is the minimum number of idle connections maintained.
	// Default: 5. Setting this ensures connections are ready for bursts.
	MinConns int32

	// MaxConnLifetime is the maximum lifetime of a connection.
	// Default: 1 hour. Prevents stale connections to rotated databases.
	MaxConnLifetime time.Duration

	// MaxConnIdleTime is the maximum time a connection can be idle.
	// Default: 30 minutes. Frees unused connections.
	MaxConnIdleTime time.Duration

	// HealthCheckPeriod is how often idle connections are health-checked.
	// Default: 1 minute.
	HealthCheckPeriod time.Duration

	// ConnectTimeout is the timeout for establishing a new connection.
	// Default: 5 seconds.
	ConnectTimeout time.Duration

	// Logger for pool events.
	Logger *zap.Logger
}

// DefaultPoolConfig returns a production-ready pool configuration.
func DefaultPoolConfig(databaseURL string) PoolConfig {
	return PoolConfig{
		DatabaseURL:       databaseURL,
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   1 * time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
		ConnectTimeout:    5 * time.Second,
	}
}

// NewPool creates a new pgxpool with the given configuration.
// It pings the database to verify connectivity before returning.
func NewPool(ctx context.Context, cfg PoolConfig) (*pgxpool.Pool, error) {
	logger := cfg.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	// Apply pool settings
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	if cfg.ConnectTimeout > 0 {
		poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Verify connectivity
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	logger.Info("database pool connected",
		zap.Int32("max_conns", cfg.MaxConns),
		zap.Int32("min_conns", cfg.MinConns),
		zap.Duration("max_conn_lifetime", cfg.MaxConnLifetime),
		zap.Duration("max_conn_idle_time", cfg.MaxConnIdleTime),
		zap.Duration("health_check_period", cfg.HealthCheckPeriod),
	)

	return pool, nil
}
