package middleware

import (
	"context"
	"net/http"
)

const (
	// HeaderIdempotencyKey es el header HTTP esperado para la clave de
	// idempotencia.
	HeaderIdempotencyKey = "Idempotency-Key"

	// ContextKeyIdempotencyKey es la clave de contexto para la idempotency key.
	ContextKeyIdempotencyKey contextKey = "idempotency_key"
)

// IdempotencyKeyFromContext extrae la idempotency key del contexto.
func IdempotencyKeyFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ContextKeyIdempotencyKey).(string)
	return v
}

// RequireIdempotencyKey es un middleware que exige el header Idempotency-Key
// en requests POST. Extrae el valor y lo inyecta en el contexto.
// La verificacion real de idempotencia ocurre en el application layer.
func RequireIdempotencyKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			key := r.Header.Get(HeaderIdempotencyKey)
			if key == "" {
				http.Error(w, `{"error":"header Idempotency-Key es requerido para POST"}`, http.StatusBadRequest)
				return
			}
			ctx := context.WithValue(r.Context(), ContextKeyIdempotencyKey, key)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		next.ServeHTTP(w, r)
	})
}
