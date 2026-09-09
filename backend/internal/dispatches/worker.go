package dispatches

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"golang.org/x/time/rate"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/correspondence"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/mailer"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/storage"
)

type Worker struct {
	repositorio     *Repositorio
	correspondencia *correspondence.Repositorio
	proveedor       mailer.Proveedor
	almacen         *storage.Almacen
	limitador       *rate.Limiter
}

func NuevoWorker(
	repositorio *Repositorio,
	correspondencia *correspondence.Repositorio,
	proveedor mailer.Proveedor,
	almacen *storage.Almacen,
	porSegundo int,
) *Worker {
	if porSegundo <= 0 {
		porSegundo = 1
	}

	return &Worker{
		repositorio:     repositorio,
		correspondencia: correspondencia,
		proveedor:       proveedor,
		almacen:         almacen,
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

	adjuntos := make([]mailer.Adjunto, 0, len(pieza.Adjuntos)+1)
	for _, adjunto := range pieza.Adjuntos {
		adjuntos = append(adjuntos, mailer.Adjunto{
			NombreArchivo: adjunto.NombreArchivo,
			UrlArchivo:    adjunto.UrlArchivo,
		})
	}

	if pieza.PlantillaURL != nil {
		generado, err := w.generarDocumento(*pieza.PlantillaURL, pieza.PlantillaNombre, contacto)
		if err != nil {
			slog.Error("no se pudo generar el documento personalizado",
				"envio_id", carga.EnvioID,
				"error", err,
			)
			return err
		}

		adjuntos = append(adjuntos, generado)
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

		if esUltimoIntento(ctx) {
			if err := w.repositorio.MarcarFallido(ctx, carga.EnvioID); err != nil {
				slog.Error("no se pudo marcar el envío como fallido", "envio_id", carga.EnvioID, "error", err)
			}
			w.cerrarSiTermino(ctx, correspondenciaID)
		}

		return err
	}

	if err := w.repositorio.MarcarEnviado(ctx, carga.EnvioID, resultado.MensajeID); err != nil {
		return err
	}

	w.cerrarSiTermino(ctx, correspondenciaID)

	slog.Info("correo entregado al proveedor",
		"envio_id", carga.EnvioID,
		"correo", contacto.Correo,
		"mensaje_id", resultado.MensajeID,
		"proveedor", w.proveedor.Nombre(),
	)

	return nil
}

func (w *Worker) cerrarSiTermino(ctx context.Context, correspondenciaID uuid.UUID) {
	estado, err := w.repositorio.CerrarSiTermino(ctx, correspondenciaID)
	if err != nil {
		slog.Error("no se pudo cerrar la correspondencia",
			"correspondencia_id", correspondenciaID,
			"error", err,
		)
		return
	}

	if estado != "" {
		slog.Info("correspondencia completada",
			"correspondencia_id", correspondenciaID,
			"estado", estado,
		)
	}
}

func (w *Worker) generarDocumento(
	urlPlantilla string,
	nombrePlantilla *string,
	contacto contacts.Contacto,
) (mailer.Adjunto, error) {
	plantilla, err := w.almacen.Leer(urlPlantilla)
	if err != nil {
		return mailer.Adjunto{}, err
	}

	documento, err := correspondence.GenerarDocumento(plantilla, contacto)
	if err != nil {
		return mailer.Adjunto{}, err
	}

	return mailer.Adjunto{
		NombreArchivo: nombreParaContacto(nombrePlantilla, contacto),
		Contenido:     documento,
	}, nil
}

func nombreParaContacto(nombrePlantilla *string, contacto contacts.Contacto) string {
	base := "carta"
	if nombrePlantilla != nil {
		base = strings.TrimSuffix(*nombrePlantilla, filepath.Ext(*nombrePlantilla))
	}

	empresa := strings.Map(func(caracter rune) rune {
		switch {
		case caracter >= 'a' && caracter <= 'z', caracter >= 'A' && caracter <= 'Z':
			return caracter
		case caracter >= '0' && caracter <= '9':
			return caracter
		case caracter == ' ':
			return '-'
		default:
			return -1
		}
	}, contacto.Empresa)

	if empresa == "" {
		return fmt.Sprintf("%s.docx", base)
	}

	return fmt.Sprintf("%s-%s.docx", base, empresa)
}

func esUltimoIntento(ctx context.Context) bool {
	intento, okIntento := asynq.GetRetryCount(ctx)
	maximo, okMaximo := asynq.GetMaxRetry(ctx)

	if !okIntento || !okMaximo {
		return false
	}

	return intento >= maximo
}
