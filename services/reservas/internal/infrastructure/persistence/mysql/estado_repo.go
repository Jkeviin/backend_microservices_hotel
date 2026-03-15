package mysql

import (
	"context"
	"database/sql"
	"time"

	"reservas/internal/domain/model"
	"reservas/internal/domain/repository"
)

// EstadoRepo implementa repository.EstadoReservaRepository con MySQL.
type EstadoRepo struct {
	db *sql.DB
}

// NewEstadoRepo crea un repositorio MySQL para estados de reserva.
func NewEstadoRepo(db *sql.DB) repository.EstadoReservaRepository {
	return &EstadoRepo{db: db}
}

// List retorna todos los estados de reserva.
func (r *EstadoRepo) List(ctx context.Context) ([]model.EstadoReserva, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `SELECT id_estado, nombre FROM Estados_Reserva ORDER BY id_estado`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var estados []model.EstadoReserva
	for rows.Next() {
		var e model.EstadoReserva
		if err := rows.Scan(&e.ID, &e.Nombre); err != nil {
			return nil, err
		}
		estados = append(estados, e)
	}

	return estados, rows.Err()
}

// GetByNombre obtiene un estado por su nombre.
func (r *EstadoRepo) GetByNombre(ctx context.Context, nombre string) (*model.EstadoReserva, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `SELECT id_estado, nombre FROM Estados_Reserva WHERE nombre = ?`

	var e model.EstadoReserva
	if err := r.db.QueryRowContext(ctx, q, nombre).Scan(&e.ID, &e.Nombre); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &e, nil
}
