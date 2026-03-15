package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"reservas/internal/infrastructure/http/middleware"
)

// SetupRoutes configura las rutas del servicio de reservas en el router chi
// proporcionado. Recibe middlewares como funciones para inyeccion de
// dependencias limpia.
func SetupRoutes(
	r chi.Router,
	handler *Handler,
	authMw func(http.Handler) http.Handler,
	logging func(http.Handler) http.Handler,
	requestID func(http.Handler) http.Handler,
	rateLimit func(http.Handler) http.Handler,
	requireRole func(roles ...string) func(http.Handler) http.Handler,
) {
	r.Use(requestID)
	r.Use(logging)

	r.Route("/api/reservas", func(r chi.Router) {
		// Todas las rutas requieren autenticacion.
		r.Use(authMw)

		// IMPORTANTE: registrar /estados ANTES de /{id} para que chi no
		// capture "estados" como un id.
		r.With(requireRole("recepcion", "admin")).
			Get("/estados", handler.ListEstados)

		// GET: recepcion o admin.
		r.With(requireRole("recepcion", "admin")).
			Get("/{id}", handler.GetReserva)

		// POST: idempotency middleware + recepcion/admin + rate limit.
		r.With(
			requireRole("recepcion", "admin"),
			middleware.RequireIdempotencyKey,
			rateLimit,
		).Post("/", handler.CreateReserva)

		// PATCH: recepcion/admin + rate limit.
		r.With(requireRole("recepcion", "admin"), rateLimit).
			Patch("/{id}", handler.PatchReserva)
	})
}
