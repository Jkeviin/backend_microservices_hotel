package model

import "errors"

var (
	// ErrPagoNotFound indica que el pago solicitado no existe.
	ErrPagoNotFound = errors.New("pago no encontrado")

	// ErrReservaNoValida indica que la reserva no existe o su estado no permite el pago.
	ErrReservaNoValida = errors.New("reserva no valida: no existe o estado no permite pago")

	// ErrMontoInvalido indica que el monto es menor o igual a cero.
	ErrMontoInvalido = errors.New("monto invalido: debe ser mayor a cero")

	// ErrEstadoPagoInvalido indica que el estado de pago proporcionado no es reconocido.
	ErrEstadoPagoInvalido = errors.New("estado de pago invalido")

	// ErrIdempotencyKeyDuplicada indica que ya existe un pago con esa idempotency key.
	ErrIdempotencyKeyDuplicada = errors.New("idempotency key duplicada: pago ya registrado")
)
