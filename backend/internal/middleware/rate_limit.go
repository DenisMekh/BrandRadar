package middleware

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	defaultRateLimitRPS   = 50000.0
	defaultRateLimitBurst = 1000000
)

type ipRateLimiter struct {
	mu       sync.Mutex
	rps      rate.Limit
	burst    int
	limiters map[string]*rate.Limiter
}

func NewIPRateLimiter(rps float64, burst int) gin.HandlerFunc {
	if rps <= 0 {
		rps = defaultRateLimitRPS
	}
	if burst <= 0 {
		burst = defaultRateLimitBurst
	}

	rl := &ipRateLimiter{
		rps:      rate.Limit(rps),
		burst:    burst,
		limiters: make(map[string]*rate.Limiter),
	}

	return rl.middleware()
}

func (r *ipRateLimiter) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := strings.TrimSpace(c.ClientIP())
		if clientIP == "" {
			clientIP = "unknown"
		}
		limiter := r.getLimiter(clientIP)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"data": nil,
				"error": gin.H{
					"code":    "TOO_MANY_REQUESTS",
					"message": "rate limit exceeded",
				},
			})
			return
		}

		c.Next()
	}
}

func (r *ipRateLimiter) getLimiter(clientIP string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	limiter, ok := r.limiters[clientIP]
	if ok {
		return limiter
	}

	limiter = rate.NewLimiter(r.rps, r.burst)
	r.limiters[clientIP] = limiter
	return limiter
}
