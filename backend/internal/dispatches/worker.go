package dispatches

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"golang.org/x/time/rate"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/correspondence"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/mailer"
)

type Worker struct {
	repositorio     *Repositorio
	correspondencia *correspondence.Repositorio
	proveedor       mailer.Proveedor
	limitador       *rate.Limiter
}

func NuevoWorker(repositorio *Repositorio, correspondencia *correspondence.Repositorio, proveedor mailer.Proveedor, porSegundo int) *Worker {
	if porSegundo <= 0 {
		porSegundo = 1
	}

	return &Worker{
		repositorio:     repositorio,
		correspondencia: correspondencia,
		proveedor:       proveedor,
		limitador:       rate.NewLimiter(rate.Limit(porSegundo), porSegundo),
	}
}

func (w *Worker) Registrar(mux *asynq.ServeMux) {
	mux.HandleFunc(TipoEnviarCorreo, w.procesarEnvio)
}

func (w *Worker) procesarEnvio(ctx context.Context, tarea *asynq.Task) error {
	carga, err := LeerCargaEnvio(tarea)
	if err != nil {
		return err
	}

	correspondenciaID, contacto, estado, err := w.repositorio.DatosDeEnvio(ctx, carga.EnvioID)
	if err != nil {
		if errors.Is(err, ErrNoEncontrado) {
			slog.Warn("el envío ya no existe, se descarta la tarea", "envio_id", carga.EnvioID)
			return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
		}
		return err
	}

	if estado != EstadoPendiente {
		slog.Info("el envío ya fue procesado, se omite", "envio_id", carga.EnvioID, "estado", estado)
		return nil
	}

	pieza, err := w.correspondencia.Obtener(ctx, correspondenciaID)
	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	asunto := correspondence.Renderizar(pieza.Asunto, contacto)
	cuerpo := correspondence.Renderizar(pieza.Cuerpo, contacto)

	adjuntos := make([]mailer.Adjunto, 0, len(pieza.Adjuntos))
	for _, adjunto := range pieza.Adjuntos {
		adjuntos = append(adjuntos, mailer.Adjunto{
			NombreArchivo: adjunto.NombreArchivo,
			UrlArchivo:    adjunto.UrlArchivo,
		})
	}

	if err := w.limitador.Wait(ctx); err != nil {
		return fmt.Errorf("se canceló la espera del límite de envío: %w", err)
	}

	resultado, err := w.proveedor.Enviar(ctx, mailer.Mensaje{
		Para:     contacto.Correo,
		Asunto:   asunto.Texto,
		Cuerpo:   cuerpo.Texto,
		Adjuntos: adjuntos,
	})
	if err != nil {
		slog.Error("no se pudo enviar el correo", "envio_id", carga.EnvioID, "correo", contacto.Correo, "error", err)

		if tarea.ResultWriter() != nil && esUltimoIntento(ctx) {
			if err := w.repositorio.MarcarFallido(ctx, carga.EnvioID); err != nil {
				slog.Error("no se pudo marcar el envío como fallido", "envio_id", carga.EnvioID, "error", err)
			}
		}

		return err
	}

	if err := w.repositorio.MarcarEnviado(ctx, carga.EnvioID, resultado.MensajeID); err != nil {
		return err
	}

	slog.Info("correo entregado al proveedor",
		"envio_id", carga.EnvioID,
		"correo", contacto.Correo,
		"mensaje_id", resultado.MensajeID,
		"proveedor", w.proveedor.Nombre(),
	)

	return nil
}

func esUltimoIntento(ctx context.Context) bool {
	intento, okIntento := asynq.GetRetryCount(ctx)
	maximo, okMaximo := asynq.GetMaxRetry(ctx)

	if !okIntento || !okMaximo {
		return false
	}

	return intento >= maximo
}
