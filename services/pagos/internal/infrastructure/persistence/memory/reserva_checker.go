package memory

import (
	"context"
	"sync"

	"pagos/internal/application"
	"pagos/internal/domain/model"
)

// reservaEstado contiene el estado simulado de una reserva.
type reservaEstado struct {
	IDEstado     int
	NombreEstado string
}

// ReservaChecker implementa application.ReservaChecker usando datos en memoria.
type ReservaChecker struct {
	mu       sync.RWMutex
	reservas map[int]*reservaEstado
}

// NewReservaChecker crea un nuevo checker de reservas en memoria con datos mock.
func NewReservaChecker() *ReservaChecker {
	return &ReservaChecker{
		reservas: map[int]*reservaEstado{
			1: {IDEstado: 3, NombreEstado: "confirmada"},
			2: {IDEstado: 1, NombreEstado: "pendiente_pago"},
			3: {IDEstado: 4, NombreEstado: "cancelada"},
		},
	}
}

// GetReservaParaPago obtiene la informacion de una reserva simulada.
func (rc *ReservaChecker) GetReservaParaPago(_ context.Context, idReserva int) (*application.ReservaInfo, error) {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	r, ok := rc.reservas[idReserva]
	if !ok {
		return nil, model.ErrReservaNoValida
	}

	return &application.ReservaInfo{
		ID:           idReserva,
		IDEstado:     r.IDEstado,
		NombreEstado: r.NombreEstado,
	}, nil
}

// ConfirmarReserva cambia el estado de la reserva a 'confirmada' en memoria.
func (rc *ReservaChecker) ConfirmarReserva(_ context.Context, idReserva int) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	r, ok := rc.reservas[idReserva]
	if !ok {
		return model.ErrReservaNoValida
	}

	r.IDEstado = 3
	r.NombreEstado = "confirmada"
	return nil
}
