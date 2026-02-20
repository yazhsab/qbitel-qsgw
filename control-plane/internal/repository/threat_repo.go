package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/quantun-opensource/qsgw/control-plane/internal/model"
)

type ThreatRepository struct {
	pool *pgxpool.Pool
}

func NewThreatRepository(pool *pgxpool.Pool) *ThreatRepository {
	return &ThreatRepository{pool: pool}
}

func (r *ThreatRepository) Create(ctx context.Context, t *model.ThreatEvent) error {
	query := `
		INSERT INTO threat_events (id, gateway_id, threat_type, severity, source_ip, description, mitigated, detected_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query, t.ID, t.GatewayID, t.ThreatType, t.Severity, t.SourceIP, t.Description, t.Mitigated, t.DetectedAt)
	if err != nil {
		return fmt.Errorf("failed to insert threat event: %w", err)
	}
	return nil
}

func (r *ThreatRepository) ListByGateway(ctx context.Context, gatewayID uuid.UUID, severity string, offset, limit int) ([]model.ThreatEvent, int, error) {
	countQuery := `SELECT COUNT(*) FROM threat_events WHERE gateway_id = $1`
	listQuery := `SELECT id, gateway_id, threat_type, severity, source_ip, description, mitigated, detected_at FROM threat_events WHERE gateway_id = $1`
	args := []interface{}{gatewayID}
	argIdx := 2

	if severity != "" {
		filter := fmt.Sprintf(" AND severity = $%d", argIdx)
		countQuery += filter
		listQuery += filter
		args = append(args, severity)
		argIdx++
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count threats: %w", err)
	}

	listQuery += fmt.Sprintf(" ORDER BY detected_at DESC OFFSET $%d LIMIT $%d", argIdx, argIdx+1)
	args = append(args, offset, limit)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list threats: %w", err)
	}
	defer rows.Close()

	var events []model.ThreatEvent
	for rows.Next() {
		var t model.ThreatEvent
		if err := rows.Scan(&t.ID, &t.GatewayID, &t.ThreatType, &t.Severity, &t.SourceIP, &t.Description, &t.Mitigated, &t.DetectedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan threat: %w", err)
		}
		events = append(events, t)
	}
	return events, total, rows.Err()
}
