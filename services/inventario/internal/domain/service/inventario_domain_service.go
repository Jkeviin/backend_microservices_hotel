package service

import "inventario/internal/domain/model"

// ValidarCambioEstado verifica que la transicion de estado sea valida.
// Regla: no se puede pasar a mantenimiento si la habitacion esta ocupada.
func ValidarCambioEstado(actual, nuevo model.EstadoHabitacion) error {
	if nuevo == model.EstadoMantenimiento && actual == model.EstadoOcupada {
		return model.ErrNoMantenimientoSiOcupada
	}
	return nil
}
