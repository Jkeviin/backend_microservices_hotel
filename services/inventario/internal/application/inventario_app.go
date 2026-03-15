package application

import (
	"context"
	"time"

	"inventario/internal/domain/model"
	"inventario/internal/domain/repository"
	domainservice "inventario/internal/domain/service"
)

// ---------------------------------------------------------------------------
// Commands
// ---------------------------------------------------------------------------

// UpdateHabitacionCommand contiene los campos actualizables de una habitacion.
// Los campos nil se ignoran (patch parcial).
type UpdateHabitacionCommand struct {
	Estado *string
	IDTipo *int
}

// ---------------------------------------------------------------------------
// Responses
// ---------------------------------------------------------------------------

// HabitacionResponse es la representacion de una habitacion para la capa de presentacion.
type HabitacionResponse struct {
	ID               int        `json:"id"`
	Numero           string     `json:"numero"`
	IDTipoHabitacion int        `json:"id_tipo_habitacion"`
	Estado           string     `json:"estado"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

// TipoHabitacionResponse es la representacion de un tipo de habitacion para la capa de presentacion.
type TipoHabitacionResponse struct {
	ID         int     `json:"id"`
	Nombre     string  `json:"nombre"`
	PrecioBase float64 `json:"precio_base"`
	Capacidad  int     `json:"capacidad"`
}

// ---------------------------------------------------------------------------
// Application Service
// ---------------------------------------------------------------------------

// InventarioApp orquesta los casos de uso del bounded context de inventario.
type InventarioApp struct {
	habitacionRepo repository.HabitacionRepository
	tipoRepo       repository.TipoHabitacionRepository
}

// NewInventarioApp crea una nueva instancia del servicio de aplicacion.
func NewInventarioApp(
	habitacionRepo repository.HabitacionRepository,
	tipoRepo repository.TipoHabitacionRepository,
) *InventarioApp {
	return &InventarioApp{
		habitacionRepo: habitacionRepo,
		tipoRepo:       tipoRepo,
	}
}

// ListHabitaciones retorna las habitaciones que coincidan con los filtros.
func (a *InventarioApp) ListHabitaciones(ctx context.Context, filtros model.HabitacionFiltros) ([]HabitacionResponse, error) {
	if err := filtros.Validar(); err != nil {
		return nil, err
	}

	habitaciones, err := a.habitacionRepo.List(ctx, filtros)
	if err != nil {
		return nil, err
	}

	result := make([]HabitacionResponse, 0, len(habitaciones))
	for _, h := range habitaciones {
		result = append(result, toHabitacionResponse(h))
	}
	return result, nil
}

// GetHabitacion retorna una habitacion por su ID.
func (a *InventarioApp) GetHabitacion(ctx context.Context, id int) (*HabitacionResponse, error) {
	h, err := a.habitacionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toHabitacionResponse(*h)
	return &resp, nil
}

// UpdateHabitacion aplica cambios parciales a una habitacion existente.
func (a *InventarioApp) UpdateHabitacion(ctx context.Context, id int, cmd UpdateHabitacionCommand) (*HabitacionResponse, error) {
	h, err := a.habitacionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if cmd.Estado != nil {
		nuevoEstado, err := model.NewEstadoHabitacion(*cmd.Estado)
		if err != nil {
			return nil, err
		}
		if err := domainservice.ValidarCambioEstado(h.Estado, nuevoEstado); err != nil {
			return nil, err
		}
		h.Estado = nuevoEstado
	}

	if cmd.IDTipo != nil {
		h.IDTipoHabitacion = *cmd.IDTipo
	}

	now := time.Now()
	h.UpdatedAt = &now

	if err := a.habitacionRepo.Update(ctx, h); err != nil {
		return nil, err
	}

	resp := toHabitacionResponse(*h)
	return &resp, nil
}

// ListTiposHabitacion retorna todos los tipos de habitacion disponibles.
func (a *InventarioApp) ListTiposHabitacion(ctx context.Context) ([]TipoHabitacionResponse, error) {
	tipos, err := a.tipoRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]TipoHabitacionResponse, 0, len(tipos))
	for _, t := range tipos {
		result = append(result, TipoHabitacionResponse{
			ID:         t.ID,
			Nombre:     t.Nombre,
			PrecioBase: t.PrecioBase,
			Capacidad:  t.Capacidad,
		})
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Mappers
// ---------------------------------------------------------------------------

func toHabitacionResponse(h model.Habitacion) HabitacionResponse {
	return HabitacionResponse{
		ID:               h.ID,
		Numero:           h.Numero.String(),
		IDTipoHabitacion: h.IDTipoHabitacion,
		Estado:           h.Estado.String(),
		UpdatedAt:        h.UpdatedAt,
	}
}
