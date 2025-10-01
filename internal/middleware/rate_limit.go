package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"smart-transit-system/internal/auth"
	"smart-transit-system/internal/httpx"
)

type limiterKeyFunc func(*gin.Context) string

func RateLimit(perMinute int, keyFn limiterKeyFunc) gin.HandlerFunc {
	if perMinute <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	store := &limiterStore{
		limiters: make(map[string]*rate.Limiter),
		expiry:   make(map[string]time.Time),
		rate:     rate.Every(time.Minute / time.Duration(perMinute)),
		burst:    perMinute,
		ttl:      5 * time.Minute,
	}
	go store.cleanup()
	return func(c *gin.Context) {
		key := keyFn(c)
		limiter := store.getLimiter(key)
		if !limiter.Allow() {
			httpx.RespondError(c, 429, "rate_limit.exceeded", "rate limit exceeded", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

func AdminReadRateLimiter() gin.HandlerFunc {
	return RateLimit(30, subjectKey)
}

func AdminWriteRateLimiter() gin.HandlerFunc {
	return RateLimit(10, subjectKey)
}

func subjectKey(c *gin.Context) string {
	if ctx, ok := auth.ValuesFromContext(c); ok && ctx.Subject != "" {
		return ctx.Subject
	}
	return c.ClientIP()
}

type limiterStore struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	expiry   map[string]time.Time
	rate     rate.Limit
	burst    int
	ttl      time.Duration
}

func (s *limiterStore) getLimiter(key string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	lim, ok := s.limiters[key]
	if !ok {
		lim = rate.NewLimiter(s.rate, s.burst)
		s.limiters[key] = lim
	}
	s.expiry[key] = time.Now().Add(s.ttl)
	return lim
}

func (s *limiterStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for key, exp := range s.expiry {
			if exp.Before(now) {
				delete(s.expiry, key)
				delete(s.limiters, key)
			}
		}
		s.mu.Unlock()
	}
}
