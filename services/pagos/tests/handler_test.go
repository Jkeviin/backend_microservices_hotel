package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"pagos/internal/application"
	infrahttp "pagos/internal/infrastructure/http"
	"pagos/internal/infrastructure/persistence/memory"
)

const testJWTSecret = "test-secret"

func newTestRouter() http.Handler {
	pagoRepo := memory.NewPagoRepo()
	reservaChecker := memory.NewReservaChecker()
	app := application.NewPagoApp(pagoRepo, reservaChecker)
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

func doRequest(t *testing.T, router http.Handler, method, path, rol string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
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
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

// ---------------------------------------------------------------------------
// POST /api/pagos
// ---------------------------------------------------------------------------

func TestHTTP_CreatePago_201(t *testing.T) {
	router := newTestRouter()
	body := map[string]interface{}{
		"id_reserva": 2,
		"monto":      300.00,
		"fecha_pago": "2026-05-01",
	}
	headers := map[string]string{"Idempotency-Key": "http-test-key-1"}
	rr := doRequest(t, router, "POST", "/api/pagos", "recepcion", body, headers)

	if rr.Code != http.StatusCreated {
		t.Errorf("esperado 201, obtenido %d: %s", rr.Code, rr.Body.String())
	}

	var result application.PagoResponse
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("error decodificando respuesta: %v", err)
	}
	if result.EstadoPago != "aprobado" {
		t.Errorf("esperado estado aprobado, obtenido %s", result.EstadoPago)
	}
}

func TestHTTP_CreatePago_SinIdempotencyKey_400(t *testing.T) {
	router := newTestRouter()
	body := map[string]interface{}{
		"id_reserva": 2,
		"monto":      300.00,
		"fecha_pago": "2026-05-01",
	}
	// Sin header Idempotency-Key
	rr := doRequest(t, router, "POST", "/api/pagos", "recepcion", body, nil)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obtenido %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHTTP_CreatePago_ReservaInvalida_422(t *testing.T) {
	router := newTestRouter()
	body := map[string]interface{}{
		"id_reserva": 3, // cancelada
		"monto":      100.00,
		"fecha_pago": "2026-05-01",
	}
	headers := map[string]string{"Idempotency-Key": "http-test-key-422"}
	rr := doRequest(t, router, "POST", "/api/pagos", "recepcion", body, headers)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("esperado 422, obtenido %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHTTP_CreatePago_MontoCero_400(t *testing.T) {
	router := newTestRouter()
	body := map[string]interface{}{
		"id_reserva": 2,
		"monto":      0,
		"fecha_pago": "2026-05-01",
	}
	headers := map[string]string{"Idempotency-Key": "http-test-key-monto0"}
	rr := doRequest(t, router, "POST", "/api/pagos", "recepcion", body, headers)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obtenido %d: %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// GET /api/pagos/{id}
// ---------------------------------------------------------------------------

func TestHTTP_GetPago_200(t *testing.T) {
	router := newTestRouter()
	rr := doRequest(t, router, "GET", "/api/pagos/1", "recepcion", nil, nil)

	if rr.Code != http.StatusOK {
		t.Errorf("esperado 200, obtenido %d: %s", rr.Code, rr.Body.String())
	}

	var result application.PagoResponse
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("error decodificando respuesta: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("esperado ID 1, obtenido %d", result.ID)
	}
}

func TestHTTP_GetPago_404(t *testing.T) {
	router := newTestRouter()
	rr := doRequest(t, router, "GET", "/api/pagos/999", "recepcion", nil, nil)

	if rr.Code != http.StatusNotFound {
		t.Errorf("esperado 404, obtenido %d: %s", rr.Code, rr.Body.String())
	}
}
