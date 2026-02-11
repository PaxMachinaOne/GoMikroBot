package httpmw

import (
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// Middleware wraps an http.Handler.
//
// The convention is:
//   mw(next) returns a handler that runs before calling next.
type Middleware func(http.Handler) http.Handler

// Chain applies middleware in the order provided.
//
// Chain(h, a, b) => a(b(h))
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Recoverer catches panics from downstream handlers and returns HTTP 500.
// It also logs the panic and stacktrace to stdout.
func Recoverer() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					fmt.Printf("❌ HTTP panic recovered: %v\n%s\n", rec, debug.Stack())
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// MaxBodyBytes limits request body size to n bytes.
//
// Note: this must run before a handler reads r.Body.
func MaxBodyBytes(n int64) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if n > 0 {
				r.Body = http.MaxBytesReader(w, r.Body, n)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter implements a simple token-bucket rate limiter.
// Default use is per-client-IP (best effort).
//
// This is intentionally dependency-free (no x/time).
//
// Semantics:
// - rps: refill tokens per second
// - burst: max tokens stored
// - cost: each request costs 1 token
//
// When out of tokens, responds 429.
type RateLimiter struct {
	rps   float64
	burst float64

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens  float64
	last    time.Time
	lastHit time.Time
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
	if rps <= 0 {
		rps = 5
	}
	if burst <= 0 {
		burst = 10
	}
	return &RateLimiter{
		rps:     rps,
		burst:   float64(burst),
		buckets: make(map[string]*bucket),
	}
}

func (rl *RateLimiter) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := clientIPKey(r)
			if !rl.allow(key) {
				w.Header().Set("Retry-After", "1")
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) allow(key string) bool {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b := rl.buckets[key]
	if b == nil {
		b = &bucket{tokens: rl.burst, last: now, lastHit: now}
		rl.buckets[key] = b
	}

	// Lazy cleanup to prevent unbounded growth.
	// Remove buckets idle for >10 minutes.
	if len(rl.buckets) > 4096 {
		for k, bb := range rl.buckets {
			if now.Sub(bb.lastHit) > 10*time.Minute {
				delete(rl.buckets, k)
			}
		}
	}

	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * rl.rps
		if b.tokens > rl.burst {
			b.tokens = rl.burst
		}
		b.last = now
	}
	b.lastHit = now

	if b.tokens < 1 {
		return false
	}
	b.tokens -= 1
	return true
}

func clientIPKey(r *http.Request) string {
	// If a reverse proxy exists, it should set X-Forwarded-For.
	// We intentionally take the *left-most* IP.
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	rip := r.Header.Get("X-Real-IP")
	if rip != "" {
		return strings.TrimSpace(rip)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}
