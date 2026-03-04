package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/wb-go/wbf/ginext"
)

// RateLimiter — простой rate limiter на основе скользящего окна
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	limit    int           // Максимальное количество запросов
	window   time.Duration // Окно времени
}

type visitor struct {
	count     int
	lastReset time.Time
}

// NewRateLimiter создает новый rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}

	// Очистка старых записей каждые window*2
	go func() {
		for range time.Tick(window * 2) {
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, v := range rl.visitors {
		if time.Since(v.lastReset) > rl.window {
			delete(rl.visitors, ip)
		}
	}
}

// Allow проверяет, можно ли сделать запрос
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{count: 1, lastReset: time.Now()}
		return true
	}

	if time.Since(v.lastReset) > rl.window {
		v.count = 1
		v.lastReset = time.Now()
		return true
	}

	if v.count >= rl.limit {
		return false
	}

	v.count++
	return true
}

// Middleware возвращает middleware для rate limiting
func (rl *RateLimiter) Middleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		ip := c.ClientIP()

		if !rl.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, ginext.H{
				"error": "too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
