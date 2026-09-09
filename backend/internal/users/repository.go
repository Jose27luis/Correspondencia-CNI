package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
)

const codigoViolacionUnicidad = "23505"

type Repositorio struct {
	consultas *sqlcgen.Queries
}

func NuevoRepositorio(pool *pgxpool.Pool) *Repositorio {
	return &Repositorio{consultas: sqlcgen.New(pool)}
}

func (r *Repositorio) Crear(ctx context.Context, nombre string, correo string, hash string) (Usuario, error) {
	fila, err := r.consultas.CrearUsuario(ctx, sqlcgen.CrearUsuarioParams{
		Nombre:       nombre,
		Correo:       correo,
		PasswordHash: hash,
	})
	if err != nil {
		return Usuario{}, traducirError(err)
	}
	return desdeFila(fila), nil
}

func (r *Repositorio) ObtenerPorCorreo(ctx context.Context, correo string) (sqlcgen.Usuario, error) {
	fila, err := r.consultas.ObtenerUsuarioPorCorreo(ctx, correo)
	if err != nil {
		return sqlcgen.Usuario{}, traducirError(err)
	}
	return fila, nil
}

func (r *Repositorio) Obtener(ctx context.Context, id uuid.UUID) (Usuario, error) {
	fila, err := r.consultas.ObtenerUsuario(ctx, id)
	if err != nil {
		return Usuario{}, traducirError(err)
	}
	return desdeFila(fila), nil
}

func (r *Repositorio) Contar(ctx context.Context) (int64, error) {
	total, err := r.consultas.ContarUsuarios(ctx)
	if err != nil {
		return 0, fmt.Errorf("no se pudieron contar los usuarios: %w", err)
	}
	return total, nil
}

func traducirError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoEncontrado
	}

	var errPg *pgconn.PgError
	if errors.As(err, &errPg) && errPg.Code == codigoViolacionUnicidad {
		return ErrCorreoEnUso
	}

	return fmt.Errorf("error de base de datos: %w", err)
}
