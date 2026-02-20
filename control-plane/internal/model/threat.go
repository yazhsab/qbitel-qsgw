package model

import (
	"time"

	"github.com/google/uuid"
)

type ThreatEvent struct {
	ID         uuid.UUID `json:"id"`
	GatewayID  uuid.UUID `json:"gateway_id"`
	ThreatType string    `json:"threat_type"`
	Severity   string    `json:"severity"`
	SourceIP   *string   `json:"source_ip"`
	Description string   `json:"description"`
	Mitigated  bool      `json:"mitigated"`
	DetectedAt time.Time `json:"detected_at"`
}

type ThreatEventResponse struct {
	ID          uuid.UUID `json:"id"`
	GatewayID   uuid.UUID `json:"gateway_id"`
	ThreatType  string    `json:"threat_type"`
	Severity    string    `json:"severity"`
	SourceIP    *string   `json:"source_ip,omitempty"`
	Description string    `json:"description"`
	Mitigated   bool      `json:"mitigated"`
	DetectedAt  string    `json:"detected_at"`
}

type ThreatEventListResponse struct {
	Events     []ThreatEventResponse `json:"events"`
	TotalCount int                   `json:"total_count"`
	Offset     int                   `json:"offset"`
	Limit      int                   `json:"limit"`
}

func (t *ThreatEvent) ToResponse() ThreatEventResponse {
	return ThreatEventResponse{
		ID:          t.ID,
		GatewayID:   t.GatewayID,
		ThreatType:  t.ThreatType,
		Severity:    t.Severity,
		SourceIP:    t.SourceIP,
		Description: t.Description,
		Mitigated:   t.Mitigated,
		DetectedAt:  t.DetectedAt.Format(time.RFC3339),
	}
}
