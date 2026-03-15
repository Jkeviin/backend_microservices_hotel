package memory

import (
	"context"
	"fmt"

	"reservas/internal/domain/model"
	"reservas/internal/domain/repository"
)

// EstadoRepo implementa repository.EstadoReservaRepository en memoria.
type EstadoRepo struct {
	data []model.EstadoReserva
}

// NewEstadoRepo crea un repositorio en memoria pre-cargado con los 4 estados.
func NewEstadoRepo() repository.EstadoReservaRepository {
	return &EstadoRepo{
		data: []model.EstadoReserva{
			{ID: 1, Nombre: model.EstadoPendientePago},
			{ID: 2, Nombre: model.EstadoConfirmada},
			{ID: 3, Nombre: model.EstadoCancelada},
			{ID: 4, Nombre: model.EstadoCompletada},
		},
	}
}

// List retorna todos los estados de reserva.
func (r *EstadoRepo) List(_ context.Context) ([]model.EstadoReserva, error) {
	result := make([]model.EstadoReserva, len(r.data))
	copy(result, r.data)
	return result, nil
}

// GetByNombre obtiene un estado por su nombre.
func (r *EstadoRepo) GetByNombre(_ context.Context, nombre string) (*model.EstadoReserva, error) {
	for _, e := range r.data {
		if e.Nombre == nombre {
			cp := e
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("estado '%s' no encontrado", nombre)
}
