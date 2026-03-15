package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"reservas/internal/application"
	"reservas/internal/domain/model"
	"reservas/internal/infrastructure/http/middleware"
)

// Handler agrupa los handlers HTTP del servicio de reservas.
type Handler struct {
	app *application.ReservaAppService
}

// NewHandler crea un Handler con el application service inyectado.
func NewHandler(app *application.ReservaAppService) *Handler {
	return &Handler{app: app}
}

// CreateReserva maneja POST /api/reservas.
func (h *Handler) CreateReserva(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateReservaCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body JSON invalido"})
		return
	}

	// Obtener idempotency key del contexto (inyectada por middleware).
	cmd.IdempotencyKey = middleware.IdempotencyKeyFromContext(r.Context())

	resp, err := h.app.CreateReserva(r.Context(), cmd)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// GetReserva maneja GET /api/reservas/{id}.
func (h *Handler) GetReserva(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id debe ser un entero"})
		return
	}

	resp, err := h.app.GetReserva(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// PatchReserva maneja PATCH /api/reservas/{id}.
// Body: {"accion":"cancelar","version":N} o {"fecha_inicio":"...","fecha_fin":"...","version":N}
func (h *Handler) PatchReserva(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id debe ser un entero"})
		return
	}

	var cmd application.PatchReservaCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body JSON invalido"})
		return
	}

	if cmd.Version == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "version es requerida"})
		return
	}

	var resp *application.ReservaResponse

	if cmd.Accion == "cancelar" {
		resp, err = h.app.CancelarReserva(r.Context(), id, cmd.Version)
	} else if cmd.FechaInicio != "" && cmd.FechaFin != "" {
		resp, err = h.app.ReprogramarReserva(r.Context(), id, cmd.FechaInicio, cmd.FechaFin, cmd.Version)
	} else {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "debe indicar accion='cancelar' o fecha_inicio y fecha_fin para reprogramar"})
		return
	}

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListEstados maneja GET /api/reservas/estados.
func (h *Handler) ListEstados(w http.ResponseWriter, r *http.Request) {
	estados, err := h.app.ListEstados(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, estados)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrReservaNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrConflictoFechas):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrVersionConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrHabitacionNoDisponible):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrFechasInvalidas):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrEstadoNoPermiteCambio):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrReservaYaIniciada):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrIdempotencyKeyDuplicada):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "error interno del servidor"})
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
