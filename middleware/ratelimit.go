package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	visitors    map[string]map[string]*visitor
	mutex       sync.Mutex
	limit       rate.Limit
	burst       int
	entryTTL    time.Duration
	lastCleanup time.Time
}

func NewRateLimiter(limit time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]map[string]*visitor),
		limit:    rate.Every(limit),
		burst:    1,
		entryTTL: 10 * time.Minute,
	}
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		path := c.Request.URL.Path
		now := time.Now()

		rl.mutex.Lock()
		defer rl.mutex.Unlock()

		if _, exists := rl.visitors[clientIP]; !exists {
			rl.visitors[clientIP] = make(map[string]*visitor)
		}

		v, exists := rl.visitors[clientIP][path]
		if !exists {
			v = &visitor{limiter: rate.NewLimiter(rl.limit, rl.burst)}
			rl.visitors[clientIP][path] = v
		}

		v.lastSeen = now
		if !v.limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded. Please try again later."})
			c.Abort()
			return
		}

		rl.cleanup(now)
		c.Next()
	}
}

func (rl *RateLimiter) cleanup(now time.Time) {
	if now.Sub(rl.lastCleanup) < time.Minute {
		return
	}

	for ip, paths := range rl.visitors {
		for path, v := range paths {
			if now.Sub(v.lastSeen) > rl.entryTTL {
				delete(paths, path)
			}
		}
		if len(paths) == 0 {
			delete(rl.visitors, ip)
		}
	}

	rl.lastCleanup = now
}
