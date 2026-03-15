package model

import (
	"time"
)

// ---------------------------------------------------------------------------
// Entity: EstadoReserva
// ---------------------------------------------------------------------------

// EstadoReserva representa un estado posible de una reserva.
type EstadoReserva struct {
	ID     int
	Nombre string
}

// Constantes de nombres de estado.
const (
	EstadoPendientePago = "pendiente_pago"
	EstadoConfirmada    = "confirmada"
	EstadoCancelada     = "cancelada"
	EstadoCompletada    = "completada"
)

// ---------------------------------------------------------------------------
// Value Object: FechaReserva
// ---------------------------------------------------------------------------

// FechaReserva encapsula el par (inicio, fin) de una reserva con sus
// invariantes: fin > inicio.
type FechaReserva struct {
	Inicio time.Time
	Fin    time.Time
}

// NewFechaReserva crea un FechaReserva validado.
// Si requireFuture es true, verifica que inicio no este en el pasado
// (para creacion de reservas nuevas).
func NewFechaReserva(inicio, fin time.Time, requireFuture bool) (FechaReserva, error) {
	// Normalizar a solo fecha (sin horas).
	inicio = truncateToDay(inicio)
	fin = truncateToDay(fin)

	if !fin.After(inicio) {
		return FechaReserva{}, ErrFechasInvalidas
	}

	if requireFuture {
		today := truncateToDay(time.Now())
		if inicio.Before(today) {
			return FechaReserva{}, ErrFechasInvalidas
		}
	}

	return FechaReserva{Inicio: inicio, Fin: fin}, nil
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// ---------------------------------------------------------------------------
// Aggregate Root: Reserva
// ---------------------------------------------------------------------------

// Reserva es el aggregate root del bounded context de reservas.
type Reserva struct {
	ID             int
	IDCliente      int
	IDHabitacion   int
	IDEstado       int
	NombreEstado   string // campo desnormalizado para respuestas
	FechaInicio    time.Time
	FechaFin       time.Time
	Total          *float64
	IdempotencyKey *string
	Version        int
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

// PuedeSerCancelada retorna true si la reserva no esta ya cancelada ni completada.
func (r *Reserva) PuedeSerCancelada() bool {
	return r.NombreEstado != EstadoCancelada && r.NombreEstado != EstadoCompletada
}

// PuedeSerReprogramada retorna true si la reserva puede cambiar sus fechas:
// no esta cancelada ni completada, y su fecha de inicio es posterior a hoy.
func (r *Reserva) PuedeSerReprogramada() bool {
	if r.NombreEstado == EstadoCancelada || r.NombreEstado == EstadoCompletada {
		return false
	}
	today := truncateToDay(time.Now())
	return !r.FechaInicio.Before(today.AddDate(0, 0, 1))
}

// Cancelar cambia el estado de la reserva a cancelada. Retorna error si
// el estado actual no lo permite.
func (r *Reserva) Cancelar(idEstadoCancelada int) error {
	if !r.PuedeSerCancelada() {
		return ErrEstadoNoPermiteCambio
	}
	r.IDEstado = idEstadoCancelada
	r.NombreEstado = EstadoCancelada
	now := time.Now()
	r.UpdatedAt = &now
	return nil
}

// Reprogramar cambia las fechas de la reserva. Verifica que el estado lo
// permita y que las nuevas fechas sean validas.
func (r *Reserva) Reprogramar(inicio, fin time.Time) error {
	if !r.PuedeSerReprogramada() {
		if r.NombreEstado == EstadoCancelada || r.NombreEstado == EstadoCompletada {
			return ErrEstadoNoPermiteCambio
		}
		return ErrReservaYaIniciada
	}

	fechas, err := NewFechaReserva(inicio, fin, true)
	if err != nil {
		return err
	}

	r.FechaInicio = fechas.Inicio
	r.FechaFin = fechas.Fin
	now := time.Now()
	r.UpdatedAt = &now
	return nil
}
