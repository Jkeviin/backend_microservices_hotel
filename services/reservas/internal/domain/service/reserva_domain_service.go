package service

import (
	"context"
	"time"

	"reservas/internal/domain/model"
	"reservas/internal/domain/repository"
)

// HabitacionChecker permite al dominio verificar el estado de una habitacion
// sin acoplar al modelo de otro bounded context.
type HabitacionChecker interface {
	// GetEstadoHabitacion retorna el estado de la habitacion (e.g. "disponible",
	// "mantenimiento"). Retorna error si la habitacion no existe.
	GetEstadoHabitacion(ctx context.Context, id int) (string, error)
}

// ReservaDomainService encapsula reglas de negocio que requieren acceso al
// repositorio (solapamiento, disponibilidad de habitacion).
type ReservaDomainService struct {
	repo             repository.ReservaRepository
	habitacionCheck  HabitacionChecker
}

// NewReservaDomainService crea un nuevo domain service.
func NewReservaDomainService(
	repo repository.ReservaRepository,
	habitacionCheck HabitacionChecker,
) *ReservaDomainService {
	return &ReservaDomainService{
		repo:            repo,
		habitacionCheck: habitacionCheck,
	}
}

// ValidarDisponibilidad verifica que no haya solapamiento de fechas para la
// habitacion. excludeReservaID permite excluir una reserva (para reprogramacion).
func (s *ReservaDomainService) ValidarDisponibilidad(
	ctx context.Context,
	idHabitacion int,
	inicio, fin time.Time,
	excludeReservaID *int,
) error {
	solapa, err := s.repo.ExisteSolapamiento(ctx, idHabitacion, inicio, fin, excludeReservaID)
	if err != nil {
		return err
	}
	if solapa {
		return model.ErrConflictoFechas
	}
	return nil
}

// ValidarFechas verifica que fin > inicio.
func (s *ReservaDomainService) ValidarFechas(inicio, fin time.Time) error {
	if !fin.After(inicio) {
		return model.ErrFechasInvalidas
	}
	return nil
}

// ValidarHabitacionDisponible verifica que la habitacion no este en
// mantenimiento.
func (s *ReservaDomainService) ValidarHabitacionDisponible(ctx context.Context, idHabitacion int) error {
	estado, err := s.habitacionCheck.GetEstadoHabitacion(ctx, idHabitacion)
	if err != nil {
		return err
	}
	if estado == "mantenimiento" {
		return model.ErrHabitacionNoDisponible
	}
	return nil
}
