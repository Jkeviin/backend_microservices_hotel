package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	// RequestIDKey es la clave de contexto para el request ID.
	RequestIDKey contextKey = "request_id"
	// RequestIDHeader es el header HTTP para el request ID.
	RequestIDHeader = "X-Request-ID"
)

// RequestID asigna un identificador unico a cada peticion.
// Si el cliente envia X-Request-ID, lo reutiliza; si no, genera uno nuevo.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set(RequestIDHeader, id)
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
