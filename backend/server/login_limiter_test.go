package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestLoginLimiter_Allow(t *testing.T) {
	burst := 3
	limiter := NewLoginLimiter(burst, rate.Limit(10), 1*time.Minute)
	key := "test_key"

	// First N requests should be allowed (burst limit)
	for i := 0; i < burst; i++ {
		allowed, retryAfter := limiter.Allow(key)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
		assert.Zero(t, retryAfter)
		limiter.OnFailure(key) // Record a failed attempt (consumes a token)
	}

	// Next request should be denied
	allowed, retryAfter := limiter.Allow(key)
	assert.False(t, allowed, "Request after burst limit should be denied")
	assert.Greater(t, retryAfter, 0, "retryAfter should be greater than 0")

	// Wait for a token to be replenished
	time.Sleep(100 * time.Millisecond)

	// Next request should be allowed again
	allowed, retryAfter = limiter.Allow(key)
	assert.True(t, allowed, "Request after waiting should be allowed")
	assert.Zero(t, retryAfter)
}

func TestLoginLimiter_TTL(t *testing.T) {
	ttl := 100 * time.Millisecond
	limiter := NewLoginLimiter(1, rate.Limit(10), ttl)
	key := "test_key"

	// Create a limiter for a key
	limiter.getLimiter(key)
	_, ok := limiter.limiters[key]
	assert.True(t, ok, "limiter should exist for the key")

	// Wait for TTL to expire
	time.Sleep(ttl + 50*time.Millisecond)

	limiter.mu.Lock()
	_, ok = limiter.limiters[key]
	assert.False(t, ok, "limiter should have been removed after TTL")
	limiter.mu.Unlock()
}

func TestLoginLimiter_Middleware(t *testing.T) {
	burst := 2
	limiter := NewLoginLimiter(burst, rate.Limit(10), 1*time.Minute)

	keyFn := func(r *http.Request) string {
		return r.Header.Get("X-Test-IP")
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate failed login by calling OnFailure (consumes a token)
		if r.URL.Query().Get("fail") == "true" {
			limiter.OnFailure(keyFn(r))
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	middleware := limiter.Middleware(keyFn)(testHandler)
	server := httptest.NewServer(middleware)
	defer server.Close()

	client := server.Client()
	ip := "192.168.1.100"

	// First N requests should be allowed (burst limit), simulate failed logins
	for i := 0; i < burst; i++ {
		req, _ := http.NewRequest("POST", server.URL+"?fail=true", nil)
		req.Header.Set("X-Test-IP", ip)
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Next request should be rate limited
	req, _ := http.NewRequest("POST", server.URL, nil)
	req.Header.Set("X-Test-IP", ip)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	retryAfterStr := resp.Header.Get("Retry-After")
	assert.NotEmpty(t, retryAfterStr, "Retry-After header should be set")
	retryAfter, err := strconv.Atoi(retryAfterStr)
	assert.NoError(t, err)
	assert.Greater(t, retryAfter, 0, "retryAfter should be positive")
}

func TestLoginLimiter_Concurrent(t *testing.T) {
	burst := 10
	// Set a zero replenishment rate to disable it for this test
	limiter := NewLoginLimiter(burst, rate.Limit(0), 1*time.Minute)
	key := "test_key"
	numRequests := 100
	var wg sync.WaitGroup
	wg.Add(numRequests)

	// Fire off a bunch of concurrent failures (consumes tokens)
	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			limiter.OnFailure(key)
		}()
	}
	wg.Wait()

	// After more failures than the burst size, the limiter should be exhausted.
	// Check that a subsequent request is not allowed.
	allowed, retryAfter := limiter.Allow(key)
	assert.False(t, allowed, "Expected subsequent requests to be denied")
	assert.Greater(t, retryAfter, 0, "Expected a retry-after duration")

	// Verify the internal state
	limiter.mu.Lock()
	internalLimiter, ok := limiter.limiters[key]
	limiter.mu.Unlock()
	assert.True(t, ok)
	// Tokens can be slightly negative due to concurrent decrements
	assert.LessOrEqual(t, internalLimiter.Tokens(), 0.0)
}
