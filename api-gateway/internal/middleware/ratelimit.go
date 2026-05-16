package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	limiters sync.Map
)

func getRateLimiter(ip string) *rate.Limiter {
	if v, ok := limiters.Load(ip); ok {
		return v.(*rate.Limiter)
	}
	l := rate.NewLimiter(10, 20)
	limiters.Store(ip, l)
	return l
}

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := getRateLimiter(c.ClientIP())
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}
