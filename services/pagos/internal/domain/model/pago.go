package model

import (
	"strings"
	"time"
)

// EstadoPagoEnum es un value object que representa el estado de un pago.
type EstadoPagoEnum string

const (
	EstadoPagoPendiente  EstadoPagoEnum = "pendiente"
	EstadoPagoAprobado   EstadoPagoEnum = "aprobado"
	EstadoPagoRechazado  EstadoPagoEnum = "rechazado"
)

// NewEstadoPago crea un EstadoPagoEnum validado a partir de un string.
func NewEstadoPago(s string) (EstadoPagoEnum, error) {
	normalized := strings.ToLower(strings.TrimSpace(s))
	switch normalized {
	case "pendiente":
		return EstadoPagoPendiente, nil
	case "aprobado":
		return EstadoPagoAprobado, nil
	case "rechazado":
		return EstadoPagoRechazado, nil
	default:
		return "", ErrEstadoPagoInvalido
	}
}

// String retorna la representacion textual del estado.
func (e EstadoPagoEnum) String() string {
	return string(e)
}

// IsValid indica si el estado tiene un valor conocido.
func (e EstadoPagoEnum) IsValid() bool {
	switch e {
	case EstadoPagoPendiente, EstadoPagoAprobado, EstadoPagoRechazado:
		return true
	}
	return false
}

// ValidarMonto verifica que un monto sea mayor a cero.
func ValidarMonto(monto float64) error {
	if monto <= 0 {
		return ErrMontoInvalido
	}
	return nil
}

// Pago es el aggregate root del dominio de pagos.
type Pago struct {
	ID             int
	IDReserva      int
	Monto          float64
	FechaPago      time.Time
	EstadoPago     EstadoPagoEnum
	IdempotencyKey *string
	CreatedAt      time.Time
}
