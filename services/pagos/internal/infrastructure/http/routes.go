package http

import (
	"time"

	"github.com/go-chi/chi/v5"

	"pagos/internal/infrastructure/http/middleware"
)

// NewRouter crea el router chi con todas las rutas y middlewares configurados.
func NewRouter(handler *Handler, jwtSecret string) *chi.Mux {
	r := chi.NewRouter()

	// Middlewares globales
	r.Use(middleware.CORS)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)

	// Rutas protegidas
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.Auth(jwtSecret))

		// POST: idempotency + roles (recepcion/admin) + rate limit
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole("recepcion", "admin"))
			r.Use(middleware.Idempotency)
			r.Use(middleware.RateLimit(30, 1*time.Minute))

			r.Post("/pagos", handler.CreatePago)
		})

		// GET: accesible por recepcion y admin
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole("recepcion", "admin"))

			r.Get("/pagos/{id}", handler.GetPago)
		})
	})

	return r
}
