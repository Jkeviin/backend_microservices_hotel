package middleware

import (
	"context"
	"net/http"
)

const (
	// IdempotencyKeyHeader es el header HTTP para la clave de idempotencia.
	IdempotencyKeyHeader = "Idempotency-Key"
	// IdempotencyKeyCtx es la clave de contexto para la idempotency key.
	IdempotencyKeyCtx contextKey = "idempotency_key"
)

// Idempotency extrae el header Idempotency-Key y lo inyecta en el context.
// Si el header no esta presente en una peticion POST, retorna 400.
func Idempotency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get(IdempotencyKeyHeader)
		if key == "" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"header Idempotency-Key es requerido para POST"}`))
			return
		}

		if key != "" {
			ctx := context.WithValue(r.Context(), IdempotencyKeyCtx, key)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetIdempotencyKey extrae la idempotency key del contexto.
func GetIdempotencyKey(ctx context.Context) string {
	if v, ok := ctx.Value(IdempotencyKeyCtx).(string); ok {
		return v
	}
	return ""
}
