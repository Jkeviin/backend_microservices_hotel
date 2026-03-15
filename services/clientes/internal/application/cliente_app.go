package application

import (
	"context"
	"strings"
	"time"

	"clientes/internal/domain/model"
	"clientes/internal/domain/repository"
	domainservice "clientes/internal/domain/service"
)

// ---------------------------------------------------------------------------
// Commands (entrada)
// ---------------------------------------------------------------------------

// CreateClienteCommand contiene los datos necesarios para crear un cliente.
type CreateClienteCommand struct {
	Nombre   string  `json:"nombre"`
	Email    string  `json:"email"`
	Telefono *string `json:"telefono,omitempty"`
}

// UpdateClienteCommand contiene los campos actualizables de un cliente.
// Los campos nil no se modifican.
type UpdateClienteCommand struct {
	Nombre   *string `json:"nombre,omitempty"`
	Email    *string `json:"email,omitempty"`
	Telefono *string `json:"telefono,omitempty"`
}

// ---------------------------------------------------------------------------
// Response DTO (salida)
// ---------------------------------------------------------------------------

// ClienteResponse es el DTO de respuesta que se expone al exterior.
type ClienteResponse struct {
	ID        int        `json:"id"`
	Nombre    string     `json:"nombre"`
	Email     string     `json:"email"`
	Telefono  *string    `json:"telefono,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func toResponse(c *model.Cliente) *ClienteResponse {
	return &ClienteResponse{
		ID:        c.ID,
		Nombre:    c.Nombre,
		Email:     c.Email.String(),
		Telefono:  c.Telefono,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// ---------------------------------------------------------------------------
// Application Service
// ---------------------------------------------------------------------------

// ClienteAppService orquesta los casos de uso del bounded context de clientes.
type ClienteAppService struct {
	repo      repository.ClienteRepository
	domainSvc *domainservice.ClienteDomainService
}

// NewClienteAppService crea el application service con sus dependencias.
func NewClienteAppService(
	repo repository.ClienteRepository,
	domainSvc *domainservice.ClienteDomainService,
) *ClienteAppService {
	return &ClienteAppService{
		repo:      repo,
		domainSvc: domainSvc,
	}
}

// CreateCliente ejecuta el caso de uso de creacion de cliente.
func (s *ClienteAppService) CreateCliente(ctx context.Context, cmd CreateClienteCommand) (*ClienteResponse, error) {
	email, err := model.NewEmail(cmd.Email)
	if err != nil {
		return nil, err
	}

	cliente, err := s.domainSvc.CrearCliente(ctx, cmd.Nombre, email, cmd.Telefono)
	if err != nil {
		return nil, err
	}

	return toResponse(cliente), nil
}

// GetCliente obtiene un cliente por ID.
func (s *ClienteAppService) GetCliente(ctx context.Context, id int) (*ClienteResponse, error) {
	cliente, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResponse(cliente), nil
}

// UpdateCliente aplica una actualizacion parcial sobre un cliente existente.
func (s *ClienteAppService) UpdateCliente(ctx context.Context, id int, cmd UpdateClienteCommand) (*ClienteResponse, error) {
	cliente, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if cmd.Nombre != nil {
		nombre := strings.TrimSpace(*cmd.Nombre)
		if nombre == "" {
			return nil, model.ErrNombreRequerido
		}
		cliente.Nombre = nombre
	}

	if cmd.Email != nil {
		email, err := model.NewEmail(*cmd.Email)
		if err != nil {
			return nil, err
		}
		// Validar unicidad excluyendo al propio cliente.
		if err := s.domainSvc.ValidarEmailUnico(ctx, email.String(), id); err != nil {
			return nil, err
		}
		cliente.Email = email
	}

	if cmd.Telefono != nil {
		cliente.Telefono = cmd.Telefono
	}

	now := time.Now()
	cliente.UpdatedAt = &now

	if err := s.repo.Update(ctx, cliente); err != nil {
		return nil, err
	}

	return toResponse(cliente), nil
}
