package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	count    int
	windowAt time.Time
}

// RateLimiter implementa un rate limiter in-memory por IP con ventana fija.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

// NewRateLimiter crea un rate limiter con el limite y ventana indicados.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
}

// Middleware retorna un middleware HTTP que aplica el rate limit.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		rl.mu.Lock()
		v, ok := rl.visitors[ip]
		now := time.Now()
		if !ok || now.Sub(v.windowAt) > rl.window {
			rl.visitors[ip] = &visitor{count: 1, windowAt: now}
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		v.count++
		if v.count > rl.limit {
			rl.mu.Unlock()
			http.Error(w, `{"error":"demasiadas solicitudes, intente mas tarde"}`, http.StatusTooManyRequests)
			return
		}
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
