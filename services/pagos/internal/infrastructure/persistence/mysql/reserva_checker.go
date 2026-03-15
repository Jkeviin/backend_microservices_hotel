package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"pagos/internal/application"
	"pagos/internal/domain/model"
)

// ReservaChecker implementa application.ReservaChecker usando MySQL.
type ReservaChecker struct {
	db *sqlx.DB
}

// NewReservaChecker crea una nueva instancia de ReservaChecker MySQL.
func NewReservaChecker(db *sqlx.DB) *ReservaChecker {
	return &ReservaChecker{db: db}
}

// GetReservaParaPago obtiene la informacion de una reserva para validar el pago.
func (rc *ReservaChecker) GetReservaParaPago(ctx context.Context, idReserva int) (*application.ReservaInfo, error) {
	var info application.ReservaInfo
	err := rc.db.QueryRowContext(ctx,
		`SELECT r.id_reserva, r.id_estado, e.nombre
		FROM Reservas r
		JOIN Estados_Reserva e ON r.id_estado = e.id_estado
		WHERE r.id_reserva = ?`, idReserva,
	).Scan(&info.ID, &info.IDEstado, &info.NombreEstado)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrReservaNoValida
		}
		return nil, err
	}
	return &info, nil
}

// ConfirmarReserva actualiza el estado de la reserva a 'confirmada'.
func (rc *ReservaChecker) ConfirmarReserva(ctx context.Context, idReserva int) error {
	_, err := rc.db.ExecContext(ctx,
		`UPDATE Reservas SET id_estado = (
			SELECT id_estado FROM Estados_Reserva WHERE nombre = 'confirmada'
		) WHERE id_reserva = ?`, idReserva)
	return err
}
