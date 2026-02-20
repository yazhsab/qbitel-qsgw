package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimitConfig configures the rate limiter.
type RateLimitConfig struct {
	// RequestsPerWindow is the maximum number of requests allowed per window.
	RequestsPerWindow int
	// Window is the time window for rate limiting.
	Window time.Duration
	// KeyFunc extracts the rate limit key from a request (e.g., IP, API key).
	// If nil, defaults to the remote address.
	KeyFunc func(r *http.Request) string
	// CleanupInterval controls how often expired entries are purged. Default: 5 minutes.
	CleanupInterval time.Duration
	// MaxEntries caps the number of tracked keys to prevent unbounded memory growth.
	// When this limit is reached, new keys are rate-limited immediately. Default: 100000.
	MaxEntries int
}

// DefaultRateLimitConfig returns a reasonable default rate limit config.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerWindow: 100,
		Window:            time.Minute,
		KeyFunc:           nil,
		CleanupInterval:   5 * time.Minute,
		MaxEntries:        100_000,
	}
}

type rateLimitEntry struct {
	count    int
	windowAt time.Time
}

// RateLimiter holds the state for rate limiting and exposes a Stop method
// to cleanly shut down the background cleanup goroutine.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	cfg     RateLimitConfig
	done    chan struct{}
}

// NewRateLimiter creates a rate limiter with a background cleanup goroutine.
// Call Stop() on shutdown to prevent goroutine leaks.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	if cfg.RequestsPerWindow <= 0 {
		cfg.RequestsPerWindow = 100
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(r *http.Request) string {
			return r.RemoteAddr
		}
	}
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = 5 * time.Minute
	}
	if cfg.MaxEntries <= 0 {
		cfg.MaxEntries = 100_000
	}

	rl := &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		cfg:     cfg,
		done:    make(chan struct{}),
	}

	// Background cleanup goroutine — stops when done is closed
	go func() {
		ticker := time.NewTicker(cfg.CleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-rl.done:
				return
			case <-ticker.C:
				rl.cleanup()
			}
		}
	}()

	return rl
}

// Stop terminates the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

// cleanup removes expired entries under lock.
func (rl *RateLimiter) cleanup() {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()
	for key, entry := range rl.entries {
		if now.Sub(entry.windowAt) > rl.cfg.Window*2 {
			delete(rl.entries, key)
		}
	}
}

// Middleware returns the HTTP middleware function.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := rl.cfg.KeyFunc(r)
			now := time.Now()

			// Entire rate-limit check + update is done under a single lock hold
			// to prevent TOCTOU race conditions.
			rl.mu.Lock()

			entry, exists := rl.entries[key]
			if !exists || now.Sub(entry.windowAt) > rl.cfg.Window {
				// Check max entries cap to prevent unbounded memory growth
				if !exists && len(rl.entries) >= rl.cfg.MaxEntries {
					rl.mu.Unlock()
					// Too many tracked keys — reject to prevent DoS via memory exhaustion
					setRateLimitHeaders(w, rl.cfg.RequestsPerWindow, 0, rl.cfg.Window)
					w.Header().Set("Retry-After", "60")
					writeAuthError(w, http.StatusTooManyRequests, "rate limit exceeded")
					return
				}

				rl.entries[key] = &rateLimitEntry{count: 1, windowAt: now}
				rl.mu.Unlock()

				setRateLimitHeaders(w, rl.cfg.RequestsPerWindow, rl.cfg.RequestsPerWindow-1, rl.cfg.Window)
				next.ServeHTTP(w, r)
				return
			}

			entry.count++
			count := entry.count
			windowStart := entry.windowAt
			allowed := count <= rl.cfg.RequestsPerWindow

			rl.mu.Unlock()

			remaining := rl.cfg.RequestsPerWindow - count
			if remaining < 0 {
				remaining = 0
			}

			setRateLimitHeaders(w, rl.cfg.RequestsPerWindow, remaining, rl.cfg.Window)

			if !allowed {
				retryAfter := rl.cfg.Window - now.Sub(windowStart)
				if retryAfter < 0 {
					retryAfter = 0
				}
				w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
				writeAuthError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit returns middleware that enforces per-key rate limiting using a
// sliding window counter stored in memory.
//
// When the limit is exceeded, it returns 429 Too Many Requests with
// Retry-After and X-RateLimit-* headers.
//
// Note: This creates a RateLimiter internally. For production use where
// you need to call Stop() on shutdown, use NewRateLimiter directly.
func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	rl := NewRateLimiter(cfg)
	return rl.Middleware()
}

func setRateLimitHeaders(w http.ResponseWriter, limit, remaining int, window time.Duration) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(window.Seconds())))
}
