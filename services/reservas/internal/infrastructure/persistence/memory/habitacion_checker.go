package memory

import (
	"context"
	"fmt"

	domainservice "reservas/internal/domain/service"
)

// HabitacionChecker implementa service.HabitacionChecker en memoria.
// Simula habitaciones: 1-5,7-9 disponible, 6 mantenimiento.
type HabitacionChecker struct{}

// NewHabitacionChecker crea un checker de habitaciones mock.
func NewHabitacionChecker() domainservice.HabitacionChecker {
	return &HabitacionChecker{}
}

// GetEstadoHabitacion retorna el estado simulado de una habitacion.
func (c *HabitacionChecker) GetEstadoHabitacion(_ context.Context, id int) (string, error) {
	if id < 1 || id > 9 {
		return "", fmt.Errorf("habitacion %d no encontrada", id)
	}
	if id == 6 {
		return "mantenimiento", nil
	}
	return "disponible", nil
}
