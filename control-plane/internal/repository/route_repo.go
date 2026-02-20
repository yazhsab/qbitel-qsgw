package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quantun-opensource/qsgw/control-plane/internal/model"
)

type RouteRepository struct {
	pool *pgxpool.Pool
}

func NewRouteRepository(pool *pgxpool.Pool) *RouteRepository {
	return &RouteRepository{pool: pool}
}

func (r *RouteRepository) Create(ctx context.Context, route *model.Route) error {
	query := `
		INSERT INTO routes (id, gateway_id, upstream_id, path_prefix, strip_prefix, priority, tls_policy, rate_limit_rps, enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		route.ID, route.GatewayID, route.UpstreamID, route.PathPrefix,
		route.StripPrefix, route.Priority, route.TlsPolicy, route.RateLimitRPS, route.Enabled, route.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert route: %w", err)
	}
	return nil
}

func (r *RouteRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Route, error) {
	query := `SELECT id, gateway_id, upstream_id, path_prefix, strip_prefix, priority, tls_policy, rate_limit_rps, enabled, created_at, updated_at FROM routes WHERE id = $1`
	route := &model.Route{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&route.ID, &route.GatewayID, &route.UpstreamID, &route.PathPrefix,
		&route.StripPrefix, &route.Priority, &route.TlsPolicy, &route.RateLimitRPS, &route.Enabled, &route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("route not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get route: %w", err)
	}
	return route, nil
}

func (r *RouteRepository) ListByGateway(ctx context.Context, gatewayID uuid.UUID) ([]model.Route, error) {
	query := `SELECT id, gateway_id, upstream_id, path_prefix, strip_prefix, priority, tls_policy, rate_limit_rps, enabled, created_at, updated_at FROM routes WHERE gateway_id = $1 ORDER BY priority DESC`
	rows, err := r.pool.Query(ctx, query, gatewayID)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes: %w", err)
	}
	defer rows.Close()

	var routes []model.Route
	for rows.Next() {
		var route model.Route
		if err := rows.Scan(
			&route.ID, &route.GatewayID, &route.UpstreamID, &route.PathPrefix,
			&route.StripPrefix, &route.Priority, &route.TlsPolicy, &route.RateLimitRPS, &route.Enabled, &route.CreatedAt, &route.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		routes = append(routes, route)
	}
	return routes, rows.Err()
}

func (r *RouteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM routes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("route not found: %s", id)
	}
	return nil
}
