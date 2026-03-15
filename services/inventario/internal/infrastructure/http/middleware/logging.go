package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter wrapper para capturar el status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging registra cada peticion HTTP con slog.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		reqID := ""
		if v, ok := r.Context().Value(RequestIDKey).(string); ok {
			reqID = v
		}

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", reqID,
		)
	})
}
