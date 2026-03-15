package repository

import (
	"context"

	"pagos/internal/domain/model"
)

// PagoRepository define las operaciones de persistencia para el aggregate Pago.
type PagoRepository interface {
	// Create persiste un nuevo pago.
	Create(ctx context.Context, pago *model.Pago) error

	// GetByID obtiene un pago por su ID.
	GetByID(ctx context.Context, id int) (*model.Pago, error)

	// GetByIdempotencyKey busca un pago existente por su clave de idempotencia.
	GetByIdempotencyKey(ctx context.Context, key string) (*model.Pago, error)
}
