package model

// TipoHabitacion representa una categoria de habitacion con su precio y capacidad.
type TipoHabitacion struct {
	ID         int
	Nombre     string
	PrecioBase float64
	Capacidad  int
}
