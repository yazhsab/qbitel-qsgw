package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/quantun-opensource/qsgw/control-plane/internal/model"
	"github.com/quantun-opensource/qsgw/control-plane/internal/service"
	qmw "github.com/quantun-opensource/qsgw/shared/go/middleware"
)

// maxNameLength is the maximum allowed length for name fields.
const maxNameLength = 255

type GatewayHandler struct {
	svc    *service.GatewayService
	logger *zap.Logger
}

func NewGatewayHandler(svc *service.GatewayService, logger *zap.Logger) *GatewayHandler {
	return &GatewayHandler{svc: svc, logger: logger}
}

func (h *GatewayHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Post("/{id}/activate", h.Activate)
	r.Post("/{id}/deactivate", h.Deactivate)
	return r
}

func (h *GatewayHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateGatewayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Hostname == "" {
		writeError(w, http.StatusBadRequest, "name and hostname are required")
		return
	}
	if len(req.Name) > maxNameLength || len(req.Hostname) > maxNameLength {
		writeError(w, http.StatusBadRequest, "name or hostname exceeds maximum length")
		return
	}
	if req.CreatedBy == "" {
		req.CreatedBy = actorFromRequest(r)
	}

	gw, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create gateway", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to create gateway")
		return
	}
	writeJSON(w, http.StatusCreated, gw.ToResponse())
}

func (h *GatewayHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid gateway ID")
		return
	}
	gw, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "gateway not found")
		return
	}
	writeJSON(w, http.StatusOK, gw.ToResponse())
}

func (h *GatewayHandler) List(w http.ResponseWriter, r *http.Request) {
	pg := qmw.ParsePagination(r)

	gateways, total, err := h.svc.List(r.Context(), pg.Offset, pg.Limit)
	if err != nil {
		h.logger.Error("failed to list gateways", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list gateways")
		return
	}

	resp := model.GatewayListResponse{TotalCount: total, Offset: pg.Offset, Limit: pg.Limit}
	for _, g := range gateways {
		resp.Gateways = append(resp.Gateways, g.ToResponse())
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *GatewayHandler) Activate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid gateway ID")
		return
	}
	if err := h.svc.Activate(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to activate gateway")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "activated"})
}

func (h *GatewayHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid gateway ID")
		return
	}
	if err := h.svc.Deactivate(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to deactivate gateway")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deactivated"})
}
