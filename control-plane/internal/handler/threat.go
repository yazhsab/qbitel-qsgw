package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/quantun-opensource/qsgw/control-plane/internal/model"
	"github.com/quantun-opensource/qsgw/control-plane/internal/repository"
)

type ThreatHandler struct {
	repo   *repository.ThreatRepository
	logger *zap.Logger
}

func NewThreatHandler(repo *repository.ThreatRepository, logger *zap.Logger) *ThreatHandler {
	return &ThreatHandler{repo: repo, logger: logger}
}

func (h *ThreatHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	return r
}

func (h *ThreatHandler) List(w http.ResponseWriter, r *http.Request) {
	gatewayIDStr := r.URL.Query().Get("gateway_id")
	if gatewayIDStr == "" {
		writeError(w, http.StatusBadRequest, "gateway_id is required")
		return
	}
	gatewayID, err := uuid.Parse(gatewayIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid gateway_id")
		return
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	severity := r.URL.Query().Get("severity")

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	events, total, err := h.repo.ListByGateway(r.Context(), gatewayID, severity, offset, limit)
	if err != nil {
		h.logger.Error("failed to list threats", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list threats")
		return
	}

	resp := model.ThreatEventListResponse{TotalCount: total, Offset: offset, Limit: limit}
	for _, e := range events {
		resp.Events = append(resp.Events, e.ToResponse())
	}
	writeJSON(w, http.StatusOK, resp)
}
