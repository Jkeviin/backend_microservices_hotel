package memory

import (
	"context"
	"sync"
	"time"

	"reservas/internal/domain/model"
	"reservas/internal/domain/repository"
)

// ReservaRepo implementa repository.ReservaRepository en memoria.
type ReservaRepo struct {
	mu     sync.Mutex
	data   map[int]*model.Reserva
	nextID int
}

// NewReservaRepo crea un repositorio en memoria pre-cargado con 3 reservas mock.
func NewReservaRepo() repository.ReservaRepository {
	r := &ReservaRepo{
		data:   make(map[int]*model.Reserva),
		nextID: 1,
	}

	now := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	seed := []model.Reserva{
		{
			IDCliente:    1,
			IDHabitacion: 1,
			IDEstado:     2, // confirmada
			NombreEstado: model.EstadoConfirmada,
			FechaInicio:  time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
			FechaFin:     time.Date(2025, 3, 18, 0, 0, 0, 0, time.UTC),
			Version:      1,
			CreatedAt:    now,
		},
		{
			IDCliente:    2,
			IDHabitacion: 3,
			IDEstado:     1, // pendiente_pago
			NombreEstado: model.EstadoPendientePago,
			FechaInicio:  time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
			FechaFin:     time.Date(2025, 4, 5, 0, 0, 0, 0, time.UTC),
			Version:      1,
			CreatedAt:    now,
		},
		{
			IDCliente:    3,
			IDHabitacion: 2,
			IDEstado:     3, // cancelada
			NombreEstado: model.EstadoCancelada,
			FechaInicio:  time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC),
			FechaFin:     time.Date(2025, 2, 12, 0, 0, 0, 0, time.UTC),
			Version:      2,
			CreatedAt:    now,
		},
	}

	for i := range seed {
		seed[i].ID = r.nextID
		stored := copyReserva(&seed[i])
		r.data[stored.ID] = stored
		r.nextID++
	}

	return r
}

// Create persiste una nueva reserva en memoria.
func (r *ReservaRepo) Create(_ context.Context, res *model.Reserva) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	res.ID = r.nextID
	r.nextID++
	res.CreatedAt = time.Now()

	stored := copyReserva(res)
	r.data[stored.ID] = stored

	return nil
}

// GetByID obtiene una reserva por ID.
func (r *ReservaRepo) GetByID(_ context.Context, id int) (*model.Reserva, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	res, ok := r.data[id]
	if !ok {
		return nil, model.ErrReservaNotFound
	}
	return copyReserva(res), nil
}

// Update actualiza una reserva existente con optimistic locking.
func (r *ReservaRepo) Update(_ context.Context, res *model.Reserva) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.data[res.ID]
	if !ok {
		return model.ErrReservaNotFound
	}

	// Optimistic locking: la version del update debe ser la version actual.
	if existing.Version != res.Version {
		return model.ErrVersionConflict
	}

	res.Version = existing.Version + 1
	stored := copyReserva(res)
	r.data[res.ID] = stored

	return nil
}

// ExisteSolapamiento verifica si hay reservas activas que se solapan con
// el rango para la habitacion indicada.
func (r *ReservaRepo) ExisteSolapamiento(_ context.Context, idHabitacion int, inicio, fin time.Time, excludeReservaID *int) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Estados activos (no cancelada).
	estadosActivos := map[string]bool{
		model.EstadoPendientePago: true,
		model.EstadoConfirmada:    true,
		model.EstadoCompletada:    true,
	}

	for _, res := range r.data {
		if res.IDHabitacion != idHabitacion {
			continue
		}
		if !estadosActivos[res.NombreEstado] {
			continue
		}
		if excludeReservaID != nil && res.ID == *excludeReservaID {
			continue
		}
		// Solapamiento: res.inicio < fin AND res.fin > inicio
		if res.FechaInicio.Before(fin) && res.FechaFin.After(inicio) {
			return true, nil
		}
	}

	return false, nil
}

// GetByIdempotencyKey busca una reserva por su clave de idempotencia.
func (r *ReservaRepo) GetByIdempotencyKey(_ context.Context, key string) (*model.Reserva, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, res := range r.data {
		if res.IdempotencyKey != nil && *res.IdempotencyKey == key {
			return copyReserva(res), nil
		}
	}

	return nil, model.ErrReservaNotFound
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func copyReserva(r *model.Reserva) *model.Reserva {
	cp := *r
	if r.Total != nil {
		t := *r.Total
		cp.Total = &t
	}
	if r.IdempotencyKey != nil {
		k := *r.IdempotencyKey
		cp.IdempotencyKey = &k
	}
	if r.UpdatedAt != nil {
		u := *r.UpdatedAt
		cp.UpdatedAt = &u
	}
	return &cp
}
