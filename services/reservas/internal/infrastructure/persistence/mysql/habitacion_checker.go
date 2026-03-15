package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	domainservice "reservas/internal/domain/service"
)

// HabitacionChecker implementa service.HabitacionChecker con lectura
// cruzada a la tabla Habitaciones (misma BD).
type HabitacionChecker struct {
	db *sql.DB
}

// NewHabitacionChecker crea un checker de habitaciones con acceso MySQL.
func NewHabitacionChecker(db *sql.DB) domainservice.HabitacionChecker {
	return &HabitacionChecker{db: db}
}

// GetEstadoHabitacion consulta el estado de la habitacion directamente
// en la tabla Habitaciones.
func (c *HabitacionChecker) GetEstadoHabitacion(ctx context.Context, id int) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `SELECT estado FROM Habitaciones WHERE id_habitacion = ?`

	var estado string
	if err := c.db.QueryRowContext(ctx, q, id).Scan(&estado); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("habitacion %d no encontrada", id)
		}
		return "", err
	}

	return estado, nil
}
