package model

import "errors"

// Errores de dominio del aggregate Reserva.
var (
	// ErrReservaNotFound indica que no se encontro una reserva con el ID solicitado.
	ErrReservaNotFound = errors.New("reserva no encontrada")

	// ErrConflictoFechas indica solapamiento de fechas con otra reserva (409).
	ErrConflictoFechas = errors.New("conflicto de fechas: ya existe una reserva en ese rango para la habitacion")

	// ErrHabitacionNoDisponible indica que la habitacion esta en mantenimiento.
	ErrHabitacionNoDisponible = errors.New("la habitacion no esta disponible (en mantenimiento)")

	// ErrReservaYaIniciada indica que la reserva ya inicio y no puede reprogramarse.
	ErrReservaYaIniciada = errors.New("la reserva ya inicio o esta por iniciar, no puede reprogramarse")

	// ErrVersionConflict indica conflicto de optimistic locking (409).
	ErrVersionConflict = errors.New("conflicto de version: la reserva fue modificada por otro proceso")

	// ErrFechasInvalidas indica que las fechas no son validas (fin <= inicio o en el pasado).
	ErrFechasInvalidas = errors.New("fechas invalidas: la fecha fin debe ser posterior a la fecha inicio")

	// ErrEstadoNoPermiteCambio indica que el estado actual no permite la operacion.
	ErrEstadoNoPermiteCambio = errors.New("el estado actual de la reserva no permite esta operacion")

	// ErrIdempotencyKeyDuplicada indica que ya se proceso una reserva con esa clave.
	ErrIdempotencyKeyDuplicada = errors.New("la clave de idempotencia ya fue utilizada")
)
