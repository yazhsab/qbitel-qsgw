package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/quantun-opensource/qsgw/control-plane/internal/model"
	"github.com/quantun-opensource/qsgw/control-plane/internal/repository"
)

type UpstreamService struct {
	repo   *repository.UpstreamRepository
	logger *zap.Logger
}

func NewUpstreamService(repo *repository.UpstreamRepository, logger *zap.Logger) *UpstreamService {
	return &UpstreamService{repo: repo, logger: logger}
}

func (s *UpstreamService) Create(ctx context.Context, req *model.CreateUpstreamRequest) (*model.Upstream, error) {
	if req.Protocol == "" {
		req.Protocol = "HTTPS"
	}
	if req.HealthCheckPath == "" {
		req.HealthCheckPath = "/health"
	}
	if req.HealthCheckIntervalS == 0 {
		req.HealthCheckIntervalS = 30
	}

	u := &model.Upstream{
		ID:                   uuid.New(),
		Name:                 req.Name,
		Host:                 req.Host,
		Port:                 req.Port,
		Protocol:             req.Protocol,
		TlsVerify:            req.TlsVerify,
		HealthCheckPath:      req.HealthCheckPath,
		HealthCheckIntervalS: req.HealthCheckIntervalS,
		IsHealthy:            true,
		CreatedAt:            time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, u); err != nil {
		s.logger.Error("failed to create upstream", zap.Error(err))
		return nil, err
	}

	s.logger.Info("upstream created", zap.String("id", u.ID.String()), zap.String("name", u.Name))
	return u, nil
}

func (s *UpstreamService) Get(ctx context.Context, id uuid.UUID) (*model.Upstream, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UpstreamService) List(ctx context.Context, offset, limit int) ([]model.Upstream, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.List(ctx, offset, limit)
}
