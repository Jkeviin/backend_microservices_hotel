package repository

import (
	"context"

	"clientes/internal/domain/model"
)

// ClienteRepository define el contrato de persistencia para el aggregate Cliente.
type ClienteRepository interface {
	// Create persiste un nuevo cliente. Retorna model.ErrEmailDuplicated si
	// el email ya existe.
	Create(ctx context.Context, c *model.Cliente) error

	// GetByID obtiene un cliente por su ID. Retorna model.ErrClienteNotFound
	// si no existe.
	GetByID(ctx context.Context, id int) (*model.Cliente, error)

	// GetByEmail obtiene un cliente por su email. Retorna model.ErrClienteNotFound
	// si no existe.
	GetByEmail(ctx context.Context, email string) (*model.Cliente, error)

	// Update actualiza un cliente existente. Retorna model.ErrClienteNotFound
	// si no existe, o model.ErrEmailDuplicated si el nuevo email ya esta en uso.
	Update(ctx context.Context, c *model.Cliente) error
}
