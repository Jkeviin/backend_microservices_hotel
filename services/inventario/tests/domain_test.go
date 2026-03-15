package tests

import (
	"testing"
	"time"

	"inventario/internal/domain/model"
)

// ---------------------------------------------------------------------------
// Value Object: NumeroHabitacion
// ---------------------------------------------------------------------------

func TestNumeroHabitacion_Valido(t *testing.T) {
	casos := []string{"101", "A-1", "PENT-01", "1234567890"}
	for _, c := range casos {
		n, err := model.NewNumeroHabitacion(c)
		if err != nil {
			t.Errorf("NewNumeroHabitacion(%q) error inesperado: %v", c, err)
		}
		if n.String() != c {
			t.Errorf("esperado %q, obtenido %q", c, n.String())
		}
	}
}

func TestNumeroHabitacion_Vacio(t *testing.T) {
	_, err := model.NewNumeroHabitacion("")
	if err != model.ErrNumeroVacio {
		t.Errorf("esperado ErrNumeroVacio, obtenido %v", err)
	}

	_, err = model.NewNumeroHabitacion("   ")
	if err != model.ErrNumeroVacio {
		t.Errorf("esperado ErrNumeroVacio para espacios, obtenido %v", err)
	}
}

func TestNumeroHabitacion_MuyLargo(t *testing.T) {
	_, err := model.NewNumeroHabitacion("12345678901") // 11 chars
	if err != model.ErrNumeroMuyLargo {
		t.Errorf("esperado ErrNumeroMuyLargo, obtenido %v", err)
	}
}

// ---------------------------------------------------------------------------
// Value Object: EstadoHabitacion
// ---------------------------------------------------------------------------

func TestEstadoHabitacion_Valido(t *testing.T) {
	casos := []struct {
		input    string
		esperado string
	}{
		{"disponible", "disponible"},
		{"Disponible", "disponible"},
		{"MANTENIMIENTO", "mantenimiento"},
		{"ocupada", "ocupada"},
		{" Ocupada ", "ocupada"},
	}
	for _, c := range casos {
		e, err := model.NewEstadoHabitacion(c.input)
		if err != nil {
			t.Errorf("NewEstadoHabitacion(%q) error inesperado: %v", c.input, err)
		}
		if e.String() != c.esperado {
			t.Errorf("esperado %q, obtenido %q", c.esperado, e.String())
		}
		if !e.IsValid() {
			t.Errorf("IsValid() deberia retornar true para %q", c.input)
		}
	}
}

func TestEstadoHabitacion_Invalido(t *testing.T) {
	invalidos := []string{"", "inexistente", "libre", "cerrada"}
	for _, v := range invalidos {
		_, err := model.NewEstadoHabitacion(v)
		if err != model.ErrEstadoInvalido {
			t.Errorf("NewEstadoHabitacion(%q): esperado ErrEstadoInvalido, obtenido %v", v, err)
		}
	}
}

func TestEstadoHabitacion_IsValid_Zero(t *testing.T) {
	var e model.EstadoHabitacion
	if e.IsValid() {
		t.Error("IsValid() deberia retornar false para zero value")
	}
}

// ---------------------------------------------------------------------------
// Aggregate: Habitacion - PuedeIrAMantenimiento
// ---------------------------------------------------------------------------

func TestPuedeIrAMantenimiento_Disponible(t *testing.T) {
	h := &model.Habitacion{Estado: model.EstadoDisponible}
	if !h.PuedeIrAMantenimiento() {
		t.Error("habitacion disponible deberia poder ir a mantenimiento")
	}
}

func TestPuedeIrAMantenimiento_Mantenimiento(t *testing.T) {
	h := &model.Habitacion{Estado: model.EstadoMantenimiento}
	if !h.PuedeIrAMantenimiento() {
		t.Error("habitacion en mantenimiento deberia poder ir a mantenimiento (no-op)")
	}
}

func TestPuedeIrAMantenimiento_Ocupada(t *testing.T) {
	h := &model.Habitacion{Estado: model.EstadoOcupada}
	if h.PuedeIrAMantenimiento() {
		t.Error("habitacion ocupada NO deberia poder ir a mantenimiento")
	}
}

// ---------------------------------------------------------------------------
// Filtros: Validacion
// ---------------------------------------------------------------------------

func TestFiltros_SinFechas_OK(t *testing.T) {
	f := model.HabitacionFiltros{}
	if err := f.Validar(); err != nil {
		t.Errorf("filtros sin fechas no deberian dar error: %v", err)
	}
}

func TestFiltros_AmbasFechas_OK(t *testing.T) {
	desde := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	hasta := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	f := model.HabitacionFiltros{DisponibleDesde: &desde, DisponibleHasta: &hasta}
	if err := f.Validar(); err != nil {
		t.Errorf("filtros con ambas fechas validas no deberian dar error: %v", err)
	}
}

func TestFiltros_SoloDesde_Error(t *testing.T) {
	desde := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	f := model.HabitacionFiltros{DisponibleDesde: &desde}
	if err := f.Validar(); err != model.ErrFiltroFechasIncompletas {
		t.Errorf("esperado ErrFiltroFechasIncompletas, obtenido %v", err)
	}
}

func TestFiltros_SoloHasta_Error(t *testing.T) {
	hasta := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	f := model.HabitacionFiltros{DisponibleHasta: &hasta}
	if err := f.Validar(); err != model.ErrFiltroFechasIncompletas {
		t.Errorf("esperado ErrFiltroFechasIncompletas, obtenido %v", err)
	}
}

func TestFiltros_HastaAntesQueDesde_Error(t *testing.T) {
	desde := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	hasta := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	f := model.HabitacionFiltros{DisponibleDesde: &desde, DisponibleHasta: &hasta}
	if err := f.Validar(); err != model.ErrFiltroFechasInvalidas {
		t.Errorf("esperado ErrFiltroFechasInvalidas, obtenido %v", err)
	}
}

func TestFiltros_FechasIguales_Error(t *testing.T) {
	fecha := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	f := model.HabitacionFiltros{DisponibleDesde: &fecha, DisponibleHasta: &fecha}
	if err := f.Validar(); err != model.ErrFiltroFechasInvalidas {
		t.Errorf("esperado ErrFiltroFechasInvalidas para fechas iguales, obtenido %v", err)
	}
}
