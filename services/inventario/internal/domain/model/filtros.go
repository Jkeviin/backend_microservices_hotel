package model

import "time"

// HabitacionFiltros contiene los criterios de busqueda para habitaciones.
type HabitacionFiltros struct {
	Tipo            *int
	Estado          *EstadoHabitacion
	DisponibleDesde *time.Time
	DisponibleHasta *time.Time
}

// Validar verifica que los filtros sean coherentes.
// Si se proporciona una fecha de disponibilidad, deben proporcionarse ambas,
// y disponible_hasta debe ser posterior a disponible_desde.
func (f HabitacionFiltros) Validar() error {
	tieneDesde := f.DisponibleDesde != nil
	tieneHasta := f.DisponibleHasta != nil

	if tieneDesde != tieneHasta {
		return ErrFiltroFechasIncompletas
	}

	if tieneDesde && tieneHasta {
		if !f.DisponibleHasta.After(*f.DisponibleDesde) {
			return ErrFiltroFechasInvalidas
		}
	}

	return nil
}
