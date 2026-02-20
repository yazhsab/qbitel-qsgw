package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/quantun-opensource/qsgw/control-plane/internal/model"
	"github.com/quantun-opensource/qsgw/control-plane/internal/repository"
)

type RouteService struct {
	repo   *repository.RouteRepository
	logger *zap.Logger
}

func NewRouteService(repo *repository.RouteRepository, logger *zap.Logger) *RouteService {
	return &RouteService{repo: repo, logger: logger}
}

func (s *RouteService) Create(ctx context.Context, req *model.CreateRouteRequest) (*model.Route, error) {
	gatewayID, err := uuid.Parse(req.GatewayID)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway_id: %w", err)
	}
	upstreamID, err := uuid.Parse(req.UpstreamID)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream_id: %w", err)
	}

	route := &model.Route{
		ID:          uuid.New(),
		GatewayID:   gatewayID,
		UpstreamID:  upstreamID,
		PathPrefix:  req.PathPrefix,
		StripPrefix: req.StripPrefix,
		Priority:    req.Priority,
		Enabled:     true,
		CreatedAt:   time.Now().UTC(),
	}

	if req.TlsPolicy != "" {
		route.TlsPolicy = &req.TlsPolicy
	}
	if req.RateLimitRPS > 0 {
		route.RateLimitRPS = &req.RateLimitRPS
	}

	if err := s.repo.Create(ctx, route); err != nil {
		s.logger.Error("failed to create route", zap.Error(err))
		return nil, err
	}

	s.logger.Info("route created", zap.String("id", route.ID.String()), zap.String("prefix", route.PathPrefix))
	return route, nil
}

func (s *RouteService) ListByGateway(ctx context.Context, gatewayID uuid.UUID) ([]model.Route, error) {
	return s.repo.ListByGateway(ctx, gatewayID)
}

func (s *RouteService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
