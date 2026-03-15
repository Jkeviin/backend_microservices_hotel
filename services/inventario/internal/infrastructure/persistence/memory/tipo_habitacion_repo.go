package memory

import (
	"context"

	"inventario/internal/domain/model"
)

// TipoHabitacionRepo es una implementacion en memoria del repositorio de tipos de habitacion.
type TipoHabitacionRepo struct {
	tipos []model.TipoHabitacion
}

// NewTipoHabitacionRepo crea un repositorio pre-cargado con tipos realistas.
func NewTipoHabitacionRepo() *TipoHabitacionRepo {
	return &TipoHabitacionRepo{
		tipos: []model.TipoHabitacion{
			{ID: 1, Nombre: "Simple", PrecioBase: 80.00, Capacidad: 2},
			{ID: 2, Nombre: "Doble", PrecioBase: 120.00, Capacidad: 3},
			{ID: 3, Nombre: "Suite", PrecioBase: 250.00, Capacidad: 4},
		},
	}
}

// List retorna todos los tipos de habitacion.
func (r *TipoHabitacionRepo) List(_ context.Context) ([]model.TipoHabitacion, error) {
	result := make([]model.TipoHabitacion, len(r.tipos))
	copy(result, r.tipos)
	return result, nil
}
