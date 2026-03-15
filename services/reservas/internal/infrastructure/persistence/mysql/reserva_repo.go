package mysql

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"reservas/internal/domain/model"
	"reservas/internal/domain/repository"
)

const queryTimeout = 5 * time.Second

// ReservaRepo implementa repository.ReservaRepository con MySQL.
type ReservaRepo struct {
	db *sql.DB
}

// NewReservaRepo crea un repositorio MySQL para reservas.
func NewReservaRepo(db *sql.DB) repository.ReservaRepository {
	return &ReservaRepo{db: db}
}

// Create inserta una nueva reserva.
func (r *ReservaRepo) Create(ctx context.Context, res *model.Reserva) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const q = `INSERT INTO Reservas (id_cliente, id_habitacion, id_estado, fecha_inicio, fecha_fin, total, idempotency_key, version)
	            VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, q,
		res.IDCliente, res.IDHabitacion, res.IDEstado,
		res.FechaInicio, res.FechaFin, res.Total,
		res.IdempotencyKey, res.Version,
	)
	if err != nil {
		if isDuplicateEntry(err) {
			return model.ErrIdempotencyKeyDuplicada
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	res.ID = int(id)

	// Leer created_at generado por la BD.
	row := r.db.QueryRowContext(ctx,
		`SELECT created_at FROM Reservas WHERE id_reserva = ?`, res.ID)
	return row.Scan(&res.CreatedAt)
}

// GetByID obtiene una reserva por su ID con JOIN a Estados_Reserva.
func (r *ReservaRepo) GetByID(ctx context.Context, id int) (*model.Reserva, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const q = `SELECT r.id_reserva, r.id_cliente, r.id_habitacion, r.id_estado, e.nombre,
	                   r.fecha_inicio, r.fecha_fin, r.total, r.idempotency_key,
	                   r.version, r.created_at, r.updated_at
	            FROM Reservas r
	            JOIN Estados_Reserva e ON r.id_estado = e.id_estado
	            WHERE r.id_reserva = ?`

	return r.scanOne(r.db.QueryRowContext(ctx, q, id))
}

// Update actualiza una reserva con optimistic locking.
func (r *ReservaRepo) Update(ctx context.Context, res *model.Reserva) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const q = `UPDATE Reservas
	            SET id_estado = ?, fecha_inicio = ?, fecha_fin = ?, total = ?, version = version + 1
	            WHERE id_reserva = ? AND version = ?`

	result, err := r.db.ExecContext(ctx, q,
		res.IDEstado, res.FechaInicio, res.FechaFin, res.Total,
		res.ID, res.Version,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return model.ErrVersionConflict
	}

	// Actualizar version en el objeto.
	res.Version++

	// Leer updated_at generado por la BD.
	row := r.db.QueryRowContext(ctx,
		`SELECT updated_at FROM Reservas WHERE id_reserva = ?`, res.ID)
	return row.Scan(&res.UpdatedAt)
}

// ExisteSolapamiento verifica si hay reservas activas que se solapan.
func (r *ReservaRepo) ExisteSolapamiento(ctx context.Context, idHabitacion int, inicio, fin time.Time, excludeReservaID *int) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	// Estados activos: pendiente_pago(1), confirmada(2), completada(4)
	q := `SELECT COUNT(*) FROM Reservas
	       WHERE id_habitacion = ?
	         AND fecha_inicio < ?
	         AND fecha_fin > ?
	         AND id_estado IN (1, 2, 4)`

	args := []interface{}{idHabitacion, fin, inicio}

	if excludeReservaID != nil {
		q += ` AND id_reserva != ?`
		args = append(args, *excludeReservaID)
	}

	var count int
	if err := r.db.QueryRowContext(ctx, q, args...).Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetByIdempotencyKey busca una reserva por su clave de idempotencia.
func (r *ReservaRepo) GetByIdempotencyKey(ctx context.Context, key string) (*model.Reserva, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const q = `SELECT r.id_reserva, r.id_cliente, r.id_habitacion, r.id_estado, e.nombre,
	                   r.fecha_inicio, r.fecha_fin, r.total, r.idempotency_key,
	                   r.version, r.created_at, r.updated_at
	            FROM Reservas r
	            JOIN Estados_Reserva e ON r.id_estado = e.id_estado
	            WHERE r.idempotency_key = ?`

	return r.scanOne(r.db.QueryRowContext(ctx, q, key))
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *ReservaRepo) scanOne(row *sql.Row) (*model.Reserva, error) {
	var (
		id             int
		idCliente      int
		idHabitacion   int
		idEstado       int
		nombreEstado   string
		fechaInicio    time.Time
		fechaFin       time.Time
		total          *float64
		idempotencyKey *string
		version        int
		createdAt      time.Time
		updatedAt      *time.Time
	)

	if err := row.Scan(
		&id, &idCliente, &idHabitacion, &idEstado, &nombreEstado,
		&fechaInicio, &fechaFin, &total, &idempotencyKey,
		&version, &createdAt, &updatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrReservaNotFound
		}
		return nil, err
	}

	return &model.Reserva{
		ID:             id,
		IDCliente:      idCliente,
		IDHabitacion:   idHabitacion,
		IDEstado:       idEstado,
		NombreEstado:   nombreEstado,
		FechaInicio:    fechaInicio,
		FechaFin:       fechaFin,
		Total:          total,
		IdempotencyKey: idempotencyKey,
		Version:        version,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}

// isDuplicateEntry detecta el error MySQL 1062 (Duplicate entry).
func isDuplicateEntry(err error) bool {
	return strings.Contains(err.Error(), "1062") ||
		strings.Contains(err.Error(), "Duplicate entry")
}
