package httpapi

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"museum-of-recollections/apps/api/internal/health"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Readiness interface{ Ready(context.Context) error }

// New exposes only generic health statuses. Fixed token buckets bound memory
// without trusting proxy headers or retaining arbitrary client IPs. Each probe
// has an independent budget so unrelated traffic cannot exhaust liveness.
func New(service Readiness, log *zap.Logger) http.Handler {
	r := gin.New()
	r.Use(recovery(log), requestDeadline(3*time.Second), routeRateLimit())
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/readyz", func(c *gin.Context) {
		if err := service.Ready(c.Request.Context()); err != nil {
			var failure *health.Failure
			if errors.As(err, &failure) {
				log.Warn("readiness check failed", zap.Int("dependency_index", failure.DependencyIndex()), zap.String("failure_reason", failure.Reason()))
			} else {
				log.Warn("readiness check failed")
			}
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	return r
}

func routeRateLimit() gin.HandlerFunc {
	other, live, ready := newLimiter(20, 40), newLimiter(20, 40), newLimiter(20, 40)
	return func(c *gin.Context) {
		budget := other
		if c.Request.Method == http.MethodGet {
			switch c.FullPath() {
			case "/healthz":
				budget = live
			case "/readyz":
				budget = ready
			}
		}
		rateLimit(budget)(c)
	}
}
func requestDeadline(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
func recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recover() != nil {
				log.Error("request failed")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error"})
			}
		}()
		c.Next()
	}
}

type limiter struct {
	mu                  sync.Mutex
	tokens, rate, burst float64
	last                time.Time
}

func newLimiter(rate, burst float64) *limiter {
	return &limiter{tokens: burst, rate: rate, burst: burst, last: time.Now()}
}
func (l *limiter) allow(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if now.After(l.last) {
		l.tokens = min(l.burst, l.tokens+now.Sub(l.last).Seconds()*l.rate)
		l.last = now
	}
	if l.tokens < 1 {
		return false
	}
	l.tokens--
	return true
}
func rateLimit(l *limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.allow(time.Now()) {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"status": "rate_limited"})
			return
		}
		c.Next()
	}
}
