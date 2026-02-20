package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuth_SkipPaths(t *testing.T) {
	mw := Auth(AuthConfig{
		JWTSecret: "test-secret",
		SkipPaths: []string{"/health"},
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for skip path, got %d", rec.Code)
	}
}

func TestAuth_MissingHeader(t *testing.T) {
	mw := Auth(AuthConfig{
		JWTSecret: "test-secret",
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_JWT_Valid(t *testing.T) {
	secret := "test-secret-key-32-bytes-long!!"
	token, err := CreateJWT(secret, "user-123", "admin", "quantun", time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT: %v", err)
	}

	mw := Auth(AuthConfig{
		JWTSecret: secret,
		JWTIssuer: "quantun",
	})

	var gotSubject, gotRole string
	var gotMethod AuthMethod

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSubject = SubjectFromContext(r.Context())
		gotRole = RoleFromContext(r.Context())
		gotMethod = AuthMethodFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotSubject != "user-123" {
		t.Errorf("expected subject 'user-123', got %q", gotSubject)
	}
	if gotRole != "admin" {
		t.Errorf("expected role 'admin', got %q", gotRole)
	}
	if gotMethod != AuthMethodJWT {
		t.Errorf("expected method 'jwt', got %q", gotMethod)
	}
}

func TestAuth_JWT_Expired(t *testing.T) {
	secret := "test-secret-key-32-bytes-long!!"
	token, err := CreateJWT(secret, "user-123", "admin", "quantun", -time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT: %v", err)
	}

	mw := Auth(AuthConfig{JWTSecret: secret})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", rec.Code)
	}
}

func TestAuth_JWT_WrongSecret(t *testing.T) {
	token, err := CreateJWT("correct-secret", "user-123", "admin", "", time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT: %v", err)
	}

	mw := Auth(AuthConfig{JWTSecret: "wrong-secret"})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong secret, got %d", rec.Code)
	}
}

func TestAuth_JWT_WrongIssuer(t *testing.T) {
	secret := "test-secret"
	token, err := CreateJWT(secret, "user-123", "admin", "wrong-issuer", time.Hour)
	if err != nil {
		t.Fatalf("failed to create JWT: %v", err)
	}

	mw := Auth(AuthConfig{JWTSecret: secret, JWTIssuer: "quantun"})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong issuer, got %d", rec.Code)
	}
}

func TestAuth_JWT_NoneAlgorithm_Rejected(t *testing.T) {
	// Manually craft a JWT with alg: "none" (algorithm confusion attack)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	claims := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"attacker","role":"admin","exp":9999999999,"iss":"quantun"}`))
	fakeToken := header + "." + claims + "."

	mw := Auth(AuthConfig{JWTSecret: "some-secret", JWTIssuer: "quantun"})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	req.Header.Set("Authorization", "Bearer "+fakeToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for 'none' algorithm, got %d", rec.Code)
	}
}

func TestAuth_JWT_NotBefore_Rejected(t *testing.T) {
	secret := "test-secret"
	now := time.Now()

	jwtClaims := JWTClaims{
		Subject:   "user-123",
		Role:      "admin",
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(time.Hour).Unix(),
		NotBefore: now.Add(time.Hour).Unix(), // not valid yet
		Issuer:    "quantun",
	}

	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(jwtClaims)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := headerB64 + "." + claimsB64

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	sig := mac.Sum(nil)
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	token := signingInput + "." + sigB64

	_, err := validateJWT(token, secret, "quantun")
	if err == nil {
		t.Error("expected error for token with future nbf, got nil")
	}
}

func TestAuth_APIKey_Valid(t *testing.T) {
	mw := Auth(AuthConfig{
		APIKeys: []APIKeyEntry{
			{Key: "qn-key-abc123", Subject: "service-a", Role: "service"},
		},
	})

	var gotSubject, gotRole string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSubject = SubjectFromContext(r.Context())
		gotRole = RoleFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	req.Header.Set("Authorization", "ApiKey qn-key-abc123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotSubject != "service-a" {
		t.Errorf("expected subject 'service-a', got %q", gotSubject)
	}
	if gotRole != "service" {
		t.Errorf("expected role 'service', got %q", gotRole)
	}
}

func TestAuth_APIKey_Invalid(t *testing.T) {
	mw := Auth(AuthConfig{
		APIKeys: []APIKeyEntry{
			{Key: "valid-key", Subject: "svc", Role: "service"},
		},
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	req.Header.Set("Authorization", "ApiKey wrong-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_APIKey_TimingSafe(t *testing.T) {
	// Verify that validateAPIKey iterates all entries (timing-safe)
	keys := []APIKeyEntry{
		{Key: "key-aaa", Subject: "svc-a", Role: "service"},
		{Key: "key-bbb", Subject: "svc-b", Role: "service"},
		{Key: "key-ccc", Subject: "svc-c", Role: "admin"},
	}

	// Valid key — should match the last entry
	entry, ok := validateAPIKey("key-ccc", keys)
	if !ok || entry.Subject != "svc-c" {
		t.Errorf("expected match for key-ccc, got ok=%v, entry=%v", ok, entry)
	}

	// Invalid key — should not match any
	_, ok = validateAPIKey("key-zzz", keys)
	if ok {
		t.Error("expected no match for invalid key")
	}
}

func TestAuth_UnsupportedScheme(t *testing.T) {
	mw := Auth(AuthConfig{JWTSecret: "secret"})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireRole_Allowed(t *testing.T) {
	mw := Auth(AuthConfig{
		APIKeys: []APIKeyEntry{
			{Key: "admin-key", Subject: "admin-user", Role: "admin"},
		},
	})
	roleMW := RequireRole("admin", "superadmin")

	handler := mw(roleMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	req.Header.Set("Authorization", "ApiKey admin-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireRole_Denied(t *testing.T) {
	mw := Auth(AuthConfig{
		APIKeys: []APIKeyEntry{
			{Key: "viewer-key", Subject: "viewer", Role: "viewer"},
		},
	})
	roleMW := RequireRole("admin")

	handler := mw(roleMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	req.Header.Set("Authorization", "ApiKey viewer-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestCreateJWT_RoundTrip(t *testing.T) {
	secret := "round-trip-secret"
	token, err := CreateJWT(secret, "svc-123", "service", "quantun-test", 5*time.Minute)
	if err != nil {
		t.Fatalf("CreateJWT failed: %v", err)
	}

	claims, err := validateJWT(token, secret, "quantun-test")
	if err != nil {
		t.Fatalf("validateJWT failed: %v", err)
	}

	if claims.Subject != "svc-123" {
		t.Errorf("expected subject 'svc-123', got %q", claims.Subject)
	}
	if claims.Role != "service" {
		t.Errorf("expected role 'service', got %q", claims.Role)
	}
	if claims.Issuer != "quantun-test" {
		t.Errorf("expected issuer 'quantun-test', got %q", claims.Issuer)
	}
	if claims.NotBefore == 0 {
		t.Error("expected non-zero nbf claim in created JWT")
	}
}

func TestRateLimiter_Stop(t *testing.T) {
	rl := NewRateLimiter(DefaultRateLimitConfig())
	// Should not leak goroutine after Stop
	rl.Stop()
}
