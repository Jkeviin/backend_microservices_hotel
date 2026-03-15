package service

import (
	"context"

	"clientes/internal/domain/model"
	"clientes/internal/domain/repository"
)

// ClienteDomainService encapsula reglas de negocio que requieren acceso al
// repositorio (por ejemplo, unicidad de email).
type ClienteDomainService struct {
	repo repository.ClienteRepository
}

// NewClienteDomainService crea un nuevo domain service.
func NewClienteDomainService(repo repository.ClienteRepository) *ClienteDomainService {
	return &ClienteDomainService{repo: repo}
}

// CrearCliente valida unicidad de email y persiste el nuevo cliente.
func (s *ClienteDomainService) CrearCliente(ctx context.Context, nombre string, email model.Email, telefono *string) (*model.Cliente, error) {
	// Verificar unicidad de email antes de crear.
	existing, err := s.repo.GetByEmail(ctx, email.String())
	if err == nil && existing != nil {
		return nil, model.ErrEmailDuplicated
	}

	cliente, err := model.NewCliente(nombre, email, telefono)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, cliente); err != nil {
		return nil, err
	}

	return cliente, nil
}

// ValidarEmailUnico verifica que un email no este en uso por otro cliente
// distinto al indicado por excludeID.
func (s *ClienteDomainService) ValidarEmailUnico(ctx context.Context, email string, excludeID int) error {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		// No encontrado => email disponible.
		return nil
	}
	if existing.ID != excludeID {
		return model.ErrEmailDuplicated
	}
	return nil
}
