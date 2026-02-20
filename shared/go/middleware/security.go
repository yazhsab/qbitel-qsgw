package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// SecurityHeadersConfig configures security response headers.
type SecurityHeadersConfig struct {
	// ContentSecurityPolicy sets the CSP header. Defaults to "default-src 'none'".
	ContentSecurityPolicy string
	// HSTSMaxAge sets Strict-Transport-Security max-age in seconds. 0 disables. Default: 63072000 (2 years).
	HSTSMaxAge int
	// FrameOptions sets X-Frame-Options. Defaults to "DENY".
	FrameOptions string
}

// DefaultSecurityHeadersConfig returns a production-ready security headers config.
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		ContentSecurityPolicy: "default-src 'none'",
		HSTSMaxAge:            63072000, // 2 years
		FrameOptions:          "DENY",
	}
}

// SecurityHeaders returns middleware that sets standard security response headers.
//
// Headers set:
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options: DENY (configurable)
//   - X-XSS-Protection: 0 (modern browsers should use CSP)
//   - Content-Security-Policy: default-src 'none' (configurable)
//   - Strict-Transport-Security: max-age=63072000; includeSubDomains (configurable)
//   - Referrer-Policy: strict-origin-when-cross-origin
//   - Permissions-Policy: geolocation=(), camera=(), microphone=()
//   - Cache-Control: no-store (for API responses)
func SecurityHeaders(cfg SecurityHeadersConfig) func(http.Handler) http.Handler {
	if cfg.ContentSecurityPolicy == "" {
		cfg.ContentSecurityPolicy = "default-src 'none'"
	}
	if cfg.FrameOptions == "" {
		cfg.FrameOptions = "DENY"
	}

	hstsValue := ""
	if cfg.HSTSMaxAge > 0 {
		hstsValue = "max-age=" + strconv.Itoa(cfg.HSTSMaxAge) + "; includeSubDomains"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", cfg.FrameOptions)
			h.Set("X-XSS-Protection", "0")
			h.Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
			h.Set("Cache-Control", "no-store")

			if hstsValue != "" {
				h.Set("Strict-Transport-Security", hstsValue)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSConfig configures Cross-Origin Resource Sharing.
type CORSConfig struct {
	// AllowedOrigins is the list of allowed origins. Use ["*"] for permissive (dev only).
	AllowedOrigins []string
	// AllowedMethods is the list of allowed HTTP methods.
	AllowedMethods []string
	// AllowedHeaders is the list of allowed request headers.
	AllowedHeaders []string
	// ExposedHeaders is the list of headers exposed to the browser.
	ExposedHeaders []string
	// AllowCredentials indicates whether cookies/auth headers are allowed.
	AllowCredentials bool
	// MaxAge is the max time (seconds) for preflight cache. Default: 86400 (24h).
	MaxAge int
}

// DefaultCORSConfig returns a restrictive CORS config suitable for production APIs.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Request-ID",
		},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// CORS returns middleware that handles CORS preflight and sets CORS headers.
func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	if len(cfg.AllowedMethods) == 0 {
		cfg.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(cfg.AllowedHeaders) == 0 {
		cfg.AllowedHeaders = []string{"Accept", "Authorization", "Content-Type"}
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = 86400
	}

	// SECURITY: Prevent wildcard + credentials misconfiguration.
	// The CORS spec forbids Access-Control-Allow-Origin: * with credentials.
	// This combination would allow any origin to make authenticated requests.
	originSet := make(map[string]bool, len(cfg.AllowedOrigins))
	wildcard := false
	for _, o := range cfg.AllowedOrigins {
		if o == "*" {
			wildcard = true
		}
		originSet[o] = true
	}
	if wildcard && cfg.AllowCredentials {
		// Silently disable credentials when wildcard is used — this prevents
		// a dangerous misconfiguration that would allow credential-bearing
		// requests from any origin.
		cfg.AllowCredentials = false
	}

	methodsStr := strings.Join(cfg.AllowedMethods, ", ")
	headersStr := strings.Join(cfg.AllowedHeaders, ", ")
	exposedStr := strings.Join(cfg.ExposedHeaders, ", ")
	maxAgeStr := strconv.Itoa(cfg.MaxAge)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// No Origin header — not a CORS request, proceed normally
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check if origin is allowed
			allowed := wildcard || originSet[origin]
			if !allowed {
				// Origin not allowed — proceed without CORS headers
				next.ServeHTTP(w, r)
				return
			}

			h := w.Header()

			if wildcard {
				h.Set("Access-Control-Allow-Origin", "*")
			} else {
				h.Set("Access-Control-Allow-Origin", origin)
				h.Add("Vary", "Origin")
			}

			if cfg.AllowCredentials {
				h.Set("Access-Control-Allow-Credentials", "true")
			}

			if exposedStr != "" {
				h.Set("Access-Control-Expose-Headers", exposedStr)
			}

			// Handle preflight
			if r.Method == http.MethodOptions {
				h.Set("Access-Control-Allow-Methods", methodsStr)
				h.Set("Access-Control-Allow-Headers", headersStr)
				h.Set("Access-Control-Max-Age", maxAgeStr)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// MaxBodySize returns middleware that limits the size of request bodies.
//
// If the Content-Length header exceeds maxBytes, it returns 413 Request Entity Too Large.
// It also wraps the body with http.MaxBytesReader to enforce the limit during reading.
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				writeAuthError(w, http.StatusRequestEntityTooLarge,
					"request body too large (max "+strconv.FormatInt(maxBytes, 10)+" bytes)")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
