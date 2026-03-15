package application

import (
	"context"
	"errors"
	"time"

	"pagos/internal/domain/model"
	"pagos/internal/domain/repository"
)

// ReservaInfo contiene la informacion minima de una reserva que necesita el servicio de pagos.
type ReservaInfo struct {
	ID          int
	IDEstado    int
	NombreEstado string
}

// ReservaChecker define las operaciones que Pagos necesita del dominio de Reservas.
type ReservaChecker interface {
	// GetReservaParaPago obtiene la informacion de una reserva para validar el pago.
	GetReservaParaPago(ctx context.Context, idReserva int) (*ReservaInfo, error)
	// ConfirmarReserva actualiza el estado de la reserva a 'confirmada'.
	ConfirmarReserva(ctx context.Context, idReserva int) error
}

// CreatePagoCommand contiene los datos necesarios para crear un pago.
type CreatePagoCommand struct {
	IDReserva      int    `json:"id_reserva"`
	Monto          float64 `json:"monto"`
	FechaPago      string `json:"fecha_pago"`
	IdempotencyKey string `json:"idempotency_key"`
}

// PagoResponse es la representacion de un pago en la capa de aplicacion.
type PagoResponse struct {
	ID         int     `json:"id"`
	IDReserva  int     `json:"id_reserva"`
	Monto      float64 `json:"monto"`
	FechaPago  string  `json:"fecha_pago"`
	EstadoPago string  `json:"estado_pago"`
	CreatedAt  string  `json:"created_at"`
}

// PagoApp orquesta los casos de uso del servicio de pagos.
type PagoApp struct {
	repo           repository.PagoRepository
	reservaChecker ReservaChecker
}

// NewPagoApp crea una nueva instancia de PagoApp.
func NewPagoApp(repo repository.PagoRepository, reservaChecker ReservaChecker) *PagoApp {
	return &PagoApp{
		repo:           repo,
		reservaChecker: reservaChecker,
	}
}

// CreatePago procesa la creacion de un pago con idempotencia.
func (a *PagoApp) CreatePago(ctx context.Context, cmd CreatePagoCommand) (*PagoResponse, error) {
	// 1. Check idempotency: si ya existe un pago con esta key, retornarlo
	if cmd.IdempotencyKey != "" {
		existing, err := a.repo.GetByIdempotencyKey(ctx, cmd.IdempotencyKey)
		if err != nil && !errors.Is(err, model.ErrPagoNotFound) {
			return nil, err
		}
		if existing != nil {
			return toPagoResponse(existing), nil
		}
	}

	// 2. Validar monto > 0
	if err := model.ValidarMonto(cmd.Monto); err != nil {
		return nil, err
	}

	// 3. Verificar que la reserva existe y tiene estado valido
	reserva, err := a.reservaChecker.GetReservaParaPago(ctx, cmd.IDReserva)
	if err != nil {
		return nil, err
	}
	if reserva.NombreEstado != "pendiente_pago" && reserva.NombreEstado != "confirmada" {
		return nil, model.ErrReservaNoValida
	}

	// 4. Parsear fecha de pago
	fechaPago, err := time.Parse("2006-01-02", cmd.FechaPago)
	if err != nil {
		fechaPago = time.Now()
	}

	// 5. Crear pago con estado 'aprobado'
	key := cmd.IdempotencyKey
	pago := &model.Pago{
		IDReserva:      cmd.IDReserva,
		Monto:          cmd.Monto,
		FechaPago:      fechaPago,
		EstadoPago:     model.EstadoPagoAprobado,
		IdempotencyKey: &key,
		CreatedAt:      time.Now(),
	}

	if err := a.repo.Create(ctx, pago); err != nil {
		return nil, err
	}

	// 6. Confirmar reserva (actualizar estado a 'confirmada')
	if err := a.reservaChecker.ConfirmarReserva(ctx, cmd.IDReserva); err != nil {
		// El pago ya se creo; logear el error pero no fallar el pago.
		// En produccion esto seria compensable/eventual.
		return toPagoResponse(pago), nil
	}

	return toPagoResponse(pago), nil
}

// GetPago obtiene un pago por su ID.
func (a *PagoApp) GetPago(ctx context.Context, id int) (*PagoResponse, error) {
	pago, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toPagoResponse(pago), nil
}

func toPagoResponse(p *model.Pago) *PagoResponse {
	return &PagoResponse{
		ID:         p.ID,
		IDReserva:  p.IDReserva,
		Monto:      p.Monto,
		FechaPago:  p.FechaPago.Format("2006-01-02"),
		EstadoPago: p.EstadoPago.String(),
		CreatedAt:  p.CreatedAt.Format(time.RFC3339),
	}
}
