package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimit limita la cantidad de peticiones de escritura por IP.
// Usa un contador simple con ventana de tiempo fija.
func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	type client struct {
		count    int
		resetAt  time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Limpieza periodica de entradas expiradas.
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for ip, c := range clients {
				if now.After(c.resetAt) {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			mu.Lock()
			c, exists := clients[ip]
			now := time.Now()

			if !exists || now.After(c.resetAt) {
				clients[ip] = &client{count: 1, resetAt: now.Add(window)}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			c.count++
			if c.count > maxRequests {
				mu.Unlock()
				http.Error(w, `{"error":"demasiadas solicitudes, intente mas tarde"}`, http.StatusTooManyRequests)
				return
			}
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
