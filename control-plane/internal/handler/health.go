package handler

import (
	"encoding/json"
	"net/http"

	qmw "github.com/quantun-opensource/qsgw/shared/go/middleware"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "qsgw-control-plane",
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// actorFromRequest extracts the authenticated subject from the request context,
// falling back to "system" if no auth context is present.
func actorFromRequest(r *http.Request) string {
	if subject := qmw.SubjectFromContext(r.Context()); subject != "" {
		return subject
	}
	return "system"
}
