package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitStore allows the HTTP middleware to use a different backing store
// when limits need to be shared across server instances.
type RateLimitStore interface {
	Allow(key string) bool
}

type rateLimitEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// KeyedRateLimiter keeps a token bucket per key in this server process.
type KeyedRateLimiter struct {
	mu          sync.Mutex
	entries     map[string]*rateLimitEntry
	limit       rate.Limit
	burst       int
	idleTTL     time.Duration
	nextCleanup time.Time
}

func NewKeyedRateLimiter(limit rate.Limit, burst int, idleTTL time.Duration) *KeyedRateLimiter {
	if limit <= 0 || burst <= 0 || idleTTL <= 0 {
		panic("rate limiter requires a positive limit, burst, and idle TTL")
	}
	return &KeyedRateLimiter{
		entries: make(map[string]*rateLimitEntry),
		limit:   limit,
		burst:   burst,
		idleTTL: idleTTL,
	}
}

func (l *KeyedRateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if !now.Before(l.nextCleanup) {
		for key, entry := range l.entries {
			if now.Sub(entry.lastSeen) >= l.idleTTL {
				delete(l.entries, key)
			}
		}
		l.nextCleanup = now.Add(l.idleTTL)
	}
	entry := l.entries[key]
	if entry == nil {
		entry = &rateLimitEntry{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.entries[key] = entry
	}
	entry.lastSeen = now
	return entry.limiter.AllowN(now, 1)
}

// RateLimit rejects requests when the store has no token for the request key.
func RateLimit(store RateLimitStore, key func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !store.Allow(key(r)) {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ClientIP uses the direct connection address, not client-supplied proxy headers.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
