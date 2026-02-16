package server

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// LoginLimiter is a specialized rate limiter for login attempts, allowing a certain number of failed attempts
// within a specified time frame before blocking further attempts for a cooldown period.
type LoginLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter

	rate  rate.Limit
	burst int
	ttl   time.Duration
}

// NewLoginLimiter creates a new LoginLimiter with the specified rate, burst size, and TTL for limiter entries.
// burst: maximum number of unsuccessful logins allowed to occur initially
// rateLimit: maximum frequency of allowed unsuccessful logins
// ttl: duration after which limiters are removed
func NewLoginLimiter(burst int, rateLimit rate.Limit, ttl time.Duration) *LoginLimiter {
	return &LoginLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rateLimit,
		burst:    burst,
		ttl:      ttl,
	}
}

// Allow checks if a login attempt is allowed without consuming a token.
// If denied, it returns the retry-after duration in whole seconds (>= 0).
func (l *LoginLimiter) Allow(key string) (bool, int) {
	limiter := l.getLimiter(key)

	// Check available tokens without consuming any (Tokens() is a read-only operation)
	if limiter.Tokens() >= 1 {
		return true, 0
	}

	// Not enough tokens - calculate wait time using Reserve+Cancel
	r := limiter.Reserve()
	delay := r.Delay()
	r.Cancel()

	retryAfter := int(delay.Seconds()) + 1
	return false, retryAfter
}

// OnFailure records a failed login attempt (consumes a token).
func (l *LoginLimiter) OnFailure(key string) {
	limiter := l.getLimiter(key)
	limiter.Allow()
}

func (l *LoginLimiter) getLimiter(key string) *rate.Limiter {
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

// Middleware returns a net/http middleware that applies the limiter's Allow check
// using a provided key extractor (e.g., client IP). If not allowed, it responds
// with HTTP 429 and sets Retry-After.
func (l *LoginLimiter) Middleware(keyFn func(*http.Request) string) func(http.Handler) http.Handler {
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
