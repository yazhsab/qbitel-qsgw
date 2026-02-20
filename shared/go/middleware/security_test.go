package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityHeaders_Default(t *testing.T) {
	mw := SecurityHeaders(DefaultSecurityHeadersConfig())
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expected := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":          "DENY",
		"X-XSS-Protection":         "0",
		"Content-Security-Policy":  "default-src 'none'",
		"Referrer-Policy":          "strict-origin-when-cross-origin",
		"Cache-Control":            "no-store",
	}

	for header, want := range expected {
		got := rec.Header().Get(header)
		if got != want {
			t.Errorf("header %s: expected %q, got %q", header, want, got)
		}
	}

	hsts := rec.Header().Get("Strict-Transport-Security")
	if !strings.Contains(hsts, "max-age=63072000") {
		t.Errorf("expected HSTS header with max-age=63072000, got %q", hsts)
	}
}

func TestCORS_PreflightAllowed(t *testing.T) {
	mw := CORS(CORSConfig{
		AllowedOrigins: []string{"https://app.quantun.io"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         3600,
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/v1/keys", nil)
	req.Header.Set("Origin", "https://app.quantun.io")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", rec.Code)
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.quantun.io" {
		t.Errorf("expected origin header 'https://app.quantun.io', got %q", got)
	}

	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST" {
		t.Errorf("expected methods 'GET, POST', got %q", got)
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	mw := CORS(CORSConfig{
		AllowedOrigins: []string{"https://app.quantun.io"},
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	req.Header.Set("Origin", "https://evil.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (request still served), got %d", rec.Code)
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS header for disallowed origin, got %q", got)
	}
}

func TestCORS_Wildcard(t *testing.T) {
	mw := CORS(CORSConfig{
		AllowedOrigins: []string{"*"},
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	req.Header.Set("Origin", "https://any-origin.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected wildcard origin, got %q", got)
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	mw := CORS(CORSConfig{
		AllowedOrigins: []string{"https://app.quantun.io"},
	})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/keys", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS header without Origin, got %q", got)
	}
}

func TestMaxBodySize_Allowed(t *testing.T) {
	mw := MaxBodySize(1024)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader(`{"name": "test"}`)
	req := httptest.NewRequest("POST", "/api/v1/keys", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMaxBodySize_TooLarge(t *testing.T) {
	mw := MaxBodySize(16)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := strings.NewReader(`{"name": "this is a very long body that exceeds the limit"}`)
	req := httptest.NewRequest("POST", "/api/v1/keys", body)
	req.ContentLength = int64(body.Len())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", rec.Code)
	}
}
