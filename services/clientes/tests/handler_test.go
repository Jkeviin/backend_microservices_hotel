package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"clientes/internal/application"
	domainservice "clientes/internal/domain/service"
	httpinfra "clientes/internal/infrastructure/http"
	"clientes/internal/infrastructure/http/middleware"
	"clientes/internal/infrastructure/persistence/memory"
)

const testJWTSecret = "test-secret"

// newTestRouter crea un router chi completo con repo in-memory y JWT auth.
func newTestRouter() chi.Router {
	repo := memory.NewClienteRepo()
	domainSvc := domainservice.NewClienteDomainService(repo)
	appSvc := application.NewClienteAppService(repo, domainSvc)
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
// POST /api/clientes
// ---------------------------------------------------------------------------

func TestHandler_CreateCliente_201(t *testing.T) {
	r := newTestRouter()

	body := `{"nombre":"Test Cliente","email":"test.handler@email.com","telefono":"3001112233"}`
	req := httptest.NewRequest(http.MethodPost, "/api/clientes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader("admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp application.ClienteResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("error decodificando respuesta: %v", err)
	}
	if resp.ID == 0 {
		t.Error("ID no deberia ser 0")
	}
	if resp.Nombre != "Test Cliente" {
		t.Errorf("Nombre = %q, esperado %q", resp.Nombre, "Test Cliente")
	}
}

func TestHandler_CreateCliente_SinNombre_400(t *testing.T) {
	r := newTestRouter()

	body := `{"nombre":"","email":"valid@email.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/clientes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader("recepcion"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestHandler_CreateCliente_EmailDuplicado_409(t *testing.T) {
	r := newTestRouter()

	body := `{"nombre":"Duplicado","email":"juan.perez@email.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/clientes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader("admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusConflict, w.Body.String())
	}
}

func TestHandler_CreateCliente_SinAuth_401(t *testing.T) {
	r := newTestRouter()

	body := `{"nombre":"Test","email":"test@email.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/clientes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, esperado %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_CreateCliente_RolInvalido_403(t *testing.T) {
	r := newTestRouter()

	body := `{"nombre":"Test","email":"test@email.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/clientes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader("huesped"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, esperado %d", w.Code, http.StatusForbidden)
	}
}

// ---------------------------------------------------------------------------
// GET /api/clientes/{id}
// ---------------------------------------------------------------------------

func TestHandler_GetCliente_200(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/clientes/1", nil)
	req.Header.Set("Authorization", authHeader("recepcion"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp application.ClienteResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("error decodificando: %v", err)
	}
	if resp.Nombre != "Juan Perez" {
		t.Errorf("Nombre = %q, esperado %q", resp.Nombre, "Juan Perez")
	}
}

func TestHandler_GetCliente_404(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/clientes/999", nil)
	req.Header.Set("Authorization", authHeader("admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestHandler_GetCliente_IDInvalido_400(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/clientes/abc", nil)
	req.Header.Set("Authorization", authHeader("admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado %d", w.Code, http.StatusBadRequest)
	}
}

// ---------------------------------------------------------------------------
// PATCH /api/clientes/{id}
// ---------------------------------------------------------------------------

func TestHandler_UpdateCliente_200(t *testing.T) {
	r := newTestRouter()

	body := `{"nombre":"Juan Actualizado"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/clientes/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader("admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp application.ClienteResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("error decodificando: %v", err)
	}
	if resp.Nombre != "Juan Actualizado" {
		t.Errorf("Nombre = %q, esperado %q", resp.Nombre, "Juan Actualizado")
	}
}

func TestHandler_UpdateCliente_404(t *testing.T) {
	r := newTestRouter()

	body := `{"nombre":"No existe"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/clientes/999", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader("admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, esperado %d", w.Code, http.StatusNotFound)
	}
}

func TestHandler_UpdateCliente_EmailInvalido_400(t *testing.T) {
	r := newTestRouter()

	body := `{"email":"invalido"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/clientes/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader("admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado %d. body: %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// X-Request-Id
// ---------------------------------------------------------------------------

func TestHandler_RequestID_Generated(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/clientes/1", nil)
	req.Header.Set("Authorization", authHeader("admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	rid := w.Header().Get("X-Request-Id")
	if rid == "" {
		t.Error("X-Request-Id no deberia estar vacio")
	}
}

func TestHandler_RequestID_Preserved(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/clientes/1", nil)
	req.Header.Set("Authorization", authHeader("admin"))
	req.Header.Set("X-Request-Id", "my-custom-id")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	rid := w.Header().Get("X-Request-Id")
	if rid != "my-custom-id" {
		t.Errorf("X-Request-Id = %q, esperado %q", rid, "my-custom-id")
	}
}
