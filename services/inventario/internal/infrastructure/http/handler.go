package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"inventario/internal/application"
	"inventario/internal/domain/model"
)

// Handler agrupa los endpoints HTTP del servicio de inventario.
type Handler struct {
	app *application.InventarioApp
}

// NewHandler crea un nuevo Handler.
func NewHandler(app *application.InventarioApp) *Handler {
	return &Handler{app: app}
}

// ListHabitaciones maneja GET /api/habitaciones con query params opcionales:
// tipo, estado, disponible_desde, disponible_hasta.
func (h *Handler) ListHabitaciones(w http.ResponseWriter, r *http.Request) {
	filtros, err := parseFiltros(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.app.ListHabitaciones(r.Context(), filtros)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetHabitacion maneja GET /api/habitaciones/{id}.
func (h *Handler) GetHabitacion(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id debe ser un numero entero")
		return
	}

	result, err := h.app.GetHabitacion(r.Context(), id)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// UpdateHabitacion maneja PATCH /api/habitaciones/{id}.
func (h *Handler) UpdateHabitacion(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id debe ser un numero entero")
		return
	}

	var cmd application.UpdateHabitacionCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de la solicitud invalido")
		return
	}

	result, err := h.app.UpdateHabitacion(r.Context(), id, cmd)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ListTiposHabitacion maneja GET /api/tipos-habitacion.
func (h *Handler) ListTiposHabitacion(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.ListTiposHabitacion(r.Context())
	if err != nil {
		mapDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func parseFiltros(r *http.Request) (model.HabitacionFiltros, error) {
	var filtros model.HabitacionFiltros

	if v := r.URL.Query().Get("tipo"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			return filtros, errors.New("parametro 'tipo' debe ser un numero entero")
		}
		filtros.Tipo = &id
	}

	if v := r.URL.Query().Get("estado"); v != "" {
		estado, err := model.NewEstadoHabitacion(v)
		if err != nil {
			return filtros, err
		}
		filtros.Estado = &estado
	}

	if v := r.URL.Query().Get("disponible_desde"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return filtros, errors.New("parametro 'disponible_desde' debe tener formato YYYY-MM-DD")
		}
		filtros.DisponibleDesde = &t
	}

	if v := r.URL.Query().Get("disponible_hasta"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return filtros, errors.New("parametro 'disponible_hasta' debe tener formato YYYY-MM-DD")
		}
		filtros.DisponibleHasta = &t
	}

	return filtros, nil
}

func mapDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrHabitacionNotFound),
		errors.Is(err, model.ErrTipoNotFound):
		writeError(w, http.StatusNotFound, err.Error())

	case errors.Is(err, model.ErrEstadoInvalido),
		errors.Is(err, model.ErrNoMantenimientoSiOcupada),
		errors.Is(err, model.ErrNumeroVacio),
		errors.Is(err, model.ErrNumeroMuyLargo),
		errors.Is(err, model.ErrNumeroHabitacionDuplicado),
		errors.Is(err, model.ErrFiltroFechasIncompletas),
		errors.Is(err, model.ErrFiltroFechasInvalidas):
		writeError(w, http.StatusBadRequest, err.Error())

	default:
		writeError(w, http.StatusInternalServerError, "error interno del servidor")
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
