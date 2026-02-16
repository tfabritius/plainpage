package server

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter is a generic rate limiter that consumes a token on every request.
// It uses a per-key token bucket algorithm.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter

	rate  rate.Limit
	burst int
	ttl   time.Duration
}

// NewRateLimiter creates a new RateLimiter with the specified rate, burst size, and TTL for limiter entries.
// burst: maximum number of requests allowed to occur initially (bucket size)
// rateLimit: rate at which tokens are replenished
// ttl: duration after which inactive limiters are cleaned up
func NewRateLimiter(burst int, rateLimit rate.Limit, ttl time.Duration) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rateLimit,
		burst:    burst,
		ttl:      ttl,
	}
}

// Allow checks if a request is allowed and consumes a token if so.
// Returns true if allowed, false if rate limited.
// If denied, retryAfter contains the number of seconds to wait (>= 1).
func (l *RateLimiter) Allow(key string) (allowed bool, retryAfter int) {
	limiter := l.getLimiter(key)

	if limiter.Allow() {
		return true, 0
	}

	// Not allowed - calculate wait time
	r := limiter.Reserve()
	delay := r.Delay()
	r.Cancel()

	retryAfter = int(delay.Seconds()) + 1
	return false, retryAfter
}

func (l *RateLimiter) getLimiter(key string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, ok := l.limiters[key]
	if !ok {
		limiter = rate.NewLimiter(l.rate, l.burst)
		l.limiters[key] = limiter

		time.AfterFunc(l.ttl, func() {
			l.mu.Lock()
			delete(l.limiters, key)
			l.mu.Unlock()
		})
	}

	return limiter
}

// Middleware returns a net/http middleware that applies rate limiting
// using a provided key extractor (e.g., client IP).
// If rate limit exceeded, responds with HTTP 429 and sets Retry-After header.
func (l *RateLimiter) Middleware(keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)
			allowed, retryAfter := l.Allow(key)
			if !allowed {
				if retryAfter < 1 {
					retryAfter = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))

				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
