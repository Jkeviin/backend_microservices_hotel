package repository

import (
	"context"
	"time"

	"reservas/internal/domain/model"
)

// ReservaRepository define el contrato de persistencia para el aggregate Reserva.
type ReservaRepository interface {
	// Create persiste una nueva reserva.
	Create(ctx context.Context, r *model.Reserva) error

	// GetByID obtiene una reserva por su ID. Retorna model.ErrReservaNotFound
	// si no existe.
	GetByID(ctx context.Context, id int) (*model.Reserva, error)

	// Update actualiza una reserva existente usando optimistic locking con
	// version. Retorna model.ErrVersionConflict si la version no coincide.
	Update(ctx context.Context, r *model.Reserva) error

	// ExisteSolapamiento verifica si hay reservas activas que se solapan con
	// el rango [inicio, fin) para la habitacion indicada. Opcionalmente excluye
	// una reserva por ID (para reprogramacion).
	ExisteSolapamiento(ctx context.Context, idHabitacion int, inicio, fin time.Time, excludeReservaID *int) (bool, error)

	// GetByIdempotencyKey busca una reserva por su clave de idempotencia.
	// Retorna model.ErrReservaNotFound si no existe.
	GetByIdempotencyKey(ctx context.Context, key string) (*model.Reserva, error)
}
