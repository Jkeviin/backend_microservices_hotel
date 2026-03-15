package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"pagos/internal/application"
	"pagos/internal/domain/model"
	"pagos/internal/infrastructure/http/middleware"
)

// Handler agrupa los endpoints HTTP del servicio de pagos.
type Handler struct {
	app *application.PagoApp
}

// NewHandler crea un nuevo Handler.
func NewHandler(app *application.PagoApp) *Handler {
	return &Handler{app: app}
}

// CreatePago maneja POST /api/pagos.
func (h *Handler) CreatePago(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreatePagoCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de la solicitud invalido")
		return
	}

	// Inyectar la idempotency key desde el contexto (puesta por el middleware).
	cmd.IdempotencyKey = middleware.GetIdempotencyKey(r.Context())

	result, err := h.app.CreatePago(r.Context(), cmd)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// GetPago maneja GET /api/pagos/{id}.
func (h *Handler) GetPago(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id debe ser un numero entero")
		return
	}

	result, err := h.app.GetPago(r.Context(), id)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mapDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrPagoNotFound):
		writeError(w, http.StatusNotFound, err.Error())

	case errors.Is(err, model.ErrReservaNoValida):
		writeError(w, http.StatusUnprocessableEntity, err.Error())

	case errors.Is(err, model.ErrMontoInvalido),
		errors.Is(err, model.ErrEstadoPagoInvalido):
		writeError(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, model.ErrIdempotencyKeyDuplicada):
		writeError(w, http.StatusConflict, err.Error())

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
