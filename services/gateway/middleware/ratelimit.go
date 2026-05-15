package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

func newIPLimiter(rps, burst int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

func (il *ipLimiter) get(ip string) *rate.Limiter {
	il.mu.Lock()
	defer il.mu.Unlock()

	if lim, ok := il.limiters[ip]; ok {
		return lim
	}

	lim := rate.NewLimiter(il.rps, il.burst)
	il.limiters[ip] = lim
	return lim
}

func RateLimit(rps, burst int) gin.HandlerFunc {
	limiter := newIPLimiter(rps, burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.get(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
