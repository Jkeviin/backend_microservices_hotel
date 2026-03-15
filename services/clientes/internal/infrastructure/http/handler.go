package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"clientes/internal/application"
	"clientes/internal/domain/model"
)

// Handler agrupa los handlers HTTP del servicio de clientes.
type Handler struct {
	app *application.ClienteAppService
}

// NewHandler crea un Handler con el application service inyectado.
func NewHandler(app *application.ClienteAppService) *Handler {
	return &Handler{app: app}
}

// CreateCliente maneja POST /api/clientes.
func (h *Handler) CreateCliente(w http.ResponseWriter, r *http.Request) {
	var cmd application.CreateClienteCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body JSON invalido"})
		return
	}

	resp, err := h.app.CreateCliente(r.Context(), cmd)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// GetCliente maneja GET /api/clientes/{id}.
func (h *Handler) GetCliente(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id debe ser un entero"})
		return
	}

	resp, err := h.app.GetCliente(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateCliente maneja PATCH /api/clientes/{id}.
func (h *Handler) UpdateCliente(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id debe ser un entero"})
		return
	}

	var cmd application.UpdateClienteCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body JSON invalido"})
		return
	}

	resp, err := h.app.UpdateCliente(r.Context(), id, cmd)
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrClienteNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrEmailDuplicated):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrNombreRequerido):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, model.ErrEmailInvalido):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "error interno del servidor"})
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
