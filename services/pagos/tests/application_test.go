package tests

import (
	"context"
	"testing"

	"pagos/internal/application"
	"pagos/internal/domain/model"
	"pagos/internal/infrastructure/persistence/memory"
)

func newTestApp() (*application.PagoApp, *memory.ReservaChecker) {
	pagoRepo := memory.NewPagoRepo()
	reservaChecker := memory.NewReservaChecker()
	app := application.NewPagoApp(pagoRepo, reservaChecker)
	return app, reservaChecker
}

// ---------------------------------------------------------------------------
// CreatePago
// ---------------------------------------------------------------------------

func TestCreatePago_Exitoso(t *testing.T) {
	app, _ := newTestApp()
	cmd := application.CreatePagoCommand{
		IDReserva:      2, // pendiente_pago
		Monto:          350.00,
		FechaPago:      "2026-01-15",
		IdempotencyKey: "key-test-1",
	}

	result, err := app.CreatePago(context.Background(), cmd)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if result.IDReserva != 2 {
		t.Errorf("esperado IDReserva 2, obtenido %d", result.IDReserva)
	}
	if result.Monto != 350.00 {
		t.Errorf("esperado monto 350.00, obtenido %f", result.Monto)
	}
	if result.EstadoPago != "aprobado" {
		t.Errorf("esperado estado aprobado, obtenido %s", result.EstadoPago)
	}
	if result.FechaPago != "2026-01-15" {
		t.Errorf("esperado fecha_pago 2026-01-15, obtenido %s", result.FechaPago)
	}
	if result.ID == 0 {
		t.Error("ID no deberia ser 0")
	}
}

func TestCreatePago_Idempotente(t *testing.T) {
	app, _ := newTestApp()
	cmd := application.CreatePagoCommand{
		IDReserva:      2,
		Monto:          200.00,
		FechaPago:      "2026-02-01",
		IdempotencyKey: "key-idempotent",
	}

	// Primera llamada: crea el pago.
	result1, err := app.CreatePago(context.Background(), cmd)
	if err != nil {
		t.Fatalf("primera llamada error: %v", err)
	}

	// Segunda llamada con misma key: retorna el mismo pago.
	result2, err := app.CreatePago(context.Background(), cmd)
	if err != nil {
		t.Fatalf("segunda llamada error: %v", err)
	}

	if result1.ID != result2.ID {
		t.Errorf("idempotencia fallida: ID1=%d, ID2=%d", result1.ID, result2.ID)
	}
	if result1.Monto != result2.Monto {
		t.Errorf("idempotencia fallida: Monto1=%f, Monto2=%f", result1.Monto, result2.Monto)
	}
}

func TestCreatePago_ReservaNoExiste(t *testing.T) {
	app, _ := newTestApp()
	cmd := application.CreatePagoCommand{
		IDReserva:      999,
		Monto:          100.00,
		FechaPago:      "2026-01-01",
		IdempotencyKey: "key-no-reserva",
	}

	_, err := app.CreatePago(context.Background(), cmd)
	if err != model.ErrReservaNoValida {
		t.Errorf("esperado ErrReservaNoValida, obtenido %v", err)
	}
}

func TestCreatePago_ReservaCancelada(t *testing.T) {
	app, _ := newTestApp()
	cmd := application.CreatePagoCommand{
		IDReserva:      3, // cancelada
		Monto:          100.00,
		FechaPago:      "2026-01-01",
		IdempotencyKey: "key-cancelada",
	}

	_, err := app.CreatePago(context.Background(), cmd)
	if err != model.ErrReservaNoValida {
		t.Errorf("esperado ErrReservaNoValida, obtenido %v", err)
	}
}

func TestCreatePago_MontoInvalido(t *testing.T) {
	app, _ := newTestApp()
	cmd := application.CreatePagoCommand{
		IDReserva:      2,
		Monto:          0,
		FechaPago:      "2026-01-01",
		IdempotencyKey: "key-monto-cero",
	}

	_, err := app.CreatePago(context.Background(), cmd)
	if err != model.ErrMontoInvalido {
		t.Errorf("esperado ErrMontoInvalido, obtenido %v", err)
	}
}

func TestCreatePago_ConfirmaReserva(t *testing.T) {
	app, reservaChecker := newTestApp()

	// La reserva 2 esta en pendiente_pago.
	info, _ := reservaChecker.GetReservaParaPago(context.Background(), 2)
	if info.NombreEstado != "pendiente_pago" {
		t.Fatalf("precondicion fallida: reserva 2 deberia estar en pendiente_pago, tiene %s", info.NombreEstado)
	}

	cmd := application.CreatePagoCommand{
		IDReserva:      2,
		Monto:          500.00,
		FechaPago:      "2026-03-01",
		IdempotencyKey: "key-confirma",
	}

	_, err := app.CreatePago(context.Background(), cmd)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	// Verificar que la reserva ahora esta confirmada.
	info, _ = reservaChecker.GetReservaParaPago(context.Background(), 2)
	if info.NombreEstado != "confirmada" {
		t.Errorf("esperado estado confirmada, obtenido %s", info.NombreEstado)
	}
}

// ---------------------------------------------------------------------------
// GetPago
// ---------------------------------------------------------------------------

func TestGetPago_Existente(t *testing.T) {
	app, _ := newTestApp()
	result, err := app.GetPago(context.Background(), 1)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("esperado ID 1, obtenido %d", result.ID)
	}
	if result.Monto != 240.00 {
		t.Errorf("esperado monto 240.00, obtenido %f", result.Monto)
	}
}

func TestGetPago_NoExistente(t *testing.T) {
	app, _ := newTestApp()
	_, err := app.GetPago(context.Background(), 999)
	if err != model.ErrPagoNotFound {
		t.Errorf("esperado ErrPagoNotFound, obtenido %v", err)
	}
}
