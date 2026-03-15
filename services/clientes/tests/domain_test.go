package tests

import (
	"testing"

	"clientes/internal/domain/model"
)

// ---------------------------------------------------------------------------
// Value Object: Email
// ---------------------------------------------------------------------------

func TestNewEmail_Valid(t *testing.T) {
	cases := []string{
		"user@example.com",
		"test.name@domain.org",
		"a@b.co",
		"juan.perez@email.com",
	}
	for _, tc := range cases {
		email, err := model.NewEmail(tc)
		if err != nil {
			t.Errorf("NewEmail(%q) retorno error inesperado: %v", tc, err)
		}
		if email.String() != tc {
			t.Errorf("NewEmail(%q).String() = %q, esperado %q", tc, email.String(), tc)
		}
	}
}

func TestNewEmail_Invalid(t *testing.T) {
	cases := []string{
		"",
		"noarroba",
		"@domain.com",
		"user@",
		"user@domain",
		"user@@domain.com",
		"user@domain.",
	}
	for _, tc := range cases {
		_, err := model.NewEmail(tc)
		if err == nil {
			t.Errorf("NewEmail(%q) no retorno error, esperado ErrEmailInvalido", tc)
		}
		if err != model.ErrEmailInvalido {
			t.Errorf("NewEmail(%q) retorno %v, esperado ErrEmailInvalido", tc, err)
		}
	}
}

func TestEmail_String(t *testing.T) {
	email, _ := model.NewEmail("test@example.com")
	if email.String() != "test@example.com" {
		t.Errorf("String() = %q, esperado %q", email.String(), "test@example.com")
	}
}

// ---------------------------------------------------------------------------
// Aggregate Root: Cliente
// ---------------------------------------------------------------------------

func TestNewCliente_Success(t *testing.T) {
	email, _ := model.NewEmail("test@example.com")
	tel := "3001234567"

	c, err := model.NewCliente("Juan Perez", email, &tel)
	if err != nil {
		t.Fatalf("NewCliente retorno error inesperado: %v", err)
	}
	if c.Nombre != "Juan Perez" {
		t.Errorf("Nombre = %q, esperado %q", c.Nombre, "Juan Perez")
	}
	if c.Email.String() != "test@example.com" {
		t.Errorf("Email = %q, esperado %q", c.Email.String(), "test@example.com")
	}
	if c.Telefono == nil || *c.Telefono != "3001234567" {
		t.Errorf("Telefono = %v, esperado %q", c.Telefono, "3001234567")
	}
	if c.CreatedAt.IsZero() {
		t.Error("CreatedAt no deberia ser zero")
	}
}

func TestNewCliente_SinTelefono(t *testing.T) {
	email, _ := model.NewEmail("test@example.com")

	c, err := model.NewCliente("Ana Rodriguez", email, nil)
	if err != nil {
		t.Fatalf("NewCliente retorno error inesperado: %v", err)
	}
	if c.Telefono != nil {
		t.Errorf("Telefono deberia ser nil, got %v", c.Telefono)
	}
}

func TestNewCliente_NombreVacio(t *testing.T) {
	email, _ := model.NewEmail("test@example.com")

	_, err := model.NewCliente("", email, nil)
	if err != model.ErrNombreRequerido {
		t.Errorf("error = %v, esperado ErrNombreRequerido", err)
	}
}

func TestNewCliente_NombreSoloEspacios(t *testing.T) {
	email, _ := model.NewEmail("test@example.com")

	_, err := model.NewCliente("   ", email, nil)
	if err != model.ErrNombreRequerido {
		t.Errorf("error = %v, esperado ErrNombreRequerido", err)
	}
}

func TestNewCliente_NombreConEspacios_TrimsCorrectamente(t *testing.T) {
	email, _ := model.NewEmail("test@example.com")

	c, err := model.NewCliente("  Juan Perez  ", email, nil)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if c.Nombre != "Juan Perez" {
		t.Errorf("Nombre = %q, esperado %q", c.Nombre, "Juan Perez")
	}
}
