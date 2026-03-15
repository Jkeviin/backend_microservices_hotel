package repository

import (
	"context"

	"inventario/internal/domain/model"
)

// HabitacionRepository define las operaciones de persistencia para el aggregate Habitacion.
type HabitacionRepository interface {
	// List retorna habitaciones filtradas segun los criterios proporcionados.
	List(ctx context.Context, filtros model.HabitacionFiltros) ([]model.Habitacion, error)

	// GetByID retorna una habitacion por su identificador.
	// Retorna ErrHabitacionNotFound si no existe.
	GetByID(ctx context.Context, id int) (*model.Habitacion, error)

	// Update persiste los cambios de una habitacion existente.
	Update(ctx context.Context, h *model.Habitacion) error
}
