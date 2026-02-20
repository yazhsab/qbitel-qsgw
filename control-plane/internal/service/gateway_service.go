package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/quantun-opensource/qsgw/control-plane/internal/model"
	"github.com/quantun-opensource/qsgw/control-plane/internal/repository"
)

type GatewayService struct {
	repo   *repository.GatewayRepository
	logger *zap.Logger
}

func NewGatewayService(repo *repository.GatewayRepository, logger *zap.Logger) *GatewayService {
	return &GatewayService{repo: repo, logger: logger}
}

func (s *GatewayService) Create(ctx context.Context, req *model.CreateGatewayRequest) (*model.Gateway, error) {
	if req.Port == 0 {
		req.Port = 443
	}
	if req.TlsPolicy == "" {
		req.TlsPolicy = "PQC_PREFERRED"
	}
	if req.MaxConnections == 0 {
		req.MaxConnections = 10000
	}

	g := &model.Gateway{
		ID:             uuid.New(),
		Name:           req.Name,
		Hostname:       req.Hostname,
		Port:           req.Port,
		Status:         "INACTIVE",
		TlsPolicy:     req.TlsPolicy,
		MaxConnections: req.MaxConnections,
		CreatedBy:      req.CreatedBy,
		CreatedAt:      time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, g); err != nil {
		s.logger.Error("failed to create gateway", zap.Error(err))
		return nil, err
	}

	s.logger.Info("gateway created", zap.String("id", g.ID.String()), zap.String("name", g.Name))
	return g, nil
}

func (s *GatewayService) Get(ctx context.Context, id uuid.UUID) (*model.Gateway, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *GatewayService) List(ctx context.Context, offset, limit int) ([]model.Gateway, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.List(ctx, offset, limit)
}

func (s *GatewayService) Activate(ctx context.Context, id uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, id, "ACTIVE")
}

func (s *GatewayService) Deactivate(ctx context.Context, id uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, id, "INACTIVE")
}
