package application

import (
	"context"
	"time"

	"reservas/internal/domain/model"
	"reservas/internal/domain/repository"
	domainservice "reservas/internal/domain/service"
)

// ---------------------------------------------------------------------------
// Commands (entrada)
// ---------------------------------------------------------------------------

// CreateReservaCommand contiene los datos necesarios para crear una reserva.
type CreateReservaCommand struct {
	IDCliente      int    `json:"id_cliente"`
	IDHabitacion   int    `json:"id_habitacion"`
	FechaInicio    string `json:"fecha_inicio"`
	FechaFin       string `json:"fecha_fin"`
	IdempotencyKey string `json:"-"` // viene del header, no del body
}

// PatchReservaCommand contiene la accion y datos para modificar una reserva.
type PatchReservaCommand struct {
	Accion     string `json:"accion,omitempty"`
	FechaInicio string `json:"fecha_inicio,omitempty"`
	FechaFin    string `json:"fecha_fin,omitempty"`
	Version    int    `json:"version"`
}

// ---------------------------------------------------------------------------
// Response DTO (salida)
// ---------------------------------------------------------------------------

// ReservaResponse es el DTO de respuesta que se expone al exterior.
type ReservaResponse struct {
	ID           int        `json:"id"`
	IDCliente    int        `json:"id_cliente"`
	IDHabitacion int        `json:"id_habitacion"`
	Estado       string     `json:"estado"`
	FechaInicio  string     `json:"fecha_inicio"`
	FechaFin     string     `json:"fecha_fin"`
	Total        *float64   `json:"total,omitempty"`
	Version      int        `json:"version"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

// EstadoResponse es el DTO para listar estados.
type EstadoResponse struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

const dateLayout = "2006-01-02"

func toResponse(r *model.Reserva) *ReservaResponse {
	return &ReservaResponse{
		ID:           r.ID,
		IDCliente:    r.IDCliente,
		IDHabitacion: r.IDHabitacion,
		Estado:       r.NombreEstado,
		FechaInicio:  r.FechaInicio.Format(dateLayout),
		FechaFin:     r.FechaFin.Format(dateLayout),
		Total:        r.Total,
		Version:      r.Version,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

// ---------------------------------------------------------------------------
// Application Service
// ---------------------------------------------------------------------------

// ReservaAppService orquesta los casos de uso del bounded context de reservas.
type ReservaAppService struct {
	repo      repository.ReservaRepository
	estadoRepo repository.EstadoReservaRepository
	domainSvc *domainservice.ReservaDomainService
}

// NewReservaAppService crea el application service con sus dependencias.
func NewReservaAppService(
	repo repository.ReservaRepository,
	estadoRepo repository.EstadoReservaRepository,
	domainSvc *domainservice.ReservaDomainService,
) *ReservaAppService {
	return &ReservaAppService{
		repo:       repo,
		estadoRepo: estadoRepo,
		domainSvc:  domainSvc,
	}
}

// CreateReserva ejecuta el caso de uso de creacion de reserva.
func (s *ReservaAppService) CreateReserva(ctx context.Context, cmd CreateReservaCommand) (*ReservaResponse, error) {
	// 1. Verificar idempotency_key: si ya existe, retornar la reserva existente.
	if cmd.IdempotencyKey != "" {
		existing, err := s.repo.GetByIdempotencyKey(ctx, cmd.IdempotencyKey)
		if err == nil && existing != nil {
			return toResponse(existing), nil
		}
	}

	// 2. Validar fechas.
	inicio, err := time.Parse(dateLayout, cmd.FechaInicio)
	if err != nil {
		return nil, model.ErrFechasInvalidas
	}
	fin, err := time.Parse(dateLayout, cmd.FechaFin)
	if err != nil {
		return nil, model.ErrFechasInvalidas
	}

	if err := s.domainSvc.ValidarFechas(inicio, fin); err != nil {
		return nil, err
	}

	// 3. Verificar que habitacion no este en mantenimiento.
	if err := s.domainSvc.ValidarHabitacionDisponible(ctx, cmd.IDHabitacion); err != nil {
		return nil, err
	}

	// 4. Check solapamiento.
	if err := s.domainSvc.ValidarDisponibilidad(ctx, cmd.IDHabitacion, inicio, fin, nil); err != nil {
		return nil, err
	}

	// 5. Obtener estado pendiente_pago.
	estado, err := s.estadoRepo.GetByNombre(ctx, model.EstadoPendientePago)
	if err != nil {
		return nil, err
	}

	// 6. Crear reserva.
	var idemKey *string
	if cmd.IdempotencyKey != "" {
		idemKey = &cmd.IdempotencyKey
	}

	reserva := &model.Reserva{
		IDCliente:      cmd.IDCliente,
		IDHabitacion:   cmd.IDHabitacion,
		IDEstado:       estado.ID,
		NombreEstado:   estado.Nombre,
		FechaInicio:    inicio,
		FechaFin:       fin,
		IdempotencyKey: idemKey,
		Version:        1,
		CreatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, reserva); err != nil {
		return nil, err
	}

	return toResponse(reserva), nil
}

// GetReserva obtiene una reserva por ID.
func (s *ReservaAppService) GetReserva(ctx context.Context, id int) (*ReservaResponse, error) {
	reserva, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResponse(reserva), nil
}

// CancelarReserva cancela una reserva existente con optimistic locking.
func (s *ReservaAppService) CancelarReserva(ctx context.Context, id int, version int) (*ReservaResponse, error) {
	reserva, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verificar version para optimistic locking.
	if reserva.Version != version {
		return nil, model.ErrVersionConflict
	}

	// Obtener ID del estado cancelada.
	estado, err := s.estadoRepo.GetByNombre(ctx, model.EstadoCancelada)
	if err != nil {
		return nil, err
	}

	if err := reserva.Cancelar(estado.ID); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, reserva); err != nil {
		return nil, err
	}

	return toResponse(reserva), nil
}

// ReprogramarReserva cambia las fechas de una reserva con optimistic locking.
func (s *ReservaAppService) ReprogramarReserva(
	ctx context.Context,
	id int,
	nuevoInicio, nuevoFin string,
	version int,
) (*ReservaResponse, error) {
	reserva, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verificar version para optimistic locking.
	if reserva.Version != version {
		return nil, model.ErrVersionConflict
	}

	inicio, err := time.Parse(dateLayout, nuevoInicio)
	if err != nil {
		return nil, model.ErrFechasInvalidas
	}
	fin, err := time.Parse(dateLayout, nuevoFin)
	if err != nil {
		return nil, model.ErrFechasInvalidas
	}

	// Validar nuevas fechas en el aggregate.
	if err := reserva.Reprogramar(inicio, fin); err != nil {
		return nil, err
	}

	// Check solapamiento excluyendo la propia reserva.
	excludeID := reserva.ID
	if err := s.domainSvc.ValidarDisponibilidad(ctx, reserva.IDHabitacion, reserva.FechaInicio, reserva.FechaFin, &excludeID); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, reserva); err != nil {
		return nil, err
	}

	return toResponse(reserva), nil
}

// ListEstados retorna todos los estados de reserva.
func (s *ReservaAppService) ListEstados(ctx context.Context) ([]EstadoResponse, error) {
	estados, err := s.estadoRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]EstadoResponse, len(estados))
	for i, e := range estados {
		result[i] = EstadoResponse{ID: e.ID, Nombre: e.Nombre}
	}
	return result, nil
}
