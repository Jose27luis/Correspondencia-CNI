package correspondence

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

func (r *Repositorio) Crear(ctx context.Context, usuarioID uuid.UUID, entrada EntradaCorrespondencia) (Correspondencia, error) {
	fila, err := r.consultas.CrearCorrespondencia(ctx, sqlcgen.CrearCorrespondenciaParams{
		UsuarioID: usuarioID,
		ListaID:   entrada.ListaID,
		Asunto:    entrada.Asunto,
		Cuerpo:    entrada.Cuerpo,
	})
	if err != nil {
		return Correspondencia{}, traducirError(err)
	}
	return desdeFila(fila, []Adjunto{}), nil
}

func (r *Repositorio) Obtener(ctx context.Context, id uuid.UUID) (Correspondencia, error) {
	fila, err := r.consultas.ObtenerCorrespondencia(ctx, id)
	if err != nil {
		return Correspondencia{}, traducirError(err)
	}

	adjuntos, err := r.ListarAdjuntos(ctx, id)
	if err != nil {
		return Correspondencia{}, err
	}

	return desdeFila(fila, adjuntos), nil
}

func (r *Repositorio) Listar(ctx context.Context, estado *string, limite int32, desfase int32) (ListadoCorrespondencia, error) {
	filas, err := r.consultas.ListarCorrespondencia(ctx, sqlcgen.ListarCorrespondenciaParams{
		Limit:  limite,
		Offset: desfase,
		Estado: estado,
	})
	if err != nil {
		return ListadoCorrespondencia{}, fmt.Errorf("no se pudo listar la correspondencia: %w", err)
	}

	total, err := r.consultas.ContarCorrespondencia(ctx, estado)
	if err != nil {
		return ListadoCorrespondencia{}, fmt.Errorf("no se pudo contar la correspondencia: %w", err)
	}

	datos := make([]Correspondencia, 0, len(filas))
	for _, fila := range filas {
		datos = append(datos, desdeFila(fila, []Adjunto{}))
	}

	return ListadoCorrespondencia{Datos: datos, Total: total}, nil
}

func (r *Repositorio) Actualizar(ctx context.Context, id uuid.UUID, entrada EntradaCorrespondencia) (Correspondencia, error) {
	fila, err := r.consultas.ActualizarCorrespondencia(ctx, sqlcgen.ActualizarCorrespondenciaParams{
		ID:      id,
		ListaID: entrada.ListaID,
		Asunto:  entrada.Asunto,
		Cuerpo:  entrada.Cuerpo,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Correspondencia{}, r.distinguirNoEncontrada(ctx, id)
		}
		return Correspondencia{}, traducirError(err)
	}

	adjuntos, err := r.ListarAdjuntos(ctx, id)
	if err != nil {
		return Correspondencia{}, err
	}

	return desdeFila(fila, adjuntos), nil
}

func (r *Repositorio) CambiarEstado(ctx context.Context, id uuid.UUID, estado string) (Correspondencia, error) {
	fila, err := r.consultas.CambiarEstadoCorrespondencia(ctx, sqlcgen.CambiarEstadoCorrespondenciaParams{
		ID:     id,
		Estado: estado,
	})
	if err != nil {
		return Correspondencia{}, traducirError(err)
	}
	return desdeFila(fila, []Adjunto{}), nil
}

func (r *Repositorio) Eliminar(ctx context.Context, id uuid.UUID) error {
	afectadas, err := r.consultas.EliminarCorrespondencia(ctx, id)
	if err != nil {
		return fmt.Errorf("no se pudo eliminar la correspondencia: %w", err)
	}
	if afectadas == 0 {
		return r.distinguirNoEncontrada(ctx, id)
	}
	return nil
}

func (r *Repositorio) ListarAdjuntos(ctx context.Context, correspondenciaID uuid.UUID) ([]Adjunto, error) {
	filas, err := r.consultas.ListarAdjuntos(ctx, correspondenciaID)
	if err != nil {
		return nil, fmt.Errorf("no se pudieron listar los adjuntos: %w", err)
	}

	adjuntos := make([]Adjunto, 0, len(filas))
	for _, fila := range filas {
		adjuntos = append(adjuntos, adjuntoDesdeFila(fila))
	}

	return adjuntos, nil
}

func (r *Repositorio) CrearAdjunto(ctx context.Context, correspondenciaID uuid.UUID, entrada EntradaAdjunto) (Adjunto, error) {
	fila, err := r.consultas.CrearAdjunto(ctx, sqlcgen.CrearAdjuntoParams{
		CorrespondenciaID: correspondenciaID,
		NombreArchivo:     entrada.NombreArchivo,
		UrlArchivo:        entrada.UrlArchivo,
		Tipo:              entrada.Tipo,
		TamanoBytes:       entrada.TamanoBytes,
	})
	if err != nil {
		return Adjunto{}, traducirError(err)
	}
	return adjuntoDesdeFila(fila), nil
}

func (r *Repositorio) EliminarAdjunto(ctx context.Context, correspondenciaID uuid.UUID, adjuntoID uuid.UUID) error {
	afectadas, err := r.consultas.EliminarAdjunto(ctx, sqlcgen.EliminarAdjuntoParams{
		ID:                adjuntoID,
		CorrespondenciaID: correspondenciaID,
	})
	if err != nil {
		return fmt.Errorf("no se pudo eliminar el adjunto: %w", err)
	}
	if afectadas == 0 {
		return ErrAdjuntoInvalido
	}
	return nil
}

func (r *Repositorio) TamanoAdjuntos(ctx context.Context, correspondenciaID uuid.UUID) (int64, error) {
	total, err := r.consultas.SumarTamanoAdjuntos(ctx, correspondenciaID)
	if err != nil {
		return 0, fmt.Errorf("no se pudo calcular el tamaño de los adjuntos: %w", err)
	}
	return total, nil
}

func (r *Repositorio) PrimerContactoDeLista(ctx context.Context, listaID uuid.UUID) (contacts.Contacto, error) {
	fila, err := r.consultas.ObtenerPrimerContactoDeLista(ctx, listaID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return contacts.Contacto{}, ErrSinDestinatario
		}
		return contacts.Contacto{}, fmt.Errorf("no se pudo obtener el destinatario: %w", err)
	}
	return contacts.DesdeFila(fila), nil
}

func (r *Repositorio) distinguirNoEncontrada(ctx context.Context, id uuid.UUID) error {
	if _, err := r.consultas.ObtenerCorrespondencia(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoEncontrada
		}
		return fmt.Errorf("error de base de datos: %w", err)
	}
	return ErrNoEsBorrador
}

func traducirError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoEncontrada
	}

	var errPg *pgconn.PgError
	if errors.As(err, &errPg) && errPg.Code == codigoViolacionLlaveForanea {
		if errPg.ConstraintName == "correspondencia_lista_id_fkey" {
			return ErrListaInvalida
		}
		return ErrNoEncontrada
	}

	return fmt.Errorf("error de base de datos: %w", err)
}
