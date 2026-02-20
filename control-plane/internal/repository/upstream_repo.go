package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quantun-opensource/qsgw/control-plane/internal/model"
)

type UpstreamRepository struct {
	pool *pgxpool.Pool
}

func NewUpstreamRepository(pool *pgxpool.Pool) *UpstreamRepository {
	return &UpstreamRepository{pool: pool}
}

func (r *UpstreamRepository) Create(ctx context.Context, u *model.Upstream) error {
	query := `
		INSERT INTO upstreams (id, name, host, port, protocol, tls_verify, health_check_path, health_check_interval_secs, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query, u.ID, u.Name, u.Host, u.Port, u.Protocol, u.TlsVerify, u.HealthCheckPath, u.HealthCheckIntervalS, u.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert upstream: %w", err)
	}
	return nil
}

func (r *UpstreamRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Upstream, error) {
	query := `SELECT id, name, host, port, protocol, tls_verify, health_check_path, health_check_interval_secs, is_healthy, created_at, updated_at FROM upstreams WHERE id = $1`
	u := &model.Upstream{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Name, &u.Host, &u.Port, &u.Protocol, &u.TlsVerify, &u.HealthCheckPath, &u.HealthCheckIntervalS, &u.IsHealthy, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("upstream not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get upstream: %w", err)
	}
	return u, nil
}

func (r *UpstreamRepository) List(ctx context.Context, offset, limit int) ([]model.Upstream, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM upstreams`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count upstreams: %w", err)
	}

	query := `SELECT id, name, host, port, protocol, tls_verify, health_check_path, health_check_interval_secs, is_healthy, created_at, updated_at FROM upstreams ORDER BY created_at DESC OFFSET $1 LIMIT $2`
	rows, err := r.pool.Query(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list upstreams: %w", err)
	}
	defer rows.Close()

	var upstreams []model.Upstream
	for rows.Next() {
		var u model.Upstream
		if err := rows.Scan(&u.ID, &u.Name, &u.Host, &u.Port, &u.Protocol, &u.TlsVerify, &u.HealthCheckPath, &u.HealthCheckIntervalS, &u.IsHealthy, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan upstream: %w", err)
		}
		upstreams = append(upstreams, u)
	}
	return upstreams, total, rows.Err()
}

func (r *UpstreamRepository) UpdateHealth(ctx context.Context, id uuid.UUID, healthy bool) error {
	query := `UPDATE upstreams SET is_healthy = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, healthy, id)
	return err
}
