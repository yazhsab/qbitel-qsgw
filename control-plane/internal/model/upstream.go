package model

import (
	"time"

	"github.com/google/uuid"
)

type Upstream struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	Host                  string    `json:"host"`
	Port                  int       `json:"port"`
	Protocol              string    `json:"protocol"`
	TlsVerify             bool      `json:"tls_verify"`
	HealthCheckPath       string    `json:"health_check_path"`
	HealthCheckIntervalS  int       `json:"health_check_interval_secs"`
	IsHealthy             bool      `json:"is_healthy"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type CreateUpstreamRequest struct {
	Name                 string `json:"name"`
	Host                 string `json:"host"`
	Port                 int    `json:"port"`
	Protocol             string `json:"protocol"`
	TlsVerify            bool   `json:"tls_verify"`
	HealthCheckPath      string `json:"health_check_path"`
	HealthCheckIntervalS int    `json:"health_check_interval_secs"`
}

type UpstreamResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Host            string    `json:"host"`
	Port            int       `json:"port"`
	Protocol        string    `json:"protocol"`
	TlsVerify       bool      `json:"tls_verify"`
	HealthCheckPath string    `json:"health_check_path"`
	IsHealthy       bool      `json:"is_healthy"`
	CreatedAt       string    `json:"created_at"`
}

type UpstreamListResponse struct {
	Upstreams  []UpstreamResponse `json:"upstreams"`
	TotalCount int                `json:"total_count"`
}

func (u *Upstream) ToResponse() UpstreamResponse {
	return UpstreamResponse{
		ID:              u.ID,
		Name:            u.Name,
		Host:            u.Host,
		Port:            u.Port,
		Protocol:        u.Protocol,
		TlsVerify:       u.TlsVerify,
		HealthCheckPath: u.HealthCheckPath,
		IsHealthy:       u.IsHealthy,
		CreatedAt:       u.CreatedAt.Format(time.RFC3339),
	}
}
