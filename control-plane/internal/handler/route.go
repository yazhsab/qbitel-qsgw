package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/quantun-opensource/qsgw/control-plane/internal/model"
	"github.com/quantun-opensource/qsgw/control-plane/internal/service"
)

type RouteHandler struct {
	svc    *service.RouteService
	logger *zap.Logger
}

func NewRouteHandler(svc *service.RouteService, logger *zap.Logger) *RouteHandler {
	return &RouteHandler{svc: svc, logger: logger}
}

func (h *RouteHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *RouteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.GatewayID == "" || req.UpstreamID == "" || req.PathPrefix == "" {
		writeError(w, http.StatusBadRequest, "gateway_id, upstream_id, and path_prefix are required")
		return
	}

	route, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create route", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to create route")
		return
	}
	writeJSON(w, http.StatusCreated, route.ToResponse())
}

func (h *RouteHandler) List(w http.ResponseWriter, r *http.Request) {
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

	routes, err := h.svc.ListByGateway(r.Context(), gatewayID)
	if err != nil {
		h.logger.Error("failed to list routes", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list routes")
		return
	}

	resp := model.RouteListResponse{TotalCount: len(routes)}
	for _, rt := range routes {
		resp.Routes = append(resp.Routes, rt.ToResponse())
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *RouteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid route ID")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete route")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
