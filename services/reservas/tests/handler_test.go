package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"reservas/internal/application"
	domainservice "reservas/internal/domain/service"
	httpinfra "reservas/internal/infrastructure/http"
	"reservas/internal/infrastructure/http/middleware"
	"reservas/internal/infrastructure/persistence/memory"
)

const testJWTSecret = "test-secret"

// newTestRouter crea un router chi completo con repos in-memory y JWT auth.
func newTestRouter() chi.Router {
	reservaRepo := memory.NewReservaRepo()
	estadoRepo := memory.NewEstadoRepo()
	habChecker := memory.NewHabitacionChecker()
	domainSvc := domainservice.NewReservaDomainService(reservaRepo, habChecker)
	appSvc := application.NewReservaAppService(reservaRepo, estadoRepo, domainSvc)
	handler := httpinfra.NewHandler(appSvc)

	rl := middleware.NewRateLimiter(100, time.Minute)

	r := chi.NewRouter()
	httpinfra.SetupRoutes(
		r,
		handler,
		middleware.JWTAuth(testJWTSecret),
		middleware.Logging,
		middleware.RequestID,
		rl.Middleware,
		middleware.RequireRole,
	)
	return r
}

// generateToken genera un JWT valido de prueba con el rol indicado.
func generateToken(role string) string {
	claims := jwt.MapClaims{
		"sub":  "test-user-1",
		"role": role,
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testJWTSecret))
	return signed
}

func authHeader(role string) string {
	return "Bearer " + generateToken(role)
}

// ---------------------------------------------------------------------------
// POST /api/reservas
// ---------------------------------------------------------------------------

func TestHandler_CreateReserva_201(t *testing.T) {
	r := newTestRouter()

	body := fmt.Sprintf(`{"id_cliente":1,"id_habitacion":4,"fecha_inicio":"%s","fecha_fin":"%s"}`,
		futureDate(10), futureDate(13))

	req := httptest.NewRequest(http.MethodPost, "/api/reservas", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader("admin"))
	req.Header.Set("Idempotency-Key", "handler-create-001")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp application.ReservaResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("error decodificando respuesta: %v", err)
	}
	if resp.ID == 0 {
		t.Error("ID no deberia ser 0")
	}
}

func TestHandler_CreateReserva_SinIdempotencyKey_400(t *testing.T) {
	r := newTestRouter()

	body := fmt.Sprintf(`{"id_cliente":1,"id_habitacion":4,"fecha_inicio":"%s","fecha_fin":"%s"}`,
		futureDate(10), futureDate(13))

	req := httptest.NewRequest(http.MethodPost, "/api/reservas", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader("admin"))
	// Sin Idempotency-Key

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestHandler_CreateReserva_Solapamiento_409(t *testing.T) {
	r := newTestRouter()

	// Crear primera reserva.
	body1 := fmt.Sprintf(`{"id_cliente":1,"id_habitacion":5,"fecha_inicio":"%s","fecha_fin":"%s"}`,
		futureDate(20), futureDate(25))

	req1 := httptest.NewRequest(http.MethodPost, "/api/reservas", bytes.NewBufferString(body1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", authHeader("admin"))
	req1.Header.Set("Idempotency-Key", "handler-solap-001")

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("primera reserva: status = %d, esperado %d. body: %s", w1.Code, http.StatusCreated, w1.Body.String())
	}

	// Intentar crear segunda solapada.
	body2 := fmt.Sprintf(`{"id_cliente":2,"id_habitacion":5,"fecha_inicio":"%s","fecha_fin":"%s"}`,
		futureDate(22), futureDate(27))

	req2 := httptest.NewRequest(http.MethodPost, "/api/reservas", bytes.NewBufferString(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", authHeader("admin"))
	req2.Header.Set("Idempotency-Key", "handler-solap-002")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("segunda reserva: status = %d, esperado %d. body: %s", w2.Code, http.StatusConflict, w2.Body.String())
	}
}

// ---------------------------------------------------------------------------
// GET /api/reservas/{id}
// ---------------------------------------------------------------------------

func TestHandler_GetReserva_200(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/reservas/1", nil)
	req.Header.Set("Authorization", authHeader("recepcion"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp application.ReservaResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("error decodificando: %v", err)
	}
	if resp.IDCliente != 1 {
		t.Errorf("IDCliente = %d, esperado 1", resp.IDCliente)
	}
}

func TestHandler_GetReserva_404(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/reservas/999", nil)
	req.Header.Set("Authorization", authHeader("admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// PATCH /api/reservas/{id}
// ---------------------------------------------------------------------------

func TestHandler_PatchReserva_Cancelar_200(t *testing.T) {
	r := newTestRouter()

	// Reserva 2: pendiente_pago, version 1.
	body := `{"accion":"cancelar","version":1}`
	req := httptest.NewRequest(http.MethodPatch, "/api/reservas/2", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader("admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp application.ReservaResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("error decodificando: %v", err)
	}
	if resp.Estado != "cancelada" {
		t.Errorf("Estado = %q, esperado %q", resp.Estado, "cancelada")
	}
}

func TestHandler_PatchReserva_Reprogramar_200(t *testing.T) {
	r := newTestRouter()

	// Primero crear una reserva con fecha futura que se pueda reprogramar.
	createBody := fmt.Sprintf(`{"id_cliente":1,"id_habitacion":4,"fecha_inicio":"%s","fecha_fin":"%s"}`,
		futureDate(30), futureDate(33))

	createReq := httptest.NewRequest(http.MethodPost, "/api/reservas", bytes.NewBufferString(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", authHeader("admin"))
	createReq.Header.Set("Idempotency-Key", "handler-reprog-001")

	cw := httptest.NewRecorder()
	r.ServeHTTP(cw, createReq)

	if cw.Code != http.StatusCreated {
		t.Fatalf("create: status = %d, esperado %d. body: %s", cw.Code, http.StatusCreated, cw.Body.String())
	}

	var created application.ReservaResponse
	if err := json.NewDecoder(cw.Body).Decode(&created); err != nil {
		t.Fatalf("error decodificando create: %v", err)
	}

	// Reprogramar.
	patchBody := fmt.Sprintf(`{"fecha_inicio":"%s","fecha_fin":"%s","version":%d}`,
		futureDate(35), futureDate(38), created.Version)

	patchReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/reservas/%d", created.ID),
		bytes.NewBufferString(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Authorization", authHeader("recepcion"))

	pw := httptest.NewRecorder()
	r.ServeHTTP(pw, patchReq)

	if pw.Code != http.StatusOK {
		t.Fatalf("patch: status = %d, esperado %d. body: %s", pw.Code, http.StatusOK, pw.Body.String())
	}
}

// ---------------------------------------------------------------------------
// GET /api/reservas/estados
// ---------------------------------------------------------------------------

func TestHandler_ListEstados_200(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/reservas/estados", nil)
	req.Header.Set("Authorization", authHeader("admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var estados []application.EstadoResponse
	if err := json.NewDecoder(w.Body).Decode(&estados); err != nil {
		t.Fatalf("error decodificando: %v", err)
	}
	if len(estados) != 4 {
		t.Errorf("len(estados) = %d, esperado 4", len(estados))
	}
}
