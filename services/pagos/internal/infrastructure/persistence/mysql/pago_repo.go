package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"pagos/internal/domain/model"
)

// PagoRepo implementa PagoRepository usando MySQL.
type PagoRepo struct {
	db *sqlx.DB
}

// NewPagoRepo crea un nuevo repositorio de pagos MySQL.
func NewPagoRepo(db *sqlx.DB) *PagoRepo {
	return &PagoRepo{db: db}
}

type pagoRow struct {
	ID             int            `db:"id_pago"`
	IDReserva      int            `db:"id_reserva"`
	Monto          float64        `db:"monto"`
	FechaPago      sql.NullTime   `db:"fecha_pago"`
	EstadoPago     string         `db:"estado_pago"`
	IdempotencyKey sql.NullString `db:"idempotency_key"`
	CreatedAt      sql.NullTime   `db:"created_at"`
}

func (r *pagoRow) toDomain() *model.Pago {
	p := &model.Pago{
		ID:         r.ID,
		IDReserva:  r.IDReserva,
		Monto:      r.Monto,
		EstadoPago: model.EstadoPagoEnum(r.EstadoPago),
	}
	if r.FechaPago.Valid {
		p.FechaPago = r.FechaPago.Time
	}
	if r.CreatedAt.Valid {
		p.CreatedAt = r.CreatedAt.Time
	}
	if r.IdempotencyKey.Valid {
		key := r.IdempotencyKey.String
		p.IdempotencyKey = &key
	}
	return p
}

// Create inserta un nuevo pago en la base de datos.
func (repo *PagoRepo) Create(ctx context.Context, pago *model.Pago) error {
	query := `INSERT INTO Pagos (id_reserva, monto, fecha_pago, estado_pago, idempotency_key, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	var idemKey sql.NullString
	if pago.IdempotencyKey != nil {
		idemKey = sql.NullString{String: *pago.IdempotencyKey, Valid: true}
	}

	result, err := repo.db.ExecContext(ctx, query,
		pago.IDReserva, pago.Monto, pago.FechaPago,
		string(pago.EstadoPago), idemKey, pago.CreatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	pago.ID = int(id)
	return nil
}

// GetByID obtiene un pago por su ID.
func (repo *PagoRepo) GetByID(ctx context.Context, id int) (*model.Pago, error) {
	var row pagoRow
	err := repo.db.GetContext(ctx, &row,
		`SELECT id_pago, id_reserva, monto, fecha_pago, estado_pago, idempotency_key, created_at
		FROM Pagos WHERE id_pago = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrPagoNotFound
		}
		return nil, err
	}
	return row.toDomain(), nil
}

// GetByIdempotencyKey busca un pago por su clave de idempotencia.
func (repo *PagoRepo) GetByIdempotencyKey(ctx context.Context, key string) (*model.Pago, error) {
	var row pagoRow
	err := repo.db.GetContext(ctx, &row,
		`SELECT id_pago, id_reserva, monto, fecha_pago, estado_pago, idempotency_key, created_at
		FROM Pagos WHERE idempotency_key = ?`, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrPagoNotFound
		}
		return nil, err
	}
	return row.toDomain(), nil
}
