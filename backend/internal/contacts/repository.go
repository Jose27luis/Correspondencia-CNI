package contacts

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

func (r *Repositorio) Crear(ctx context.Context, entrada EntradaContacto) (Contacto, error) {
	fila, err := r.consultas.CrearContacto(ctx, sqlcgen.CrearContactoParams{
		Nombre:      entrada.Nombre,
		Empresa:     entrada.Empresa,
		Correo:      entrada.Correo,
		Pais:        entrada.Pais,
		CamposExtra: entrada.CamposExtra,
	})
	if err != nil {
		return Contacto{}, traducirError(err)
	}
	return desdeFila(fila), nil
}

func (r *Repositorio) Obtener(ctx context.Context, id uuid.UUID) (Contacto, error) {
	fila, err := r.consultas.ObtenerContacto(ctx, id)
	if err != nil {
		return Contacto{}, traducirError(err)
	}
	return desdeFila(fila), nil
}

func (r *Repositorio) Listar(ctx context.Context, filtro FiltroListado) (ListadoContactos, error) {
	filas, err := r.consultas.ListarContactos(ctx, sqlcgen.ListarContactosParams{
		Limit:    filtro.Limite,
		Offset:   filtro.Desfase,
		Busqueda: filtro.Busqueda,
	})
	if err != nil {
		return ListadoContactos{}, fmt.Errorf("no se pudieron listar los contactos: %w", err)
	}

	total, err := r.consultas.ContarContactos(ctx, filtro.Busqueda)
	if err != nil {
		return ListadoContactos{}, fmt.Errorf("no se pudieron contar los contactos: %w", err)
	}

	datos := make([]Contacto, 0, len(filas))
	for _, fila := range filas {
		datos = append(datos, desdeFila(fila))
	}

	return ListadoContactos{Datos: datos, Total: total}, nil
}

func (r *Repositorio) Actualizar(ctx context.Context, id uuid.UUID, entrada EntradaContacto) (Contacto, error) {
	fila, err := r.consultas.ActualizarContacto(ctx, sqlcgen.ActualizarContactoParams{
		ID:          id,
		Nombre:      entrada.Nombre,
		Empresa:     entrada.Empresa,
		Correo:      entrada.Correo,
		Pais:        entrada.Pais,
		CamposExtra: entrada.CamposExtra,
	})
	if err != nil {
		return Contacto{}, traducirError(err)
	}
	return desdeFila(fila), nil
}

func (r *Repositorio) Eliminar(ctx context.Context, id uuid.UUID) error {
	afectadas, err := r.consultas.EliminarContacto(ctx, id)
	if err != nil {
		return fmt.Errorf("no se pudo eliminar el contacto: %w", err)
	}
	if afectadas == 0 {
		return ErrNoEncontrado
	}
	return nil
}

func (r *Repositorio) Importar(ctx context.Context, entrada EntradaContacto) (Contacto, bool, error) {
	fila, err := r.consultas.ImportarContacto(ctx, sqlcgen.ImportarContactoParams{
		Nombre:      entrada.Nombre,
		Empresa:     entrada.Empresa,
		Correo:      entrada.Correo,
		Pais:        entrada.Pais,
		CamposExtra: entrada.CamposExtra,
	})
	if err != nil {
		return Contacto{}, false, traducirError(err)
	}

	contacto := desdeFila(sqlcgen.Contacto{
		ID:            fila.ID,
		Nombre:        fila.Nombre,
		Empresa:       fila.Empresa,
		Correo:        fila.Correo,
		Pais:          fila.Pais,
		CamposExtra:   fila.CamposExtra,
		CreadoEn:      fila.CreadoEn,
		ActualizadoEn: fila.ActualizadoEn,
	})

	return contacto, fila.FueCreado, nil
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
