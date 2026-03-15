package memory

import (
	"context"
	"sync"
	"time"

	"inventario/internal/domain/model"
)

// reservaMock simula una reserva para verificar disponibilidad.
type reservaMock struct {
	IDHabitacion int
	CheckIn      time.Time
	CheckOut     time.Time
}

// HabitacionRepo es una implementacion en memoria del repositorio de habitaciones.
type HabitacionRepo struct {
	mu           sync.RWMutex
	habitaciones map[int]model.Habitacion
	nextID       int
	reservas     []reservaMock
}

// NewHabitacionRepo crea un repositorio pre-cargado con datos realistas.
func NewHabitacionRepo() *HabitacionRepo {
	mustNumero := func(s string) model.NumeroHabitacion {
		n, _ := model.NewNumeroHabitacion(s)
		return n
	}

	habitaciones := map[int]model.Habitacion{
		1: {ID: 1, Numero: mustNumero("101"), IDTipoHabitacion: 1, Estado: model.EstadoDisponible},
		2: {ID: 2, Numero: mustNumero("102"), IDTipoHabitacion: 1, Estado: model.EstadoDisponible},
		3: {ID: 3, Numero: mustNumero("201"), IDTipoHabitacion: 2, Estado: model.EstadoDisponible},
		4: {ID: 4, Numero: mustNumero("202"), IDTipoHabitacion: 2, Estado: model.EstadoOcupada},
		5: {ID: 5, Numero: mustNumero("301"), IDTipoHabitacion: 3, Estado: model.EstadoDisponible},
		6: {ID: 6, Numero: mustNumero("302"), IDTipoHabitacion: 3, Estado: model.EstadoMantenimiento},
		7: {ID: 7, Numero: mustNumero("401"), IDTipoHabitacion: 2, Estado: model.EstadoDisponible},
		8: {ID: 8, Numero: mustNumero("402"), IDTipoHabitacion: 1, Estado: model.EstadoDisponible},
		9: {ID: 9, Numero: mustNumero("501"), IDTipoHabitacion: 3, Estado: model.EstadoDisponible},
	}

	// Reservas mock para simular disponibilidad.
	now := time.Now()
	reservas := []reservaMock{
		{IDHabitacion: 4, CheckIn: now.AddDate(0, 0, -1), CheckOut: now.AddDate(0, 0, 3)},
		{IDHabitacion: 1, CheckIn: now.AddDate(0, 0, 5), CheckOut: now.AddDate(0, 0, 8)},
		{IDHabitacion: 3, CheckIn: now.AddDate(0, 0, 10), CheckOut: now.AddDate(0, 0, 14)},
	}

	return &HabitacionRepo{
		habitaciones: habitaciones,
		nextID:       10,
		reservas:     reservas,
	}
}

// List retorna habitaciones filtradas segun los criterios proporcionados.
func (r *HabitacionRepo) List(_ context.Context, filtros model.HabitacionFiltros) ([]model.Habitacion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Determinar habitaciones excluidas por solapamiento de reservas.
	excluidas := make(map[int]bool)
	if filtros.DisponibleDesde != nil && filtros.DisponibleHasta != nil {
		for _, res := range r.reservas {
			if seSolapan(*filtros.DisponibleDesde, *filtros.DisponibleHasta, res.CheckIn, res.CheckOut) {
				excluidas[res.IDHabitacion] = true
			}
		}
	}

	result := make([]model.Habitacion, 0)
	for _, h := range r.habitaciones {
		if filtros.Tipo != nil && h.IDTipoHabitacion != *filtros.Tipo {
			continue
		}
		if filtros.Estado != nil && h.Estado != *filtros.Estado {
			continue
		}
		if excluidas[h.ID] {
			continue
		}
		result = append(result, h)
	}

	return result, nil
}

// GetByID retorna una habitacion por su identificador.
func (r *HabitacionRepo) GetByID(_ context.Context, id int) (*model.Habitacion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	h, ok := r.habitaciones[id]
	if !ok {
		return nil, model.ErrHabitacionNotFound
	}
	return &h, nil
}

// Update persiste los cambios de una habitacion existente.
func (r *HabitacionRepo) Update(_ context.Context, h *model.Habitacion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.habitaciones[h.ID]; !ok {
		return model.ErrHabitacionNotFound
	}
	r.habitaciones[h.ID] = *h
	return nil
}

// seSolapan determina si dos rangos de fechas se solapan.
// [desde1, hasta1) y [desde2, hasta2)
func seSolapan(desde1, hasta1, desde2, hasta2 time.Time) bool {
	return desde1.Before(hasta2) && desde2.Before(hasta1)
}
