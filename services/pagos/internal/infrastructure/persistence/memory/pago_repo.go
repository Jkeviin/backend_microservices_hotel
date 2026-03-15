package memory

import (
	"context"
	"sync"
	"time"

	"pagos/internal/domain/model"
)

// PagoRepo implementa PagoRepository usando almacenamiento en memoria.
type PagoRepo struct {
	mu        sync.RWMutex
	pagos     map[int]*model.Pago
	nextID    int
	idemIndex map[string]int // idempotency_key -> pago ID
}

// NewPagoRepo crea un nuevo repositorio de pagos en memoria con datos mock pre-cargados.
func NewPagoRepo() *PagoRepo {
	key1 := "mock-key-1"
	key2 := "mock-key-2"

	pagos := map[int]*model.Pago{
		1: {
			ID:             1,
			IDReserva:      1,
			Monto:          240.00,
			FechaPago:      time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
			EstadoPago:     model.EstadoPagoAprobado,
			IdempotencyKey: &key1,
			CreatedAt:      time.Date(2025, 3, 15, 10, 0, 0, 0, time.UTC),
		},
		2: {
			ID:             2,
			IDReserva:      2,
			Monto:          480.00,
			FechaPago:      time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
			EstadoPago:     model.EstadoPagoPendiente,
			IdempotencyKey: &key2,
			CreatedAt:      time.Date(2025, 4, 1, 10, 0, 0, 0, time.UTC),
		},
	}

	idemIndex := map[string]int{
		"mock-key-1": 1,
		"mock-key-2": 2,
	}

	return &PagoRepo{
		pagos:     pagos,
		nextID:    3,
		idemIndex: idemIndex,
	}
}

// Create persiste un nuevo pago en memoria con auto-increment.
func (r *PagoRepo) Create(_ context.Context, pago *model.Pago) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check idempotency
	if pago.IdempotencyKey != nil && *pago.IdempotencyKey != "" {
		if existingID, ok := r.idemIndex[*pago.IdempotencyKey]; ok {
			existing := r.pagos[existingID]
			*pago = *existing
			return nil
		}
	}

	pago.ID = r.nextID
	r.nextID++

	// Clonar para evitar mutaciones externas.
	clone := *pago
	r.pagos[pago.ID] = &clone

	if pago.IdempotencyKey != nil && *pago.IdempotencyKey != "" {
		r.idemIndex[*pago.IdempotencyKey] = pago.ID
	}

	return nil
}

// GetByID obtiene un pago por su ID.
func (r *PagoRepo) GetByID(_ context.Context, id int) (*model.Pago, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.pagos[id]
	if !ok {
		return nil, model.ErrPagoNotFound
	}

	clone := *p
	return &clone, nil
}

// GetByIdempotencyKey busca un pago por su clave de idempotencia.
func (r *PagoRepo) GetByIdempotencyKey(_ context.Context, key string) (*model.Pago, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.idemIndex[key]
	if !ok {
		return nil, model.ErrPagoNotFound
	}

	p := r.pagos[id]
	clone := *p
	return &clone, nil
}
