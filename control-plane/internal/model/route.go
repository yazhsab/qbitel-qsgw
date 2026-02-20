package model

import (
	"time"

	"github.com/google/uuid"
)

type Route struct {
	ID           uuid.UUID `json:"id"`
	GatewayID    uuid.UUID `json:"gateway_id"`
	UpstreamID   uuid.UUID `json:"upstream_id"`
	PathPrefix   string    `json:"path_prefix"`
	StripPrefix  bool      `json:"strip_prefix"`
	Priority     int       `json:"priority"`
	TlsPolicy   *string   `json:"tls_policy"`
	RateLimitRPS *int      `json:"rate_limit_rps"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateRouteRequest struct {
	GatewayID    string `json:"gateway_id"`
	UpstreamID   string `json:"upstream_id"`
	PathPrefix   string `json:"path_prefix"`
	StripPrefix  bool   `json:"strip_prefix"`
	Priority     int    `json:"priority"`
	TlsPolicy   string `json:"tls_policy,omitempty"`
	RateLimitRPS int    `json:"rate_limit_rps,omitempty"`
}

type RouteResponse struct {
	ID           uuid.UUID `json:"id"`
	GatewayID    uuid.UUID `json:"gateway_id"`
	UpstreamID   uuid.UUID `json:"upstream_id"`
	PathPrefix   string    `json:"path_prefix"`
	StripPrefix  bool      `json:"strip_prefix"`
	Priority     int       `json:"priority"`
	TlsPolicy   *string   `json:"tls_policy,omitempty"`
	RateLimitRPS *int      `json:"rate_limit_rps,omitempty"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    string    `json:"created_at"`
}

type RouteListResponse struct {
	Routes     []RouteResponse `json:"routes"`
	TotalCount int             `json:"total_count"`
}

func (r *Route) ToResponse() RouteResponse {
	return RouteResponse{
		ID:           r.ID,
		GatewayID:    r.GatewayID,
		UpstreamID:   r.UpstreamID,
		PathPrefix:   r.PathPrefix,
		StripPrefix:  r.StripPrefix,
		Priority:     r.Priority,
		TlsPolicy:   r.TlsPolicy,
		RateLimitRPS: r.RateLimitRPS,
		Enabled:      r.Enabled,
		CreatedAt:    r.CreatedAt.Format(time.RFC3339),
	}
}
