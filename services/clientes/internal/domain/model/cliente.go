package model

import (
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Value Object: Email
// ---------------------------------------------------------------------------

// Email representa una direccion de correo electronico validada.
type Email struct {
	value string
}

// NewEmail crea un Email validado. Retorna ErrEmailInvalido si el formato es
// incorrecto (debe contener exactamente un @ y el dominio debe tener al menos
// un punto).
func NewEmail(raw string) (Email, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return Email{}, ErrEmailInvalido
	}

	atIdx := strings.Index(v, "@")
	if atIdx <= 0 || atIdx == len(v)-1 {
		return Email{}, ErrEmailInvalido
	}

	domain := v[atIdx+1:]
	if !strings.Contains(domain, ".") {
		return Email{}, ErrEmailInvalido
	}

	// Verificar que no haya multiples @
	if strings.Count(v, "@") != 1 {
		return Email{}, ErrEmailInvalido
	}

	// Verificar que el dominio no termine en punto
	if strings.HasSuffix(domain, ".") {
		return Email{}, ErrEmailInvalido
	}

	return Email{value: v}, nil
}

// String retorna la representacion textual del email.
func (e Email) String() string {
	return e.value
}

// ---------------------------------------------------------------------------
// Aggregate Root: Cliente
// ---------------------------------------------------------------------------

// Cliente es el aggregate root del bounded context de clientes.
type Cliente struct {
	ID        int
	Nombre    string
	Email     Email
	Telefono  *string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// NewCliente crea una instancia valida de Cliente. Retorna error si el nombre
// esta vacio.
func NewCliente(nombre string, email Email, telefono *string) (*Cliente, error) {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return nil, ErrNombreRequerido
	}

	return &Cliente{
		Nombre:    nombre,
		Email:     email,
		Telefono:  telefono,
		CreatedAt: time.Now(),
	}, nil
}
