package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quantun-opensource/qsgw/control-plane/internal/model"
)

type GatewayRepository struct {
	pool *pgxpool.Pool
}

func NewGatewayRepository(pool *pgxpool.Pool) *GatewayRepository {
	return &GatewayRepository{pool: pool}
}

func (r *GatewayRepository) Create(ctx context.Context, g *model.Gateway) error {
	query := `
		INSERT INTO gateways (id, name, hostname, port, status, tls_policy, max_connections, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query, g.ID, g.Name, g.Hostname, g.Port, g.Status, g.TlsPolicy, g.MaxConnections, g.CreatedBy, g.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert gateway: %w", err)
	}
	return nil
}

func (r *GatewayRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Gateway, error) {
	query := `SELECT id, name, hostname, port, status, tls_policy, max_connections, created_by, created_at, updated_at FROM gateways WHERE id = $1`
	g := &model.Gateway{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&g.ID, &g.Name, &g.Hostname, &g.Port, &g.Status, &g.TlsPolicy, &g.MaxConnections, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("gateway not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get gateway: %w", err)
	}
	return g, nil
}

func (r *GatewayRepository) List(ctx context.Context, offset, limit int) ([]model.Gateway, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM gateways`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count gateways: %w", err)
	}

	query := `SELECT id, name, hostname, port, status, tls_policy, max_connections, created_by, created_at, updated_at FROM gateways ORDER BY created_at DESC OFFSET $1 LIMIT $2`
	rows, err := r.pool.Query(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list gateways: %w", err)
	}
	defer rows.Close()

	var gateways []model.Gateway
	for rows.Next() {
		var g model.Gateway
		if err := rows.Scan(&g.ID, &g.Name, &g.Hostname, &g.Port, &g.Status, &g.TlsPolicy, &g.MaxConnections, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan gateway: %w", err)
		}
		gateways = append(gateways, g)
	}
	return gateways, total, rows.Err()
}

func (r *GatewayRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE gateways SET status = $1, updated_at = $2 WHERE id = $3`
	result, err := r.pool.Exec(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update gateway status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("gateway not found: %s", id)
	}
	return nil
}
