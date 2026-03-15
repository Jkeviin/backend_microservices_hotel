package mysql

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"clientes/internal/domain/model"
	"clientes/internal/domain/repository"
)

const queryTimeout = 5 * time.Second

// ClienteRepo implementa repository.ClienteRepository con MySQL.
type ClienteRepo struct {
	db *sql.DB
}

// NewClienteRepo crea un repositorio MySQL para clientes.
func NewClienteRepo(db *sql.DB) repository.ClienteRepository {
	return &ClienteRepo{db: db}
}

// Create inserta un nuevo cliente. Detecta error 1062 (duplicate entry) para
// retornar model.ErrEmailDuplicated.
func (r *ClienteRepo) Create(ctx context.Context, c *model.Cliente) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const q = `INSERT INTO Clientes (nombre, email, telefono) VALUES (?, ?, ?)`

	result, err := r.db.ExecContext(ctx, q, c.Nombre, c.Email.String(), c.Telefono)
	if err != nil {
		if isDuplicateEntry(err) {
			return model.ErrEmailDuplicated
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = int(id)

	// Leer created_at generado por la BD.
	row := r.db.QueryRowContext(ctx,
		`SELECT created_at FROM Clientes WHERE id_cliente = ?`, c.ID)
	return row.Scan(&c.CreatedAt)
}

// GetByID obtiene un cliente por su ID.
func (r *ClienteRepo) GetByID(ctx context.Context, id int) (*model.Cliente, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const q = `SELECT id_cliente, nombre, email, telefono, created_at, updated_at
	            FROM Clientes WHERE id_cliente = ?`

	return r.scanOne(r.db.QueryRowContext(ctx, q, id))
}

// GetByEmail obtiene un cliente por su email.
func (r *ClienteRepo) GetByEmail(ctx context.Context, email string) (*model.Cliente, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const q = `SELECT id_cliente, nombre, email, telefono, created_at, updated_at
	            FROM Clientes WHERE email = ?`

	return r.scanOne(r.db.QueryRowContext(ctx, q, email))
}

// Update actualiza un cliente existente.
func (r *ClienteRepo) Update(ctx context.Context, c *model.Cliente) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const q = `UPDATE Clientes SET nombre = ?, email = ?, telefono = ?
	            WHERE id_cliente = ?`

	result, err := r.db.ExecContext(ctx, q, c.Nombre, c.Email.String(), c.Telefono, c.ID)
	if err != nil {
		if isDuplicateEntry(err) {
			return model.ErrEmailDuplicated
		}
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return model.ErrClienteNotFound
	}

	// Leer updated_at generado por la BD.
	row := r.db.QueryRowContext(ctx,
		`SELECT updated_at FROM Clientes WHERE id_cliente = ?`, c.ID)
	return row.Scan(&c.UpdatedAt)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *ClienteRepo) scanOne(row *sql.Row) (*model.Cliente, error) {
	var (
		id        int
		nombre    string
		emailStr  string
		telefono  *string
		createdAt time.Time
		updatedAt *time.Time
	)

	if err := row.Scan(&id, &nombre, &emailStr, &telefono, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrClienteNotFound
		}
		return nil, err
	}

	email, err := model.NewEmail(emailStr)
	if err != nil {
		return nil, err
	}

	return &model.Cliente{
		ID:        id,
		Nombre:    nombre,
		Email:     email,
		Telefono:  telefono,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

// isDuplicateEntry detecta el error MySQL 1062 (Duplicate entry).
func isDuplicateEntry(err error) bool {
	return strings.Contains(err.Error(), "1062") ||
		strings.Contains(err.Error(), "Duplicate entry")
}
