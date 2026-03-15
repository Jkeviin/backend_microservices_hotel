package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder captura el status code escrito por el handler.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// Logging registra cada request con slog incluyendo method, path, status,
// duration_ms y request_id.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		reqID := RequestIDFromContext(r.Context())

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", reqID,
		)
	})
}
