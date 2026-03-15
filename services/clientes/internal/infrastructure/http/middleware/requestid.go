package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	// HeaderRequestID es el nombre del header HTTP para el request ID.
	HeaderRequestID = "X-Request-Id"

	// ContextKeyRequestID es la clave de contexto para el request ID.
	ContextKeyRequestID contextKey = "request_id"
)

// RequestIDFromContext extrae el request ID del contexto.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ContextKeyRequestID).(string)
	return v
}

// RequestID inyecta un identificador unico por request. Si el header
// X-Request-Id ya viene en la peticion lo reutiliza; de lo contrario genera
// un UUID v4.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderRequestID)
		if id == "" {
			id = uuid.New().String()
		}

		w.Header().Set(HeaderRequestID, id)

		ctx := context.WithValue(r.Context(), ContextKeyRequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
