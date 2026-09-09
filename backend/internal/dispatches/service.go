package dispatches

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/correspondence"
)

const (
	limitePorDefecto = 50
	limiteMaximo     = 200
)

type Servicio struct {
	repositorio     *Repositorio
	correspondencia *correspondence.Repositorio
	cliente         *asynq.Client
}

func NuevoServicio(repositorio *Repositorio, correspondenciaRepo *correspondence.Repositorio, cliente *asynq.Client) *Servicio {
	return &Servicio{
		repositorio:     repositorio,
		correspondencia: correspondenciaRepo,
		cliente:         cliente,
	}
}

func (s *Servicio) Encolar(ctx context.Context, correspondenciaID uuid.UUID) (ResultadoEncolado, error) {
	pieza, err := s.correspondencia.Obtener(ctx, correspondenciaID)
	if err != nil {
		return ResultadoEncolado{}, err
	}

	if pieza.Estado != correspondence.EstadoBorrador {
		return ResultadoEncolado{}, ErrYaProcesada
	}

	if pieza.ListaID == nil {
		return ResultadoEncolado{}, ErrSinLista
	}

	creados, err := s.repositorio.CrearEnviosDeLista(ctx, correspondenciaID, *pieza.ListaID)
	if err != nil {
		return ResultadoEncolado{}, err
	}

	if creados == 0 {
		return ResultadoEncolado{}, ErrListaVacia
	}

	if _, err := s.correspondencia.CambiarEstado(ctx, correspondenciaID, correspondence.EstadoEncolada); err != nil {
		return ResultadoEncolado{}, err
	}

	pendientes, err := s.repositorio.IDsPendientes(ctx, correspondenciaID)
	if err != nil {
		return ResultadoEncolado{}, err
	}

	var encolados int64
	for _, envioID := range pendientes {
		tarea, err := NuevaTareaEnvio(CargaEnvio{EnvioID: envioID, CorrespondenciaID: correspondenciaID})
		if err != nil {
			return ResultadoEncolado{}, err
		}

		if _, err := s.cliente.EnqueueContext(ctx, tarea); err != nil {
			slog.Error("no se pudo encolar el envío", "envio_id", envioID, "error", err)
			continue
		}

		encolados++
	}

	if encolados == 0 {
		if _, err := s.correspondencia.CambiarEstado(ctx, correspondenciaID, correspondence.EstadoFallida); err != nil {
			slog.Error("no se pudo revertir el estado de la correspondencia", "correspondencia_id", correspondenciaID, "error", err)
		}
		return ResultadoEncolado{}, fmt.Errorf("no se pudo encolar ningún envío")
	}

	return ResultadoEncolado{
		CorrespondenciaID: correspondenciaID,
		Encolados:         encolados,
		Estado:            correspondence.EstadoEncolada,
	}, nil
}

func (s *Servicio) Listar(ctx context.Context, correspondenciaID uuid.UUID, estado *string, limite int32, desfase int32) (ListadoEnvios, error) {
	if limite <= 0 {
		limite = limitePorDefecto
	}
	if limite > limiteMaximo {
		limite = limiteMaximo
	}
	if desfase < 0 {
		desfase = 0
	}

	if _, err := s.correspondencia.Obtener(ctx, correspondenciaID); err != nil {
		return ListadoEnvios{}, err
	}

	return s.repositorio.Listar(ctx, correspondenciaID, estado, limite, desfase)
}

func (s *Servicio) Resumen(ctx context.Context, correspondenciaID uuid.UUID) (Resumen, error) {
	if _, err := s.correspondencia.Obtener(ctx, correspondenciaID); err != nil {
		return Resumen{}, err
	}
	return s.repositorio.Resumen(ctx, correspondenciaID)
}
