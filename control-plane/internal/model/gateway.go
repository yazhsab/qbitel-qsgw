package model

import (
	"time"

	"github.com/google/uuid"
)

type Gateway struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Hostname       string    `json:"hostname"`
	Port           int       `json:"port"`
	Status         string    `json:"status"`
	TlsPolicy     string    `json:"tls_policy"`
	MaxConnections int       `json:"max_connections"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateGatewayRequest struct {
	Name           string `json:"name"`
	Hostname       string `json:"hostname"`
	Port           int    `json:"port"`
	TlsPolicy     string `json:"tls_policy"`
	MaxConnections int    `json:"max_connections"`
	CreatedBy      string `json:"created_by"`
}

type GatewayResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Hostname       string    `json:"hostname"`
	Port           int       `json:"port"`
	Status         string    `json:"status"`
	TlsPolicy     string    `json:"tls_policy"`
	MaxConnections int       `json:"max_connections"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}

type GatewayListResponse struct {
	Gateways   []GatewayResponse `json:"gateways"`
	TotalCount int               `json:"total_count"`
	Offset     int               `json:"offset"`
	Limit      int               `json:"limit"`
}

func (g *Gateway) ToResponse() GatewayResponse {
	return GatewayResponse{
		ID:             g.ID,
		Name:           g.Name,
		Hostname:       g.Hostname,
		Port:           g.Port,
		Status:         g.Status,
		TlsPolicy:     g.TlsPolicy,
		MaxConnections: g.MaxConnections,
		CreatedAt:      g.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      g.UpdatedAt.Format(time.RFC3339),
	}
}
