package model

import "errors"

// Errores de dominio del aggregate Cliente.
var (
	// ErrClienteNotFound indica que no se encontro un cliente con el ID solicitado.
	ErrClienteNotFound = errors.New("cliente no encontrado")

	// ErrEmailDuplicated indica que ya existe un cliente registrado con ese email.
	ErrEmailDuplicated = errors.New("el email ya esta registrado por otro cliente")

	// ErrNombreRequerido indica que el nombre del cliente no puede estar vacio.
	ErrNombreRequerido = errors.New("el nombre del cliente es requerido")

	// ErrEmailInvalido indica que el formato del email no es valido.
	ErrEmailInvalido = errors.New("el formato del email no es valido")
)
