package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"inventario/internal/application"
	infrahttp "inventario/internal/infrastructure/http"
	"inventario/internal/infrastructure/persistence/memory"
)

const testJWTSecret = "test-secret"

func newTestRouter() http.Handler {
	app := application.NewInventarioApp(
		memory.NewHabitacionRepo(),
		memory.NewTipoHabitacionRepo(),
	)
	handler := infrahttp.NewHandler(app)
	return infrahttp.NewRouter(handler, testJWTSecret)
}

func generateToken(rol string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "user-1",
		"role": rol,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	})
	s, _ := token.SignedString([]byte(testJWTSecret))
	return s
}

func doRequest(t *testing.T, router http.Handler, method, path, rol string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if rol != "" {
		req.Header.Set("Authorization", "Bearer "+generateToken(rol))
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// ---------------------------------------------------------------------------
// GET /api/habitaciones
// ---------------------------------------------------------------------------

func TestHTTP_ListHabitaciones_200(t *testing.T) {
	router := newTestRouter()
	rr := doRequest(t, router, "GET", "/api/habitaciones", "recepcion", nil)

	if rr.Code != http.StatusOK {
		t.Errorf("esperado 200, obtenido %d: %s", rr.Code, rr.Body.String())
	}

	var result []application.HabitacionResponse
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("error decodificando respuesta: %v", err)
	}
	if len(result) == 0 {
		t.Error("se esperaban habitaciones en la respuesta")
	}
}

func TestHTTP_ListHabitaciones_FiltroTipo_200(t *testing.T) {
	router := newTestRouter()
	rr := doRequest(t, router, "GET", "/api/habitaciones?tipo=1", "admin", nil)

	if rr.Code != http.StatusOK {
		t.Errorf("esperado 200, obtenido %d: %s", rr.Code, rr.Body.String())
	}

	var result []application.HabitacionResponse
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("error decodificando respuesta: %v", err)
	}
	for _, h := range result {
		if h.IDTipoHabitacion != 1 {
			t.Errorf("habitacion %s tiene tipo %d, esperado 1", h.Numero, h.IDTipoHabitacion)
		}
	}
}

func TestHTTP_ListHabitaciones_FiltroEstado_200(t *testing.T) {
	router := newTestRouter()
	rr := doRequest(t, router, "GET", "/api/habitaciones?estado=mantenimiento", "recepcion", nil)

	if rr.Code != http.StatusOK {
		t.Errorf("esperado 200, obtenido %d: %s", rr.Code, rr.Body.String())
	}

	var result []application.HabitacionResponse
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("error decodificando respuesta: %v", err)
	}
	for _, h := range result {
		if h.Estado != "mantenimiento" {
			t.Errorf("habitacion %s tiene estado %s, esperado mantenimiento", h.Numero, h.Estado)
		}
	}
}

func TestHTTP_ListHabitaciones_SinAuth_401(t *testing.T) {
	router := newTestRouter()
	rr := doRequest(t, router, "GET", "/api/habitaciones", "", nil)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("esperado 401, obtenido %d", rr.Code)
	}
}

func TestHTTP_ListHabitaciones_RolInvalido_403(t *testing.T) {
	router := newTestRouter()
	rr := doRequest(t, router, "GET", "/api/habitaciones", "huesped", nil)

	if rr.Code != http.StatusForbidden {
		t.Errorf("esperado 403, obtenido %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// GET /api/habitaciones/{id}
// ---------------------------------------------------------------------------

func TestHTTP_GetHabitacion_200(t *testing.T) {
	router := newTestRouter()
	rr := doRequest(t, router, "GET", "/api/habitaciones/1", "recepcion", nil)

	if rr.Code != http.StatusOK {
		t.Errorf("esperado 200, obtenido %d: %s", rr.Code, rr.Body.String())
	}

	var result application.HabitacionResponse
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("error decodificando respuesta: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("esperado ID 1, obtenido %d", result.ID)
	}
}

func TestHTTP_GetHabitacion_404(t *testing.T) {
	router := newTestRouter()
	rr := doRequest(t, router, "GET", "/api/habitaciones/999", "recepcion", nil)

	if rr.Code != http.StatusNotFound {
		t.Errorf("esperado 404, obtenido %d: %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// PATCH /api/habitaciones/{id}
// ---------------------------------------------------------------------------

func TestHTTP_UpdateHabitacion_200(t *testing.T) {
	router := newTestRouter()
	body := map[string]interface{}{"estado": "mantenimiento"}
	rr := doRequest(t, router, "PATCH", "/api/habitaciones/1", "admin", body)

	if rr.Code != http.StatusOK {
		t.Errorf("esperado 200, obtenido %d: %s", rr.Code, rr.Body.String())
	}

	var result application.HabitacionResponse
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("error decodificando respuesta: %v", err)
	}
	if result.Estado != "mantenimiento" {
		t.Errorf("esperado estado mantenimiento, obtenido %s", result.Estado)
	}
}

func TestHTTP_UpdateHabitacion_NoAdmin_403(t *testing.T) {
	router := newTestRouter()
	body := map[string]interface{}{"estado": "mantenimiento"}
	rr := doRequest(t, router, "PATCH", "/api/habitaciones/1", "recepcion", body)

	if rr.Code != http.StatusForbidden {
		t.Errorf("esperado 403, obtenido %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHTTP_UpdateHabitacion_OcupadaAMantenimiento_400(t *testing.T) {
	router := newTestRouter()
	body := map[string]interface{}{"estado": "mantenimiento"}
	// Habitacion 4 esta ocupada
	rr := doRequest(t, router, "PATCH", "/api/habitaciones/4", "admin", body)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obtenido %d: %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// GET /api/tipos-habitacion
// ---------------------------------------------------------------------------

func TestHTTP_ListTiposHabitacion_200(t *testing.T) {
	router := newTestRouter()
	rr := doRequest(t, router, "GET", "/api/tipos-habitacion", "recepcion", nil)

	if rr.Code != http.StatusOK {
		t.Errorf("esperado 200, obtenido %d: %s", rr.Code, rr.Body.String())
	}

	var result []application.TipoHabitacionResponse
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("error decodificando respuesta: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("esperado 3 tipos, obtenido %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func mustParseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(fmt.Sprintf("fecha invalida en test: %s", s))
	}
	return t
}
