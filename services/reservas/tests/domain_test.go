package tests

import (
	"testing"
	"time"

	"reservas/internal/domain/model"
)

// ---------------------------------------------------------------------------
// Value Object: FechaReserva
// ---------------------------------------------------------------------------

func TestNewFechaReserva_Valid(t *testing.T) {
	inicio := time.Now().AddDate(0, 1, 0)
	fin := inicio.AddDate(0, 0, 3)

	fr, err := model.NewFechaReserva(inicio, fin, true)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !fr.Fin.After(fr.Inicio) {
		t.Error("Fin deberia ser posterior a Inicio")
	}
}

func TestNewFechaReserva_FinBeforeInicio(t *testing.T) {
	inicio := time.Now().AddDate(0, 1, 0)
	fin := inicio.AddDate(0, 0, -1)

	_, err := model.NewFechaReserva(inicio, fin, false)
	if err != model.ErrFechasInvalidas {
		t.Errorf("error = %v, esperado ErrFechasInvalidas", err)
	}
}

func TestNewFechaReserva_FinEqualsInicio(t *testing.T) {
	inicio := time.Now().AddDate(0, 1, 0)

	_, err := model.NewFechaReserva(inicio, inicio, false)
	if err != model.ErrFechasInvalidas {
		t.Errorf("error = %v, esperado ErrFechasInvalidas", err)
	}
}

func TestNewFechaReserva_PastDate_RequireFuture(t *testing.T) {
	inicio := time.Now().AddDate(0, 0, -5)
	fin := inicio.AddDate(0, 0, 3)

	_, err := model.NewFechaReserva(inicio, fin, true)
	if err != model.ErrFechasInvalidas {
		t.Errorf("error = %v, esperado ErrFechasInvalidas", err)
	}
}

func TestNewFechaReserva_PastDate_NoRequireFuture(t *testing.T) {
	inicio := time.Now().AddDate(0, 0, -5)
	fin := inicio.AddDate(0, 0, 3)

	_, err := model.NewFechaReserva(inicio, fin, false)
	if err != nil {
		t.Fatalf("error inesperado: %v (no se requiere futuro)", err)
	}
}

// ---------------------------------------------------------------------------
// Aggregate Root: Reserva - PuedeSerCancelada
// ---------------------------------------------------------------------------

func TestReserva_PuedeSerCancelada_PendientePago(t *testing.T) {
	r := &model.Reserva{NombreEstado: model.EstadoPendientePago}
	if !r.PuedeSerCancelada() {
		t.Error("reserva pendiente_pago deberia poder cancelarse")
	}
}

func TestReserva_PuedeSerCancelada_Confirmada(t *testing.T) {
	r := &model.Reserva{NombreEstado: model.EstadoConfirmada}
	if !r.PuedeSerCancelada() {
		t.Error("reserva confirmada deberia poder cancelarse")
	}
}

func TestReserva_PuedeSerCancelada_YaCancelada(t *testing.T) {
	r := &model.Reserva{NombreEstado: model.EstadoCancelada}
	if r.PuedeSerCancelada() {
		t.Error("reserva cancelada NO deberia poder cancelarse")
	}
}

func TestReserva_PuedeSerCancelada_Completada(t *testing.T) {
	r := &model.Reserva{NombreEstado: model.EstadoCompletada}
	if r.PuedeSerCancelada() {
		t.Error("reserva completada NO deberia poder cancelarse")
	}
}

// ---------------------------------------------------------------------------
// Aggregate Root: Reserva - PuedeSerReprogramada
// ---------------------------------------------------------------------------

func TestReserva_PuedeSerReprogramada_FechaFutura(t *testing.T) {
	r := &model.Reserva{
		NombreEstado: model.EstadoConfirmada,
		FechaInicio:  time.Now().AddDate(0, 1, 0),
	}
	if !r.PuedeSerReprogramada() {
		t.Error("reserva con fecha futura y confirmada deberia poder reprogramarse")
	}
}

func TestReserva_PuedeSerReprogramada_FechaPasada(t *testing.T) {
	r := &model.Reserva{
		NombreEstado: model.EstadoConfirmada,
		FechaInicio:  time.Now().AddDate(0, 0, -5),
	}
	if r.PuedeSerReprogramada() {
		t.Error("reserva con fecha pasada NO deberia poder reprogramarse")
	}
}

func TestReserva_PuedeSerReprogramada_Cancelada(t *testing.T) {
	r := &model.Reserva{
		NombreEstado: model.EstadoCancelada,
		FechaInicio:  time.Now().AddDate(0, 1, 0),
	}
	if r.PuedeSerReprogramada() {
		t.Error("reserva cancelada NO deberia poder reprogramarse")
	}
}

// ---------------------------------------------------------------------------
// Aggregate Root: Reserva - Cancelar
// ---------------------------------------------------------------------------

func TestReserva_Cancelar_Success(t *testing.T) {
	r := &model.Reserva{
		IDEstado:     1,
		NombreEstado: model.EstadoPendientePago,
	}
	err := r.Cancelar(3)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if r.NombreEstado != model.EstadoCancelada {
		t.Errorf("NombreEstado = %q, esperado %q", r.NombreEstado, model.EstadoCancelada)
	}
	if r.IDEstado != 3 {
		t.Errorf("IDEstado = %d, esperado 3", r.IDEstado)
	}
	if r.UpdatedAt == nil {
		t.Error("UpdatedAt deberia estar seteado")
	}
}

func TestReserva_Cancelar_YaCancelada(t *testing.T) {
	r := &model.Reserva{NombreEstado: model.EstadoCancelada}
	err := r.Cancelar(3)
	if err != model.ErrEstadoNoPermiteCambio {
		t.Errorf("error = %v, esperado ErrEstadoNoPermiteCambio", err)
	}
}

// ---------------------------------------------------------------------------
// Aggregate Root: Reserva - Reprogramar
// ---------------------------------------------------------------------------

func TestReserva_Reprogramar_Success(t *testing.T) {
	nuevoInicio := time.Now().AddDate(0, 2, 0)
	nuevoFin := nuevoInicio.AddDate(0, 0, 3)

	r := &model.Reserva{
		NombreEstado: model.EstadoConfirmada,
		FechaInicio:  time.Now().AddDate(0, 1, 0),
		FechaFin:     time.Now().AddDate(0, 1, 3),
	}
	err := r.Reprogramar(nuevoInicio, nuevoFin)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if r.UpdatedAt == nil {
		t.Error("UpdatedAt deberia estar seteado")
	}
}

func TestReserva_Reprogramar_EstadoNoPermite(t *testing.T) {
	r := &model.Reserva{
		NombreEstado: model.EstadoCancelada,
		FechaInicio:  time.Now().AddDate(0, 1, 0),
	}
	err := r.Reprogramar(time.Now().AddDate(0, 2, 0), time.Now().AddDate(0, 2, 3))
	if err != model.ErrEstadoNoPermiteCambio {
		t.Errorf("error = %v, esperado ErrEstadoNoPermiteCambio", err)
	}
}

func TestReserva_Reprogramar_YaIniciada(t *testing.T) {
	r := &model.Reserva{
		NombreEstado: model.EstadoConfirmada,
		FechaInicio:  time.Now().AddDate(0, 0, -2),
	}
	err := r.Reprogramar(time.Now().AddDate(0, 2, 0), time.Now().AddDate(0, 2, 3))
	if err != model.ErrReservaYaIniciada {
		t.Errorf("error = %v, esperado ErrReservaYaIniciada", err)
	}
}
