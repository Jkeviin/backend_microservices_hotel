package model

import "errors"

// Errores de dominio del bounded context de inventario.
var (
	// ErrHabitacionNotFound indica que no se encontro una habitacion con el ID solicitado.
	ErrHabitacionNotFound = errors.New("habitacion no encontrada")

	// ErrTipoNotFound indica que no se encontro un tipo de habitacion con el ID solicitado.
	ErrTipoNotFound = errors.New("tipo de habitacion no encontrado")

	// ErrNumeroHabitacionDuplicado indica que ya existe una habitacion con ese numero.
	ErrNumeroHabitacionDuplicado = errors.New("el numero de habitacion ya esta en uso")

	// ErrEstadoInvalido indica que el estado proporcionado no es valido.
	ErrEstadoInvalido = errors.New("estado de habitacion invalido")

	// ErrNoMantenimientoSiOcupada indica que no se puede enviar a mantenimiento una habitacion ocupada.
	ErrNoMantenimientoSiOcupada = errors.New("no se puede enviar a mantenimiento una habitacion ocupada")

	// ErrNumeroVacio indica que el numero de habitacion no puede estar vacio.
	ErrNumeroVacio = errors.New("el numero de habitacion no puede estar vacio")

	// ErrNumeroMuyLargo indica que el numero de habitacion excede el maximo de 10 caracteres.
	ErrNumeroMuyLargo = errors.New("el numero de habitacion no puede exceder 10 caracteres")

	// ErrFiltroFechasIncompletas indica que se debe proporcionar ambas fechas o ninguna.
	ErrFiltroFechasIncompletas = errors.New("debe proporcionar tanto disponible_desde como disponible_hasta")

	// ErrFiltroFechasInvalidas indica que disponible_hasta debe ser posterior a disponible_desde.
	ErrFiltroFechasInvalidas = errors.New("disponible_hasta debe ser posterior a disponible_desde")
)
