package http

import (
	"time"

	"github.com/go-chi/chi/v5"

	"inventario/internal/infrastructure/http/middleware"
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

		// GET: accesible por recepcion y admin
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole("recepcion", "admin"))

			r.Get("/habitaciones", handler.ListHabitaciones)
			r.Get("/habitaciones/{id}", handler.GetHabitacion)
			r.Get("/tipos-habitacion", handler.ListTiposHabitacion)
		})

		// PATCH: solo admin
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole("admin"))
			r.Use(middleware.RateLimit(30, 1*time.Minute))

			r.Patch("/habitaciones/{id}", handler.UpdateHabitacion)
		})
	})

	return r
}
