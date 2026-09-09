package dispatches

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

func (r *Repositorio) CrearEnviosDeLista(ctx context.Context, correspondenciaID uuid.UUID, listaID uuid.UUID) (int64, error) {
	creados, err := r.consultas.CrearEnviosDeLista(ctx, sqlcgen.CrearEnviosDeListaParams{
		CorrespondenciaID: correspondenciaID,
		ListaID:           listaID,
	})
	if err != nil {
		return 0, fmt.Errorf("no se pudieron crear los envíos: %w", err)
	}
	return creados, nil
}

func (r *Repositorio) IDsPendientes(ctx context.Context, correspondenciaID uuid.UUID) ([]uuid.UUID, error) {
	filas, err := r.consultas.ListarEnviosPendientes(ctx, correspondenciaID)
	if err != nil {
		return nil, fmt.Errorf("no se pudieron listar los envíos pendientes: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(filas))
	for _, fila := range filas {
		ids = append(ids, fila.ID)
	}

	return ids, nil
}

func (r *Repositorio) DatosDeEnvio(ctx context.Context, envioID uuid.UUID) (uuid.UUID, contacts.Contacto, string, error) {
	fila, err := r.consultas.ObtenerEnvioConDatos(ctx, envioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, contacts.Contacto{}, "", ErrNoEncontrado
		}
		return uuid.Nil, contacts.Contacto{}, "", fmt.Errorf("no se pudo obtener el envío: %w", err)
	}

	contacto := contacts.DesdeFila(sqlcgen.Contacto{
		ID:            fila.ContactoID,
		Nombre:        fila.Nombre,
		Empresa:       fila.Empresa,
		Correo:        fila.Correo,
		Pais:          fila.Pais,
		CamposExtra:   fila.CamposExtra,
		CreadoEn:      fila.ContactoCreadoEn,
		ActualizadoEn: fila.ContactoActualizadoEn,
	})

	return fila.CorrespondenciaID, contacto, fila.Estado, nil
}

func (r *Repositorio) MarcarEnviado(ctx context.Context, envioID uuid.UUID, mensajeID string) error {
	if _, err := r.consultas.MarcarEnvioEnviado(ctx, sqlcgen.MarcarEnvioEnviadoParams{
		ID:                 envioID,
		ProveedorMensajeID: &mensajeID,
	}); err != nil {
		return fmt.Errorf("no se pudo marcar el envío como enviado: %w", err)
	}
	return nil
}

func (r *Repositorio) MarcarFallido(ctx context.Context, envioID uuid.UUID) error {
	if _, err := r.consultas.MarcarEnvioFallido(ctx, envioID); err != nil {
		return fmt.Errorf("no se pudo marcar el envío como fallido: %w", err)
	}
	return nil
}

func (r *Repositorio) ActualizarPorMensaje(ctx context.Context, mensajeID string, estado string) (uuid.UUID, error) {
	envioID, err := r.consultas.ActualizarEstadoPorMensaje(ctx, sqlcgen.ActualizarEstadoPorMensajeParams{
		ProveedorMensajeID: &mensajeID,
		Estado:             estado,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNoEncontrado
		}
		return uuid.Nil, fmt.Errorf("no se pudo actualizar el envío: %w", err)
	}
	return envioID, nil
}

func (r *Repositorio) BuscarPorMensaje(ctx context.Context, mensajeID string) (uuid.UUID, error) {
	envioID, err := r.consultas.ObtenerEnvioPorMensaje(ctx, &mensajeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNoEncontrado
		}
		return uuid.Nil, fmt.Errorf("no se pudo ubicar el envío: %w", err)
	}
	return envioID, nil
}

func (r *Repositorio) RegistrarEvento(ctx context.Context, envioID uuid.UUID, tipo string, fecha time.Time, payload json.RawMessage) error {
	if len(payload) == 0 {
		payload = json.RawMessage("{}")
	}

	if _, err := r.consultas.RegistrarEventoEnvio(ctx, sqlcgen.RegistrarEventoEnvioParams{
		EnvioID: envioID,
		Tipo:    tipo,
		Fecha:   fecha,
		Payload: payload,
	}); err != nil {
		return fmt.Errorf("no se pudo registrar el evento: %w", err)
	}

	return nil
}

func (r *Repositorio) CerrarSiTermino(ctx context.Context, correspondenciaID uuid.UUID) (string, error) {
	estado, err := r.consultas.CerrarCorrespondenciaSiTermino(ctx, correspondenciaID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("no se pudo cerrar la correspondencia: %w", err)
	}
	return estado, nil
}

func (r *Repositorio) Listar(ctx context.Context, correspondenciaID uuid.UUID, estado *string, limite int32, desfase int32) (ListadoEnvios, error) {
	filas, err := r.consultas.ListarEnviosDeCorrespondencia(ctx, sqlcgen.ListarEnviosDeCorrespondenciaParams{
		CorrespondenciaID: correspondenciaID,
		Estado:            estado,
		Limit:             limite,
		Offset:            desfase,
	})
	if err != nil {
		return ListadoEnvios{}, fmt.Errorf("no se pudieron listar los envíos: %w", err)
	}

	total, err := r.consultas.ContarEnviosDeCorrespondencia(ctx, sqlcgen.ContarEnviosDeCorrespondenciaParams{
		CorrespondenciaID: correspondenciaID,
		Estado:            estado,
	})
	if err != nil {
		return ListadoEnvios{}, fmt.Errorf("no se pudieron contar los envíos: %w", err)
	}

	datos := make([]Envio, 0, len(filas))
	for _, fila := range filas {
		datos = append(datos, Envio{
			ID:                 fila.ID,
			Estado:             fila.Estado,
			ProveedorMensajeID: fila.ProveedorMensajeID,
			FechaEnvio:         fila.FechaEnvio,
			Nombre:             fila.Nombre,
			Empresa:            fila.Empresa,
			Correo:             fila.Correo,
		})
	}

	return ListadoEnvios{Datos: datos, Total: total}, nil
}

func (r *Repositorio) Resumen(ctx context.Context, correspondenciaID uuid.UUID) (Resumen, error) {
	filas, err := r.consultas.ResumenEnvios(ctx, correspondenciaID)
	if err != nil {
		return Resumen{}, fmt.Errorf("no se pudo obtener el resumen: %w", err)
	}

	resumen := Resumen{PorEstado: make(map[string]int64, len(filas))}
	for _, fila := range filas {
		resumen.PorEstado[fila.Estado] = fila.Total
		resumen.Total += fila.Total
	}

	return resumen, nil
}
