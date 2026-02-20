package middleware

import (
	"net/http"
	"strconv"
)

const (
	// DefaultPageLimit is the default number of items per page.
	DefaultPageLimit = 20
	// MaxPageLimit is the maximum number of items per page.
	MaxPageLimit = 100
	// MaxOffset is the maximum allowed offset to prevent abuse.
	MaxOffset = 1_000_000
	// DefaultOffset is the default starting offset.
	DefaultOffset = 0
)

// Pagination holds validated pagination parameters.
type Pagination struct {
	Offset int
	Limit  int
}

// ParsePagination extracts and validates offset/limit query parameters.
//
// Rules:
//   - offset defaults to 0, must be >= 0, capped at MaxOffset (1,000,000)
//   - limit defaults to DefaultPageLimit (20), clamped to [1, MaxPageLimit (100)]
func ParsePagination(r *http.Request) Pagination {
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = DefaultOffset
	}
	if offset > MaxOffset {
		offset = MaxOffset
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}

	return Pagination{Offset: offset, Limit: limit}
}
