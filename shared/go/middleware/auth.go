// Package middleware provides shared HTTP middleware for QSGW services.
//
// Authentication supports two modes:
//   - JWT Bearer tokens: validated using HMAC-SHA256 with a configurable secret
//   - API keys: validated against a configurable set of valid keys
//
// Both modes extract claims/identity and inject them into the request context.
package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// contextKey is an unexported type used for context keys to avoid collisions.
type contextKey string

const (
	// ContextKeySubject holds the authenticated subject (user/service ID).
	ContextKeySubject contextKey = "auth.subject"
	// ContextKeyRole holds the authenticated role.
	ContextKeyRole contextKey = "auth.role"
	// ContextKeyAuthMethod holds how the request was authenticated.
	ContextKeyAuthMethod contextKey = "auth.method"
)

// AuthMethod describes how a request was authenticated.
type AuthMethod string

const (
	AuthMethodJWT    AuthMethod = "jwt"
	AuthMethodAPIKey AuthMethod = "api_key"
)

// JWTClaims represents the payload of a JWT token.
type JWTClaims struct {
	Subject   string `json:"sub"`
	Role      string `json:"role"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	NotBefore int64  `json:"nbf,omitempty"`
	Issuer    string `json:"iss"`
	JTI       string `json:"jti,omitempty"`
}

// APIKeyEntry maps an API key to an identity.
type APIKeyEntry struct {
	Key     string
	Subject string
	Role    string
}

// AuthConfig configures the authentication middleware.
type AuthConfig struct {
	// JWTSecret is the HMAC-SHA256 secret for JWT validation.
	// If empty, JWT auth is disabled.
	JWTSecret string

	// JWTIssuer is the expected "iss" claim. If empty, issuer is not checked.
	JWTIssuer string

	// APIKeys is a list of valid API keys. If empty, API key auth is disabled.
	APIKeys []APIKeyEntry

	// SkipPaths are URL paths that bypass authentication (e.g., /health).
	SkipPaths []string

	// Logger for auth events. If nil, a no-op logger is used.
	Logger *zap.Logger
}

// Auth returns a Chi-compatible middleware that enforces authentication.
//
// It checks the Authorization header for:
//   - "Bearer <jwt>" -- validates the JWT using HMAC-SHA256
//   - "ApiKey <key>" -- validates against the configured API key list
//
// On success, it injects subject, role, and auth method into the context.
// On failure, it returns 401 Unauthorized.
func Auth(cfg AuthConfig) func(http.Handler) http.Handler {
	logger := cfg.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	skipSet := make(map[string]bool, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		skipSet[p] = true
	}

	// Build list of keys for constant-time iteration (no map lookup timing leak).
	apiKeyList := make([]APIKeyEntry, len(cfg.APIKeys))
	copy(apiKeyList, cfg.APIKeys)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for configured paths
			if skipSet[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeAuthError(w, http.StatusUnauthorized, "missing Authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 {
				writeAuthError(w, http.StatusUnauthorized, "malformed Authorization header")
				return
			}

			scheme := parts[0]
			credential := parts[1]

			var (
				subject string
				role    string
				method  AuthMethod
			)

			switch strings.ToLower(scheme) {
			case "bearer":
				if cfg.JWTSecret == "" {
					writeAuthError(w, http.StatusUnauthorized, "JWT authentication not configured")
					return
				}
				claims, err := validateJWT(credential, cfg.JWTSecret, cfg.JWTIssuer)
				if err != nil {
					logger.Debug("JWT validation failed",
						zap.Error(err),
						zap.String("remote_addr", r.RemoteAddr),
					)
					writeAuthError(w, http.StatusUnauthorized, "invalid or expired token")
					return
				}
				subject = claims.Subject
				role = claims.Role
				method = AuthMethodJWT

			case "apikey":
				if len(apiKeyList) == 0 {
					writeAuthError(w, http.StatusUnauthorized, "API key authentication not configured")
					return
				}
				entry, ok := validateAPIKey(credential, apiKeyList)
				if !ok {
					logger.Debug("API key validation failed",
						zap.String("remote_addr", r.RemoteAddr),
					)
					writeAuthError(w, http.StatusUnauthorized, "invalid API key")
					return
				}
				subject = entry.Subject
				role = entry.Role
				method = AuthMethodAPIKey

			default:
				writeAuthError(w, http.StatusUnauthorized, fmt.Sprintf("unsupported auth scheme: %s", scheme))
				return
			}

			// Inject identity into request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, ContextKeySubject, subject)
			ctx = context.WithValue(ctx, ContextKeyRole, role)
			ctx = context.WithValue(ctx, ContextKeyAuthMethod, method)

			logger.Debug("authenticated request",
				zap.String("subject", subject),
				zap.String("role", role),
				zap.String("method", string(method)),
				zap.String("path", r.URL.Path),
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns middleware that enforces the request was authenticated
// with one of the specified roles. Must be used after Auth middleware.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(ContextKeyRole).(string)
			if !roleSet[role] {
				writeAuthError(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SubjectFromContext extracts the authenticated subject from a request context.
func SubjectFromContext(ctx context.Context) string {
	s, _ := ctx.Value(ContextKeySubject).(string)
	return s
}

// RoleFromContext extracts the authenticated role from a request context.
func RoleFromContext(ctx context.Context) string {
	s, _ := ctx.Value(ContextKeyRole).(string)
	return s
}

// AuthMethodFromContext extracts the authentication method from a request context.
func AuthMethodFromContext(ctx context.Context) AuthMethod {
	m, _ := ctx.Value(ContextKeyAuthMethod).(AuthMethod)
	return m
}

// --- JWT helpers (minimal, dependency-free HMAC-SHA256) ---

// validateJWT parses and validates a compact JWT (header.payload.signature).
// Only HS256 (HMAC-SHA256) is supported. The "none" algorithm is explicitly rejected.
//
// Validates: signature, algorithm, expiration (exp), not-before (nbf), and issuer (iss).
func validateJWT(tokenString, secret, expectedIssuer string) (*JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Verify header declares HS256
	headerJSON, err := base64URLDecode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid JWT header encoding: %w", err)
	}

	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("invalid JWT header: %w", err)
	}

	// Explicitly reject "none" algorithm and any algorithm that is not HS256
	if strings.EqualFold(header.Alg, "none") {
		return nil, fmt.Errorf("JWT algorithm 'none' is not permitted")
	}
	if header.Alg != "HS256" {
		return nil, fmt.Errorf("unsupported JWT algorithm: %s", header.Alg)
	}

	// Verify signature
	signingInput := parts[0] + "." + parts[1]
	signature, err := base64URLDecode(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid JWT signature encoding: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	expectedSig := mac.Sum(nil)

	if subtle.ConstantTimeCompare(signature, expectedSig) != 1 {
		return nil, fmt.Errorf("invalid JWT signature")
	}

	// Decode claims
	claimsJSON, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid JWT claims encoding: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("invalid JWT claims: %w", err)
	}

	now := time.Now().Unix()

	// Check expiration (exp)
	if claims.ExpiresAt > 0 && now > claims.ExpiresAt {
		return nil, fmt.Errorf("JWT expired at %d, current time %d", claims.ExpiresAt, now)
	}

	// Check not-before (nbf) -- token must not be used before this time
	if claims.NotBefore > 0 && now < claims.NotBefore {
		return nil, fmt.Errorf("JWT not valid before %d, current time %d", claims.NotBefore, now)
	}

	// Check issuer
	if expectedIssuer != "" && claims.Issuer != expectedIssuer {
		return nil, fmt.Errorf("JWT issuer mismatch: got %q, want %q", claims.Issuer, expectedIssuer)
	}

	// Require subject claim
	if claims.Subject == "" {
		return nil, fmt.Errorf("JWT missing required 'sub' claim")
	}

	return &claims, nil
}

// validateAPIKey checks a key against the configured list using constant-time
// comparison for every entry to prevent timing side-channel attacks.
func validateAPIKey(key string, keys []APIKeyEntry) (*APIKeyEntry, bool) {
	keyBytes := []byte(key)
	var matched *APIKeyEntry

	for i := range keys {
		if subtle.ConstantTimeCompare(keyBytes, []byte(keys[i].Key)) == 1 {
			matched = &keys[i]
		}
	}

	if matched == nil {
		return nil, false
	}
	return matched, true
}

// base64URLDecode decodes a base64url-encoded string (no padding).
func base64URLDecode(s string) ([]byte, error) {
	// Add padding if necessary
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// CreateJWT creates a signed JWT token (HS256) for testing or internal service-to-service auth.
func CreateJWT(secret, subject, role, issuer string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		Subject:   subject,
		Role:      role,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(duration).Unix(),
		NotBefore: now.Unix(),
		Issuer:    issuer,
	}

	header := map[string]string{"alg": "HS256", "typ": "JWT"}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	signature := mac.Sum(nil)
	sigB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + sigB64, nil
}
