package tests

import (
	"context"
	"testing"

	"clientes/internal/application"
	"clientes/internal/domain/model"
	domainservice "clientes/internal/domain/service"
	"clientes/internal/infrastructure/persistence/memory"
)

// newTestAppService crea un application service con repo in-memory fresco.
func newTestAppService() *application.ClienteAppService {
	repo := memory.NewClienteRepo()
	domainSvc := domainservice.NewClienteDomainService(repo)
	return application.NewClienteAppService(repo, domainSvc)
}

func strPtr(s string) *string {
	return &s
}

// ---------------------------------------------------------------------------
// CreateCliente
// ---------------------------------------------------------------------------

func TestCreateCliente_Success(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	resp, err := svc.CreateCliente(ctx, application.CreateClienteCommand{
		Nombre:   "Nuevo Cliente",
		Email:    "nuevo@email.com",
		Telefono: strPtr("3009999999"),
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.ID == 0 {
		t.Error("ID no deberia ser 0")
	}
	if resp.Nombre != "Nuevo Cliente" {
		t.Errorf("Nombre = %q, esperado %q", resp.Nombre, "Nuevo Cliente")
	}
	if resp.Email != "nuevo@email.com" {
		t.Errorf("Email = %q, esperado %q", resp.Email, "nuevo@email.com")
	}
	if resp.Telefono == nil || *resp.Telefono != "3009999999" {
		t.Errorf("Telefono = %v, esperado %q", resp.Telefono, "3009999999")
	}
}

func TestCreateCliente_SinTelefono(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	resp, err := svc.CreateCliente(ctx, application.CreateClienteCommand{
		Nombre: "Sin Telefono",
		Email:  "sintel@email.com",
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.Telefono != nil {
		t.Errorf("Telefono deberia ser nil, got %v", resp.Telefono)
	}
}

func TestCreateCliente_NombreVacio(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	_, err := svc.CreateCliente(ctx, application.CreateClienteCommand{
		Nombre: "",
		Email:  "test@email.com",
	})
	if err != model.ErrNombreRequerido {
		t.Errorf("error = %v, esperado ErrNombreRequerido", err)
	}
}

func TestCreateCliente_EmailInvalido(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	_, err := svc.CreateCliente(ctx, application.CreateClienteCommand{
		Nombre: "Test",
		Email:  "invalido",
	})
	if err != model.ErrEmailInvalido {
		t.Errorf("error = %v, esperado ErrEmailInvalido", err)
	}
}

func TestCreateCliente_EmailDuplicado(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	// juan.perez@email.com ya existe en los datos semilla.
	_, err := svc.CreateCliente(ctx, application.CreateClienteCommand{
		Nombre: "Otro Juan",
		Email:  "juan.perez@email.com",
	})
	if err != model.ErrEmailDuplicated {
		t.Errorf("error = %v, esperado ErrEmailDuplicated", err)
	}
}

// ---------------------------------------------------------------------------
// GetCliente
// ---------------------------------------------------------------------------

func TestGetCliente_Existente(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	resp, err := svc.GetCliente(ctx, 1)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.Nombre != "Juan Perez" {
		t.Errorf("Nombre = %q, esperado %q", resp.Nombre, "Juan Perez")
	}
	if resp.Email != "juan.perez@email.com" {
		t.Errorf("Email = %q, esperado %q", resp.Email, "juan.perez@email.com")
	}
}

func TestGetCliente_NoExistente(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	_, err := svc.GetCliente(ctx, 999)
	if err != model.ErrClienteNotFound {
		t.Errorf("error = %v, esperado ErrClienteNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// UpdateCliente
// ---------------------------------------------------------------------------

func TestUpdateCliente_SoloNombre(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	resp, err := svc.UpdateCliente(ctx, 1, application.UpdateClienteCommand{
		Nombre: strPtr("Juan Actualizado"),
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.Nombre != "Juan Actualizado" {
		t.Errorf("Nombre = %q, esperado %q", resp.Nombre, "Juan Actualizado")
	}
	// Email no debe cambiar.
	if resp.Email != "juan.perez@email.com" {
		t.Errorf("Email = %q, esperado %q", resp.Email, "juan.perez@email.com")
	}
	if resp.UpdatedAt == nil {
		t.Error("UpdatedAt deberia estar seteado")
	}
}

func TestUpdateCliente_SoloEmail(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	resp, err := svc.UpdateCliente(ctx, 1, application.UpdateClienteCommand{
		Email: strPtr("juan.nuevo@email.com"),
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.Email != "juan.nuevo@email.com" {
		t.Errorf("Email = %q, esperado %q", resp.Email, "juan.nuevo@email.com")
	}
	// Nombre no debe cambiar.
	if resp.Nombre != "Juan Perez" {
		t.Errorf("Nombre = %q, esperado %q", resp.Nombre, "Juan Perez")
	}
}

func TestUpdateCliente_NombreYEmail(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	resp, err := svc.UpdateCliente(ctx, 2, application.UpdateClienteCommand{
		Nombre: strPtr("Maria Actualizada"),
		Email:  strPtr("maria.nueva@email.com"),
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.Nombre != "Maria Actualizada" {
		t.Errorf("Nombre = %q, esperado %q", resp.Nombre, "Maria Actualizada")
	}
	if resp.Email != "maria.nueva@email.com" {
		t.Errorf("Email = %q, esperado %q", resp.Email, "maria.nueva@email.com")
	}
}

func TestUpdateCliente_EmailDuplicado(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	// Intentar cambiar email de cliente 1 al email del cliente 2.
	_, err := svc.UpdateCliente(ctx, 1, application.UpdateClienteCommand{
		Email: strPtr("maria.garcia@email.com"),
	})
	if err != model.ErrEmailDuplicated {
		t.Errorf("error = %v, esperado ErrEmailDuplicated", err)
	}
}

func TestUpdateCliente_NombreVacio(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	_, err := svc.UpdateCliente(ctx, 1, application.UpdateClienteCommand{
		Nombre: strPtr(""),
	})
	if err != model.ErrNombreRequerido {
		t.Errorf("error = %v, esperado ErrNombreRequerido", err)
	}
}

func TestUpdateCliente_NoExistente(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	_, err := svc.UpdateCliente(ctx, 999, application.UpdateClienteCommand{
		Nombre: strPtr("No existe"),
	})
	if err != model.ErrClienteNotFound {
		t.Errorf("error = %v, esperado ErrClienteNotFound", err)
	}
}

func TestUpdateCliente_Telefono(t *testing.T) {
	svc := newTestAppService()
	ctx := context.Background()

	resp, err := svc.UpdateCliente(ctx, 1, application.UpdateClienteCommand{
		Telefono: strPtr("3005555555"),
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if resp.Telefono == nil || *resp.Telefono != "3005555555" {
		t.Errorf("Telefono = %v, esperado %q", resp.Telefono, "3005555555")
	}
}
