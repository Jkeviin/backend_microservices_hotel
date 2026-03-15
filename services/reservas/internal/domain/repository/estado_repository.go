package repository

import (
	"context"

	"reservas/internal/domain/model"
)

// EstadoReservaRepository define el contrato de persistencia para la entidad
// EstadoReserva.
type EstadoReservaRepository interface {
	// List retorna todos los estados de reserva.
	List(ctx context.Context) ([]model.EstadoReserva, error)

	// GetByNombre obtiene un estado por su nombre.
	GetByNombre(ctx context.Context, nombre string) (*model.EstadoReserva, error)
}
