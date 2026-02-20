package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/quantun-opensource/qsgw/control-plane/internal/model"
	"github.com/quantun-opensource/qsgw/control-plane/internal/service"
)

type UpstreamHandler struct {
	svc    *service.UpstreamService
	logger *zap.Logger
}

func NewUpstreamHandler(svc *service.UpstreamService, logger *zap.Logger) *UpstreamHandler {
	return &UpstreamHandler{svc: svc, logger: logger}
}

func (h *UpstreamHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	return r
}

func (h *UpstreamHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUpstreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Host == "" || req.Port == 0 {
		writeError(w, http.StatusBadRequest, "name, host, and port are required")
		return
	}

	u, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create upstream", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to create upstream")
		return
	}
	writeJSON(w, http.StatusCreated, u.ToResponse())
}

func (h *UpstreamHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid upstream ID")
		return
	}
	u, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "upstream not found")
		return
	}
	writeJSON(w, http.StatusOK, u.ToResponse())
}

func (h *UpstreamHandler) List(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	upstreams, total, err := h.svc.List(r.Context(), offset, limit)
	if err != nil {
		h.logger.Error("failed to list upstreams", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "failed to list upstreams")
		return
	}

	resp := model.UpstreamListResponse{TotalCount: total}
	for _, u := range upstreams {
		resp.Upstreams = append(resp.Upstreams, u.ToResponse())
	}
	writeJSON(w, http.StatusOK, resp)
}
