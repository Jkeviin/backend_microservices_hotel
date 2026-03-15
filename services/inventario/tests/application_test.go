package tests

import (
	"context"
	"testing"

	"inventario/internal/application"
	"inventario/internal/domain/model"
	"inventario/internal/infrastructure/persistence/memory"
)

func newTestApp() *application.InventarioApp {
	return application.NewInventarioApp(
		memory.NewHabitacionRepo(),
		memory.NewTipoHabitacionRepo(),
	)
}

// ---------------------------------------------------------------------------
// ListHabitaciones
// ---------------------------------------------------------------------------

func TestListHabitaciones_SinFiltros(t *testing.T) {
	app := newTestApp()
	result, err := app.ListHabitaciones(context.Background(), model.HabitacionFiltros{})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("se esperaban habitaciones, obtenido 0")
	}
	// El repo memory tiene 9 habitaciones pre-cargadas.
	if len(result) != 9 {
		t.Errorf("esperado 9 habitaciones, obtenido %d", len(result))
	}
}

func TestListHabitaciones_FiltroTipo(t *testing.T) {
	app := newTestApp()
	tipo := 1 // Simple
	result, err := app.ListHabitaciones(context.Background(), model.HabitacionFiltros{Tipo: &tipo})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	// Habitaciones tipo 1 (Simple): 101, 102, 402 = 3
	if len(result) != 3 {
		t.Errorf("esperado 3 habitaciones tipo Simple, obtenido %d", len(result))
	}
	for _, h := range result {
		if h.IDTipoHabitacion != 1 {
			t.Errorf("habitacion %s tiene tipo %d, esperado 1", h.Numero, h.IDTipoHabitacion)
		}
	}
}

func TestListHabitaciones_FiltroEstado(t *testing.T) {
	app := newTestApp()
	estado := model.EstadoMantenimiento
	result, err := app.ListHabitaciones(context.Background(), model.HabitacionFiltros{Estado: &estado})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	// Solo habitacion 302 (ID 6) esta en mantenimiento.
	if len(result) != 1 {
		t.Errorf("esperado 1 habitacion en mantenimiento, obtenido %d", len(result))
	}
}

func TestListHabitaciones_FiltroDisponibilidad(t *testing.T) {
	app := newTestApp()

	// Las reservas mock excluyen ciertas habitaciones segun el rango.
	// Usar un rango lejano en el futuro donde no hay reservas mock.
	desde := mustParseDate("2027-06-01")
	hasta := mustParseDate("2027-06-05")
	result, err := app.ListHabitaciones(context.Background(), model.HabitacionFiltros{
		DisponibleDesde: &desde,
		DisponibleHasta: &hasta,
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	// Sin reservas en ese rango, deberian retornar todas (9).
	if len(result) != 9 {
		t.Errorf("esperado 9 habitaciones disponibles en rango futuro, obtenido %d", len(result))
	}
}

func TestListHabitaciones_FiltroFechasIncompletas(t *testing.T) {
	app := newTestApp()
	desde := mustParseDate("2026-03-01")
	_, err := app.ListHabitaciones(context.Background(), model.HabitacionFiltros{
		DisponibleDesde: &desde,
	})
	if err != model.ErrFiltroFechasIncompletas {
		t.Errorf("esperado ErrFiltroFechasIncompletas, obtenido %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetHabitacion
// ---------------------------------------------------------------------------

func TestGetHabitacion_Existente(t *testing.T) {
	app := newTestApp()
	result, err := app.GetHabitacion(context.Background(), 1)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("esperado ID 1, obtenido %d", result.ID)
	}
	if result.Numero != "101" {
		t.Errorf("esperado numero 101, obtenido %s", result.Numero)
	}
}

func TestGetHabitacion_NoExistente(t *testing.T) {
	app := newTestApp()
	_, err := app.GetHabitacion(context.Background(), 999)
	if err != model.ErrHabitacionNotFound {
		t.Errorf("esperado ErrHabitacionNotFound, obtenido %v", err)
	}
}

// ---------------------------------------------------------------------------
// UpdateHabitacion
// ---------------------------------------------------------------------------

func TestUpdateHabitacion_CambioEstado_Valido(t *testing.T) {
	app := newTestApp()

	// Habitacion 1 (101) esta disponible. Cambiar a mantenimiento.
	estado := "mantenimiento"
	result, err := app.UpdateHabitacion(context.Background(), 1, application.UpdateHabitacionCommand{
		Estado: &estado,
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if result.Estado != "mantenimiento" {
		t.Errorf("esperado estado mantenimiento, obtenido %s", result.Estado)
	}
	if result.UpdatedAt == nil {
		t.Error("UpdatedAt deberia estar seteado")
	}
}

func TestUpdateHabitacion_CambioTipo(t *testing.T) {
	app := newTestApp()

	nuevoTipo := 3
	result, err := app.UpdateHabitacion(context.Background(), 1, application.UpdateHabitacionCommand{
		IDTipo: &nuevoTipo,
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if result.IDTipoHabitacion != 3 {
		t.Errorf("esperado tipo 3, obtenido %d", result.IDTipoHabitacion)
	}
}

func TestUpdateHabitacion_NoMantenimientoSiOcupada(t *testing.T) {
	app := newTestApp()

	// Habitacion 4 (202) esta ocupada. Intentar cambiar a mantenimiento.
	estado := "mantenimiento"
	_, err := app.UpdateHabitacion(context.Background(), 4, application.UpdateHabitacionCommand{
		Estado: &estado,
	})
	if err != model.ErrNoMantenimientoSiOcupada {
		t.Errorf("esperado ErrNoMantenimientoSiOcupada, obtenido %v", err)
	}
}

func TestUpdateHabitacion_EstadoInvalido(t *testing.T) {
	app := newTestApp()

	estado := "inexistente"
	_, err := app.UpdateHabitacion(context.Background(), 1, application.UpdateHabitacionCommand{
		Estado: &estado,
	})
	if err != model.ErrEstadoInvalido {
		t.Errorf("esperado ErrEstadoInvalido, obtenido %v", err)
	}
}

func TestUpdateHabitacion_NoExistente(t *testing.T) {
	app := newTestApp()

	estado := "disponible"
	_, err := app.UpdateHabitacion(context.Background(), 999, application.UpdateHabitacionCommand{
		Estado: &estado,
	})
	if err != model.ErrHabitacionNotFound {
		t.Errorf("esperado ErrHabitacionNotFound, obtenido %v", err)
	}
}

// ---------------------------------------------------------------------------
// ListTiposHabitacion
// ---------------------------------------------------------------------------

func TestListTiposHabitacion(t *testing.T) {
	app := newTestApp()
	result, err := app.ListTiposHabitacion(context.Background())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("esperado 3 tipos, obtenido %d", len(result))
	}
}
