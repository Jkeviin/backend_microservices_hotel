package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"inventario/internal/domain/model"
)

// HabitacionRepo es la implementacion MySQL del repositorio de habitaciones.
type HabitacionRepo struct {
	db *sqlx.DB
}

// NewHabitacionRepo crea un nuevo repositorio MySQL de habitaciones.
func NewHabitacionRepo(db *sqlx.DB) *HabitacionRepo {
	return &HabitacionRepo{db: db}
}

// habitacionRow es la estructura de mapeo para la tabla Habitaciones.
type habitacionRow struct {
	ID               int        `db:"id_habitacion"`
	Numero           string     `db:"numero"`
	IDTipoHabitacion int        `db:"id_tipo_habitacion"`
	Estado           string     `db:"estado"`
	UpdatedAt        *time.Time `db:"updated_at"`
}

// List retorna habitaciones filtradas. Construye la query dinamicamente segun los filtros.
func (r *HabitacionRepo) List(ctx context.Context, filtros model.HabitacionFiltros) ([]model.Habitacion, error) {
	var (
		conditions []string
		args       []interface{}
	)

	baseQuery := `SELECT id_habitacion, numero, id_tipo_habitacion, estado, updated_at FROM Habitaciones`

	if filtros.Tipo != nil {
		conditions = append(conditions, "id_tipo_habitacion = ?")
		args = append(args, *filtros.Tipo)
	}

	if filtros.Estado != nil {
		conditions = append(conditions, "estado = ?")
		args = append(args, filtros.Estado.String())
	}

	// Subquery NOT IN para verificar disponibilidad contra la tabla Reservas.
	if filtros.DisponibleDesde != nil && filtros.DisponibleHasta != nil {
		conditions = append(conditions, `id_habitacion NOT IN (
			SELECT id_habitacion FROM Reservas r
			JOIN Estados_Reserva er ON r.id_estado = er.id_estado
			WHERE r.fecha_inicio < ? AND r.fecha_fin > ?
			AND er.nombre NOT IN ('cancelada')
		)`)
		args = append(args, *filtros.DisponibleHasta, *filtros.DisponibleDesde)
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY id_habitacion"

	var rows []habitacionRow
	if err := r.db.SelectContext(ctx, &rows, baseQuery, args...); err != nil {
		return nil, fmt.Errorf("error listando habitaciones: %w", err)
	}

	return toHabitaciones(rows)
}

// GetByID retorna una habitacion por su identificador.
func (r *HabitacionRepo) GetByID(ctx context.Context, id int) (*model.Habitacion, error) {
	const query = `SELECT id_habitacion, numero, id_tipo_habitacion, estado, updated_at
	               FROM Habitaciones WHERE id_habitacion = ?`

	var row habitacionRow
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrHabitacionNotFound
		}
		return nil, fmt.Errorf("error obteniendo habitacion %d: %w", id, err)
	}

	h, err := toHabitacion(row)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// Update persiste los cambios de una habitacion existente.
func (r *HabitacionRepo) Update(ctx context.Context, h *model.Habitacion) error {
	const query = `UPDATE Habitaciones
	               SET estado = ?, id_tipo_habitacion = ?, updated_at = ?
	               WHERE id_habitacion = ?`

	result, err := r.db.ExecContext(ctx, query, h.Estado.String(), h.IDTipoHabitacion, h.UpdatedAt, h.ID)
	if err != nil {
		return fmt.Errorf("error actualizando habitacion %d: %w", h.ID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error verificando actualizacion: %w", err)
	}
	if rows == 0 {
		return model.ErrHabitacionNotFound
	}

	return nil
}

// ---------------------------------------------------------------------------
// Mappers
// ---------------------------------------------------------------------------

func toHabitaciones(rows []habitacionRow) ([]model.Habitacion, error) {
	result := make([]model.Habitacion, 0, len(rows))
	for _, row := range rows {
		h, err := toHabitacion(row)
		if err != nil {
			return nil, err
		}
		result = append(result, h)
	}
	return result, nil
}

func toHabitacion(row habitacionRow) (model.Habitacion, error) {
	numero, err := model.NewNumeroHabitacion(row.Numero)
	if err != nil {
		return model.Habitacion{}, fmt.Errorf("dato corrupto en habitacion %d: %w", row.ID, err)
	}

	estado, err := model.NewEstadoHabitacion(row.Estado)
	if err != nil {
		return model.Habitacion{}, fmt.Errorf("dato corrupto en habitacion %d: %w", row.ID, err)
	}

	return model.Habitacion{
		ID:               row.ID,
		Numero:           numero,
		IDTipoHabitacion: row.IDTipoHabitacion,
		Estado:           estado,
		UpdatedAt:        row.UpdatedAt,
	}, nil
}
