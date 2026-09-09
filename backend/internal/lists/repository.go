package lists

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
)

const codigoViolacionLlaveForanea = "23503"

type Repositorio struct {
	consultas *sqlcgen.Queries
}

func NuevoRepositorio(pool *pgxpool.Pool) *Repositorio {
	return &Repositorio{consultas: sqlcgen.New(pool)}
}

func (r *Repositorio) Crear(ctx context.Context, entrada EntradaLista) (Lista, error) {
	fila, err := r.consultas.CrearLista(ctx, sqlcgen.CrearListaParams{
		Nombre:      entrada.Nombre,
		Descripcion: entrada.Descripcion,
	})
	if err != nil {
		return Lista{}, traducirError(err)
	}
	return desdeFila(fila, 0), nil
}

func (r *Repositorio) Obtener(ctx context.Context, id uuid.UUID) (Lista, error) {
	fila, err := r.consultas.ObtenerLista(ctx, id)
	if err != nil {
		return Lista{}, traducirError(err)
	}

	return Lista{
		ID:             fila.ID,
		Nombre:         fila.Nombre,
		Descripcion:    fila.Descripcion,
		TotalContactos: fila.TotalContactos,
		CreadoEn:       fila.CreadoEn,
	}, nil
}

func (r *Repositorio) Listar(ctx context.Context, limite int32, desfase int32) (ListadoListas, error) {
	filas, err := r.consultas.ListarListas(ctx, sqlcgen.ListarListasParams{
		Limit:  limite,
		Offset: desfase,
	})
	if err != nil {
		return ListadoListas{}, fmt.Errorf("no se pudieron listar las listas: %w", err)
	}

	total, err := r.consultas.ContarListas(ctx)
	if err != nil {
		return ListadoListas{}, fmt.Errorf("no se pudieron contar las listas: %w", err)
	}

	datos := make([]Lista, 0, len(filas))
	for _, fila := range filas {
		datos = append(datos, Lista{
			ID:             fila.ID,
			Nombre:         fila.Nombre,
			Descripcion:    fila.Descripcion,
			TotalContactos: fila.TotalContactos,
			CreadoEn:       fila.CreadoEn,
		})
	}

	return ListadoListas{Datos: datos, Total: total}, nil
}

func (r *Repositorio) Actualizar(ctx context.Context, id uuid.UUID, entrada EntradaLista) (Lista, error) {
	fila, err := r.consultas.ActualizarLista(ctx, sqlcgen.ActualizarListaParams{
		ID:          id,
		Nombre:      entrada.Nombre,
		Descripcion: entrada.Descripcion,
	})
	if err != nil {
		return Lista{}, traducirError(err)
	}

	total, err := r.consultas.ContarContactosDeLista(ctx, id)
	if err != nil {
		return Lista{}, fmt.Errorf("no se pudieron contar los contactos de la lista: %w", err)
	}

	return desdeFila(fila, total), nil
}

func (r *Repositorio) Eliminar(ctx context.Context, id uuid.UUID) error {
	afectadas, err := r.consultas.EliminarLista(ctx, id)
	if err != nil {
		return fmt.Errorf("no se pudo eliminar la lista: %w", err)
	}
	if afectadas == 0 {
		return ErrNoEncontrada
	}
	return nil
}

func (r *Repositorio) Existe(ctx context.Context, id uuid.UUID) (bool, error) {
	existe, err := r.consultas.ExisteLista(ctx, id)
	if err != nil {
		return false, fmt.Errorf("no se pudo verificar la lista: %w", err)
	}
	return existe, nil
}

func (r *Repositorio) AgregarContactos(ctx context.Context, listaID uuid.UUID, contactoIDs []uuid.UUID) (int64, error) {
	agregados, err := r.consultas.AgregarContactosALista(ctx, sqlcgen.AgregarContactosAListaParams{
		ListaID:     listaID,
		ContactoIds: contactoIDs,
	})
	if err != nil {
		return 0, traducirError(err)
	}
	return agregados, nil
}

func (r *Repositorio) QuitarContacto(ctx context.Context, listaID uuid.UUID, contactoID uuid.UUID) error {
	afectadas, err := r.consultas.QuitarContactoDeLista(ctx, sqlcgen.QuitarContactoDeListaParams{
		ListaID:    listaID,
		ContactoID: contactoID,
	})
	if err != nil {
		return fmt.Errorf("no se pudo quitar el contacto de la lista: %w", err)
	}
	if afectadas == 0 {
		return ErrNoEncontrada
	}
	return nil
}

func (r *Repositorio) ListarMiembros(ctx context.Context, listaID uuid.UUID, limite int32, desfase int32) (ListadoMiembros, error) {
	filas, err := r.consultas.ListarContactosDeLista(ctx, sqlcgen.ListarContactosDeListaParams{
		ListaID: listaID,
		Limit:   limite,
		Offset:  desfase,
	})
	if err != nil {
		return ListadoMiembros{}, fmt.Errorf("no se pudieron listar los contactos de la lista: %w", err)
	}

	total, err := r.consultas.ContarContactosDeLista(ctx, listaID)
	if err != nil {
		return ListadoMiembros{}, fmt.Errorf("no se pudieron contar los contactos de la lista: %w", err)
	}

	datos := make([]contacts.Contacto, 0, len(filas))
	for _, fila := range filas {
		datos = append(datos, contacts.DesdeFila(fila))
	}

	return ListadoMiembros{Datos: datos, Total: total}, nil
}

func traducirError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoEncontrada
	}

	var errPg *pgconn.PgError
	if errors.As(err, &errPg) && errPg.Code == codigoViolacionLlaveForanea {
		if errPg.ConstraintName == "contacto_lista_lista_id_fkey" {
			return ErrNoEncontrada
		}
		return ErrContactoInvalido
	}

	return fmt.Errorf("error de base de datos: %w", err)
}
