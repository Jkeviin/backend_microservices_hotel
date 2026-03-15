package memory

import (
	"context"
	"sync"
	"time"

	"clientes/internal/domain/model"
	"clientes/internal/domain/repository"
)

// ClienteRepo implementa repository.ClienteRepository en memoria.
// Pre-cargado con datos que coinciden con InserUserData.sql.
type ClienteRepo struct {
	mu      sync.Mutex
	data    map[int]*model.Cliente
	nextID  int
	emailIx map[string]int // email -> id para unicidad
}

// NewClienteRepo crea un repositorio en memoria pre-cargado con los 5
// clientes de InserUserData.sql.
func NewClienteRepo() repository.ClienteRepository {
	r := &ClienteRepo{
		data:    make(map[int]*model.Cliente),
		emailIx: make(map[string]int),
		nextID:  1,
	}

	now := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	seed := []struct {
		nombre   string
		email    string
		telefono *string
	}{
		{"Juan Perez", "juan.perez@email.com", strPtr("3001234567")},
		{"Maria Garcia", "maria.garcia@email.com", strPtr("3007654321")},
		{"Carlos Lopez", "carlos.lopez@email.com", strPtr("3011111111")},
		{"Ana Rodriguez", "ana.rodriguez@email.com", strPtr("3022222222")},
		{"Luis Martinez", "luis.martinez@email.com", strPtr("3033333333")},
	}

	for _, s := range seed {
		email, _ := model.NewEmail(s.email)
		c := &model.Cliente{
			ID:        r.nextID,
			Nombre:    s.nombre,
			Email:     email,
			Telefono:  s.telefono,
			CreatedAt: now,
		}
		r.data[c.ID] = c
		r.emailIx[s.email] = c.ID
		r.nextID++
	}

	return r
}

// Create persiste un nuevo cliente en memoria.
func (r *ClienteRepo) Create(_ context.Context, c *model.Cliente) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	email := c.Email.String()
	if _, exists := r.emailIx[email]; exists {
		return model.ErrEmailDuplicated
	}

	c.ID = r.nextID
	r.nextID++
	c.CreatedAt = time.Now()

	stored := copyCliente(c)
	r.data[stored.ID] = stored
	r.emailIx[email] = stored.ID

	return nil
}

// GetByID obtiene un cliente por ID.
func (r *ClienteRepo) GetByID(_ context.Context, id int) (*model.Cliente, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.data[id]
	if !ok {
		return nil, model.ErrClienteNotFound
	}
	return copyCliente(c), nil
}

// GetByEmail obtiene un cliente por email.
func (r *ClienteRepo) GetByEmail(_ context.Context, email string) (*model.Cliente, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, ok := r.emailIx[email]
	if !ok {
		return nil, model.ErrClienteNotFound
	}
	return copyCliente(r.data[id]), nil
}

// Update actualiza un cliente existente en memoria.
func (r *ClienteRepo) Update(_ context.Context, c *model.Cliente) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.data[c.ID]
	if !ok {
		return model.ErrClienteNotFound
	}

	newEmail := c.Email.String()
	oldEmail := existing.Email.String()

	// Verificar unicidad de email si cambio.
	if newEmail != oldEmail {
		if ownerID, taken := r.emailIx[newEmail]; taken && ownerID != c.ID {
			return model.ErrEmailDuplicated
		}
		delete(r.emailIx, oldEmail)
		r.emailIx[newEmail] = c.ID
	}

	stored := copyCliente(c)
	r.data[c.ID] = stored

	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func strPtr(s string) *string {
	return &s
}

func copyCliente(c *model.Cliente) *model.Cliente {
	cp := *c
	if c.Telefono != nil {
		t := *c.Telefono
		cp.Telefono = &t
	}
	if c.UpdatedAt != nil {
		u := *c.UpdatedAt
		cp.UpdatedAt = &u
	}
	return &cp
}
