package mysql

import (
	"context"

	"github.com/jmoiron/sqlx"

	"inventario/internal/domain/model"
)

// TipoHabitacionRepo es la implementacion MySQL del repositorio de tipos de habitacion.
type TipoHabitacionRepo struct {
	db *sqlx.DB
}

// NewTipoHabitacionRepo crea un nuevo repositorio MySQL de tipos de habitacion.
func NewTipoHabitacionRepo(db *sqlx.DB) *TipoHabitacionRepo {
	return &TipoHabitacionRepo{db: db}
}

// tipoHabitacionRow es la estructura de mapeo para la tabla Tipos_Habitacion.
type tipoHabitacionRow struct {
	ID         int     `db:"id_tipo"`
	Nombre     string  `db:"nombre"`
	PrecioBase float64 `db:"precio_base"`
	Capacidad  int     `db:"capacidad"`
}

// List retorna todos los tipos de habitacion.
func (r *TipoHabitacionRepo) List(ctx context.Context) ([]model.TipoHabitacion, error) {
	const query = `SELECT id_tipo, nombre, precio_base, capacidad FROM Tipos_Habitacion ORDER BY id_tipo`

	var rows []tipoHabitacionRow
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, err
	}

	result := make([]model.TipoHabitacion, 0, len(rows))
	for _, row := range rows {
		result = append(result, model.TipoHabitacion{
			ID:         row.ID,
			Nombre:     row.Nombre,
			PrecioBase: row.PrecioBase,
			Capacidad:  row.Capacidad,
		})
	}
	return result, nil
}
