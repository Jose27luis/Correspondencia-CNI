package suppression

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
)

type Repositorio struct {
	consultas *sqlcgen.Queries
}

func NuevoRepositorio(pool *pgxpool.Pool) *Repositorio {
	return &Repositorio{consultas: sqlcgen.New(pool)}
}

func (r *Repositorio) Suprimir(ctx context.Context, correo string, motivo string, detalle string) error {
	var texto *string
	if detalle != "" {
		texto = &detalle
	}

	if err := r.consultas.SuprimirCorreo(ctx, sqlcgen.SuprimirCorreoParams{
		Correo:  correo,
		Motivo:  motivo,
		Detalle: texto,
	}); err != nil {
		return fmt.Errorf("no se pudo excluir el correo: %w", err)
	}

	return nil
}

func (r *Repositorio) SuprimirPorEnvio(ctx context.Context, envioID uuid.UUID, motivo string, detalle string) error {
	var texto *string
	if detalle != "" {
		texto = &detalle
	}

	if err := r.consultas.SuprimirPorEnvio(ctx, sqlcgen.SuprimirPorEnvioParams{
		ID:      envioID,
		Motivo:  motivo,
		Detalle: texto,
	}); err != nil {
		return fmt.Errorf("no se pudo excluir el correo del envío: %w", err)
	}

	return nil
}

func (r *Repositorio) Listar(ctx context.Context, limite int32, desfase int32) (ListadoSupresiones, error) {
	filas, err := r.consultas.ListarSuprimidos(ctx, sqlcgen.ListarSuprimidosParams{
		Limit:  limite,
		Offset: desfase,
	})
	if err != nil {
		return ListadoSupresiones{}, fmt.Errorf("no se pudieron listar las exclusiones: %w", err)
	}

	total, err := r.consultas.ContarSuprimidos(ctx)
	if err != nil {
		return ListadoSupresiones{}, fmt.Errorf("no se pudieron contar las exclusiones: %w", err)
	}

	datos := make([]Supresion, 0, len(filas))
	for _, fila := range filas {
		datos = append(datos, Supresion{
			Correo:   fila.Correo,
			Motivo:   fila.Motivo,
			Detalle:  fila.Detalle,
			CreadoEn: fila.CreadoEn,
		})
	}

	return ListadoSupresiones{Datos: datos, Total: total}, nil
}

func (r *Repositorio) Quitar(ctx context.Context, correo string) error {
	afectadas, err := r.consultas.QuitarSupresion(ctx, correo)
	if err != nil {
		return fmt.Errorf("no se pudo quitar la exclusión: %w", err)
	}

	if afectadas == 0 {
		return ErrNoEncontrado
	}

	return nil
}

func (r *Repositorio) ContarEnLista(ctx context.Context, listaID uuid.UUID) (int64, error) {
	total, err := r.consultas.ContarSuprimidosDeLista(ctx, listaID)
	if err != nil {
		return 0, fmt.Errorf("no se pudieron contar las exclusiones de la lista: %w", err)
	}
	return total, nil
}

func (r *Repositorio) ContactoPorID(ctx context.Context, id uuid.UUID) (contacts.Contacto, error) {
	fila, err := r.consultas.ObtenerContactoPorID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return contacts.Contacto{}, ErrTokenInvalido
		}
		return contacts.Contacto{}, fmt.Errorf("no se pudo obtener el contacto: %w", err)
	}

	return contacts.DesdeFila(fila), nil
}
