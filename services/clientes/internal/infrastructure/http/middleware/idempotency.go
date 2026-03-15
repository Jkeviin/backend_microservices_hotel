// Package middleware contiene middlewares HTTP para el servicio de clientes.
//
// Este archivo define un middleware de idempotencia basado en BD. En el
// servicio de Clientes NO se utiliza activamente porque no tiene operaciones
// POST criticas que requieran idempotency_key.
//
// Los servicios de Reservas y Pagos si usan idempotencia con sus columnas
// idempotency_key en la BD (ver AlterTables.sql).
//
// Se incluye aqui por consistencia entre servicios.
package middleware

import (
	"net/http"
)

const (
	// HeaderIdempotencyKey es el header HTTP esperado para la clave de
	// idempotencia.
	HeaderIdempotencyKey = "X-Idempotency-Key"
)

// IdempotencyChecker es la interfaz que deben implementar los stores de
// idempotencia (MySQL, Redis, etc.).
type IdempotencyChecker interface {
	// Check retorna true si la clave ya fue procesada, junto con la respuesta
	// cacheada (status, body).
	Check(key string) (exists bool, status int, body []byte, err error)

	// Store guarda el resultado de una operacion idempotente.
	Store(key string, status int, body []byte) error
}

// Idempotency retorna un middleware que verifica la clave de idempotencia.
// Si checker es nil el middleware es un no-op (pasa la request directamente).
func Idempotency(checker IdempotencyChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if checker == nil {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get(HeaderIdempotencyKey)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			exists, status, body, err := checker.Check(key)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if exists {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_, _ = w.Write(body)
				return
			}

			// Para un uso completo se deberia capturar la respuesta y
			// almacenarla con checker.Store(). Se deja como estructura base
			// para Reservas y Pagos.
			next.ServeHTTP(w, r)
		})
	}
}
