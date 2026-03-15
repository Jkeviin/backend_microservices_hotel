package tests

import (
	"testing"

	"pagos/internal/domain/model"
)

// ---------------------------------------------------------------------------
// Value Object: EstadoPago
// ---------------------------------------------------------------------------

func TestEstadoPago_Valido(t *testing.T) {
	casos := []struct {
		input    string
		esperado string
	}{
		{"pendiente", "pendiente"},
		{"Pendiente", "pendiente"},
		{"APROBADO", "aprobado"},
		{"rechazado", "rechazado"},
		{" Rechazado ", "rechazado"},
	}
	for _, c := range casos {
		e, err := model.NewEstadoPago(c.input)
		if err != nil {
			t.Errorf("NewEstadoPago(%q) error inesperado: %v", c.input, err)
		}
		if e.String() != c.esperado {
			t.Errorf("esperado %q, obtenido %q", c.esperado, e.String())
		}
		if !e.IsValid() {
			t.Errorf("IsValid() deberia retornar true para %q", c.input)
		}
	}
}

func TestEstadoPago_Invalido(t *testing.T) {
	invalidos := []string{"", "inexistente", "cancelado", "completado"}
	for _, v := range invalidos {
		_, err := model.NewEstadoPago(v)
		if err != model.ErrEstadoPagoInvalido {
			t.Errorf("NewEstadoPago(%q): esperado ErrEstadoPagoInvalido, obtenido %v", v, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Value Object: Monto
// ---------------------------------------------------------------------------

func TestMonto_Valido(t *testing.T) {
	montos := []float64{0.01, 1.0, 100.50, 9999.99}
	for _, m := range montos {
		if err := model.ValidarMonto(m); err != nil {
			t.Errorf("ValidarMonto(%v) error inesperado: %v", m, err)
		}
	}
}

func TestMonto_Cero(t *testing.T) {
	if err := model.ValidarMonto(0); err != model.ErrMontoInvalido {
		t.Errorf("ValidarMonto(0): esperado ErrMontoInvalido, obtenido %v", err)
	}
}

func TestMonto_Negativo(t *testing.T) {
	negativos := []float64{-1, -0.01, -100}
	for _, m := range negativos {
		if err := model.ValidarMonto(m); err != model.ErrMontoInvalido {
			t.Errorf("ValidarMonto(%v): esperado ErrMontoInvalido, obtenido %v", m, err)
		}
	}
}
