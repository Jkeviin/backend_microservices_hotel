package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// SetupRoutes configura las rutas del servicio de clientes en el router chi
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

	r.Route("/api/clientes", func(r chi.Router) {
		// Todas las rutas requieren autenticacion.
		r.Use(authMw)

		// GET: recepcion o admin.
		r.With(requireRole("recepcion", "admin")).
			Get("/{id}", handler.GetCliente)

		// POST y PATCH: recepcion o admin + rate limit en escrituras.
		r.With(requireRole("recepcion", "admin"), rateLimit).
			Post("/", handler.CreateCliente)

		r.With(requireRole("recepcion", "admin"), rateLimit).
			Patch("/{id}", handler.UpdateCliente)
	})
}
