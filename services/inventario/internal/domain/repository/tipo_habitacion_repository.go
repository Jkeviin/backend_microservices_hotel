package repository

import (
	"context"

	"inventario/internal/domain/model"
)

// TipoHabitacionRepository define las operaciones de lectura para tipos de habitacion.
type TipoHabitacionRepository interface {
	// List retorna todos los tipos de habitacion disponibles.
	List(ctx context.Context) ([]model.TipoHabitacion, error)
}
