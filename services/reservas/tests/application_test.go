package tests

import (
	"context"
	"testing"
	"time"

	"reservas/internal/application"
	"reservas/internal/domain/model"
	domainservice "reservas/internal/domain/service"
	"reservas/internal/infrastructure/persistence/memory"
)

// newTestAppService crea un application service con repos in-memory frescos.
func newTestAppService() *application.ReservaAppService {
	reservaRepo := memory.NewReservaRepo()
	estadoRepo := memory.NewEstadoRepo()
	habChecker := memory.NewHabitacionChecker()
	domainSvc := domainservice.NewReservaDomainService(reservaRepo, habChecker)
	return application.NewReservaAppService(reservaRepo, estadoRepo, domainSvc)
}

func futureDate(daysFromNow int) string {
	return time.Now().AddDate(0, 0, daysFromNow).Format("2006-01-02")
}

// ---------------------------------------------------------------------------
// CreateReserva
// ---------------------------------------------------------------------------

func TestCreateReserva_Success(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	resp, err := svc.CreateReserva(ctx, application.CreateReservaCommand{
		IDCliente:      1,
		IDHabitacion:   4, // disponible, sin reservas previas
		FechaInicio:    futureDate(10),
		FechaFin:       futureDate(13),
		IdempotencyKey: "create-001",
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.ID == 0 {
		t.Error("ID no deberia ser 0")
	}
	if resp.Estado != model.EstadoPendientePago {
		t.Errorf("Estado = %q, esperado %q", resp.Estado, model.EstadoPendientePago)
	}
	if resp.Version != 1 {
		t.Errorf("Version = %d, esperado 1", resp.Version)
	}
}

func TestCreateReserva_Idempotente(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	cmd := application.CreateReservaCommand{
		IDCliente:      1,
		IDHabitacion:   4,
		FechaInicio:    futureDate(10),
		FechaFin:       futureDate(13),
		IdempotencyKey: "idem-001",
	}

	resp1, err := svc.CreateReserva(ctx, cmd)
	if err != nil {
		t.Fatalf("error en primer create: %v", err)
	}

	// Segunda llamada con misma idempotency key debe retornar la misma reserva.
	resp2, err := svc.CreateReserva(ctx, cmd)
	if err != nil {
		t.Fatalf("error en segundo create: %v", err)
	}

	if resp1.ID != resp2.ID {
		t.Errorf("IDs distintos: %d vs %d, deberian ser iguales por idempotencia", resp1.ID, resp2.ID)
	}
}

func TestCreateReserva_Solapamiento(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	// Crear primera reserva.
	_, err := svc.CreateReserva(ctx, application.CreateReservaCommand{
		IDCliente:      1,
		IDHabitacion:   5,
		FechaInicio:    futureDate(20),
		FechaFin:       futureDate(25),
		IdempotencyKey: "solap-001",
	})
	if err != nil {
		t.Fatalf("error creando primera reserva: %v", err)
	}

	// Intentar crear segunda reserva solapada en misma habitacion.
	_, err = svc.CreateReserva(ctx, application.CreateReservaCommand{
		IDCliente:      2,
		IDHabitacion:   5,
		FechaInicio:    futureDate(22),
		FechaFin:       futureDate(27),
		IdempotencyKey: "solap-002",
	})
	if err != model.ErrConflictoFechas {
		t.Errorf("error = %v, esperado ErrConflictoFechas", err)
	}
}

func TestCreateReserva_HabitacionMantenimiento(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	_, err := svc.CreateReserva(ctx, application.CreateReservaCommand{
		IDCliente:      1,
		IDHabitacion:   6, // en mantenimiento
		FechaInicio:    futureDate(10),
		FechaFin:       futureDate(13),
		IdempotencyKey: "mant-001",
	})
	if err != model.ErrHabitacionNoDisponible {
		t.Errorf("error = %v, esperado ErrHabitacionNoDisponible", err)
	}
}

func TestCreateReserva_FechasInvalidas(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	// Fin antes de inicio.
	_, err := svc.CreateReserva(ctx, application.CreateReservaCommand{
		IDCliente:      1,
		IDHabitacion:   4,
		FechaInicio:    futureDate(15),
		FechaFin:       futureDate(10),
		IdempotencyKey: "fechas-001",
	})
	if err != model.ErrFechasInvalidas {
		t.Errorf("error = %v, esperado ErrFechasInvalidas", err)
	}
}

// ---------------------------------------------------------------------------
// CancelarReserva
// ---------------------------------------------------------------------------

func TestCancelarReserva_Success(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	// Reserva 2 en datos semilla: pendiente_pago, version 1.
	resp, err := svc.CancelarReserva(ctx, 2, 1)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.Estado != model.EstadoCancelada {
		t.Errorf("Estado = %q, esperado %q", resp.Estado, model.EstadoCancelada)
	}
}

func TestCancelarReserva_YaCancelada(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	// Reserva 3 en datos semilla: cancelada, version 2.
	_, err := svc.CancelarReserva(ctx, 3, 2)
	if err != model.ErrEstadoNoPermiteCambio {
		t.Errorf("error = %v, esperado ErrEstadoNoPermiteCambio", err)
	}
}

func TestCancelarReserva_VersionConflict(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	// Reserva 2 tiene version 1, enviamos version 99.
	_, err := svc.CancelarReserva(ctx, 2, 99)
	if err != model.ErrVersionConflict {
		t.Errorf("error = %v, esperado ErrVersionConflict", err)
	}
}

// ---------------------------------------------------------------------------
// ReprogramarReserva
// ---------------------------------------------------------------------------

func TestReprogramarReserva_Success(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	// Crear una reserva fresca con fecha futura para poder reprogramar.
	resp, err := svc.CreateReserva(ctx, application.CreateReservaCommand{
		IDCliente:      1,
		IDHabitacion:   4,
		FechaInicio:    futureDate(30),
		FechaFin:       futureDate(33),
		IdempotencyKey: "reprog-001",
	})
	if err != nil {
		t.Fatalf("error creando reserva: %v", err)
	}

	// Reprogramar a nuevas fechas.
	respR, err := svc.ReprogramarReserva(ctx, resp.ID, futureDate(35), futureDate(38), resp.Version)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if respR.FechaInicio != futureDate(35) {
		t.Errorf("FechaInicio = %q, esperado %q", respR.FechaInicio, futureDate(35))
	}
}

func TestReprogramarReserva_YaIniciada(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	// Reserva 1 en datos semilla: confirmada, inicio 2025-03-15 (pasado).
	_, err := svc.ReprogramarReserva(ctx, 1, futureDate(40), futureDate(43), 1)
	if err != model.ErrReservaYaIniciada {
		t.Errorf("error = %v, esperado ErrReservaYaIniciada", err)
	}
}

// ---------------------------------------------------------------------------
// GetReserva
// ---------------------------------------------------------------------------

func TestGetReserva_Existente(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	resp, err := svc.GetReserva(ctx, 1)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.IDCliente != 1 {
		t.Errorf("IDCliente = %d, esperado 1", resp.IDCliente)
	}
	if resp.Estado != model.EstadoConfirmada {
		t.Errorf("Estado = %q, esperado %q", resp.Estado, model.EstadoConfirmada)
	}
}

func TestGetReserva_NoExistente(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	_, err := svc.GetReserva(ctx, 999)
	if err != model.ErrReservaNotFound {
		t.Errorf("error = %v, esperado ErrReservaNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// ListEstados
// ---------------------------------------------------------------------------

func TestListEstados(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	estados, err := svc.ListEstados(ctx)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(estados) != 4 {
		t.Errorf("len(estados) = %d, esperado 4", len(estados))
	}

	nombres := map[string]bool{}
	for _, e := range estados {
		nombres[e.Nombre] = true
	}
	for _, expected := range []string{model.EstadoPendientePago, model.EstadoConfirmada, model.EstadoCancelada, model.EstadoCompletada} {
		if !nombres[expected] {
			t.Errorf("estado %q no encontrado en la lista", expected)
		}
	}
}
