package model

import (
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Value Object: NumeroHabitacion
// ---------------------------------------------------------------------------

// NumeroHabitacion representa un numero de habitacion validado.
// No puede estar vacio ni exceder 10 caracteres.
type NumeroHabitacion struct {
	value string
}

// NewNumeroHabitacion crea un NumeroHabitacion validado.
func NewNumeroHabitacion(raw string) (NumeroHabitacion, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return NumeroHabitacion{}, ErrNumeroVacio
	}
	if len(v) > 10 {
		return NumeroHabitacion{}, ErrNumeroMuyLargo
	}
	return NumeroHabitacion{value: v}, nil
}

// String retorna la representacion textual del numero de habitacion.
func (n NumeroHabitacion) String() string {
	return n.value
}

// ---------------------------------------------------------------------------
// Value Object: EstadoHabitacion
// ---------------------------------------------------------------------------

// EstadoHabitacion representa el estado actual de una habitacion.
type EstadoHabitacion struct {
	value string
}

// Constantes de estados validos.
var (
	EstadoDisponible    = EstadoHabitacion{value: "disponible"}
	EstadoMantenimiento = EstadoHabitacion{value: "mantenimiento"}
	EstadoOcupada       = EstadoHabitacion{value: "ocupada"}
)

// estadosValidos mapea strings a sus valores de EstadoHabitacion correspondientes.
var estadosValidos = map[string]EstadoHabitacion{
	"disponible":    EstadoDisponible,
	"mantenimiento": EstadoMantenimiento,
	"ocupada":       EstadoOcupada,
}

// NewEstadoHabitacion crea un EstadoHabitacion validado a partir de un string.
func NewEstadoHabitacion(raw string) (EstadoHabitacion, error) {
	v := strings.ToLower(strings.TrimSpace(raw))
	estado, ok := estadosValidos[v]
	if !ok {
		return EstadoHabitacion{}, ErrEstadoInvalido
	}
	return estado, nil
}

// String retorna la representacion textual del estado.
func (e EstadoHabitacion) String() string {
	return e.value
}

// IsValid retorna true si el estado tiene un valor valido.
func (e EstadoHabitacion) IsValid() bool {
	_, ok := estadosValidos[e.value]
	return ok
}

// ---------------------------------------------------------------------------
// Aggregate Root: Habitacion
// ---------------------------------------------------------------------------

// Habitacion es el aggregate root del bounded context de inventario.
type Habitacion struct {
	ID               int
	Numero           NumeroHabitacion
	IDTipoHabitacion int
	Estado           EstadoHabitacion
	UpdatedAt        *time.Time
}

// PuedeIrAMantenimiento retorna false si la habitacion esta ocupada,
// dado que no se puede enviar a mantenimiento una habitacion con huesped activo.
func (h *Habitacion) PuedeIrAMantenimiento() bool {
	return h.Estado != EstadoOcupada
}
