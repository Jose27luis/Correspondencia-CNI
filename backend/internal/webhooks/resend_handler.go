package webhooks

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/dispatches"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/suppression"
)

const tamanoMaximoCuerpo = 1 << 20

var estadosPorEvento = map[string]string{
	"email.sent":             dispatches.EstadoEnviado,
	"email.delivered":        dispatches.EstadoEntregado,
	"email.delivery_delayed": dispatches.EstadoEnviado,
	"email.bounced":          dispatches.EstadoRebotado,
	"email.complained":       dispatches.EstadoRebotado,
	"email.opened":           "",
	"email.clicked":          "",
}

type eventoResend struct {
	Tipo      string          `json:"type"`
	CreadoEn  time.Time       `json:"created_at"`
	Datos     datosEvento     `json:"data"`
	Contenido json.RawMessage `json:"-"`
}

type datosEvento struct {
	EmailID string `json:"email_id"`
}

type Handler struct {
	verificador *Verificador
	repositorio *dispatches.Repositorio
	exclusiones *suppression.Repositorio
}

func NuevoHandler(
	verificador *Verificador,
	repositorio *dispatches.Repositorio,
	exclusiones *suppression.Repositorio,
) *Handler {
	return &Handler{verificador: verificador, repositorio: repositorio, exclusiones: exclusiones}
}

func (h *Handler) Registrar(r chi.Router) {
	r.Post("/webhooks/resend", h.recibir)
}

func (h *Handler) recibir(w http.ResponseWriter, r *http.Request) {
	cuerpo, err := io.ReadAll(http.MaxBytesReader(w, r.Body, tamanoMaximoCuerpo))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "no se pudo leer el cuerpo del webhook")
		return
	}

	if err := h.verificador.Verificar(r.Header, cuerpo); err != nil {
		slog.Warn("webhook rechazado", "error", err)
		httpx.Error(w, http.StatusUnauthorized, "firma del webhook inválida")
		return
	}

	var evento eventoResend
	if err := json.Unmarshal(cuerpo, &evento); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el webhook no tiene un cuerpo JSON válido")
		return
	}

	if evento.Datos.EmailID == "" {
		httpx.JSON(w, http.StatusOK, map[string]string{"estado": "ignorado"})
		return
	}

	estado, conocido := estadosPorEvento[evento.Tipo]
	if !conocido {
		slog.Info("evento de Resend no contemplado", "tipo", evento.Tipo)
		httpx.JSON(w, http.StatusOK, map[string]string{"estado": "ignorado"})
		return
	}

	if err := h.registrar(r, evento, estado, cuerpo); err != nil {
		if errors.Is(err, dispatches.ErrNoEncontrado) {
			slog.Warn("el webhook refiere a un envío desconocido", "mensaje_id", evento.Datos.EmailID)
			httpx.JSON(w, http.StatusOK, map[string]string{"estado": "ignorado"})
			return
		}

		slog.Error("no se pudo procesar el webhook", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "no se pudo registrar el evento")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"estado": "registrado"})
}

func (h *Handler) registrar(r *http.Request, evento eventoResend, estado string, cuerpo []byte) error {
	ctx := r.Context()

	var envioID uuid.UUID
	var err error

	if estado == "" {
		envioID, err = h.repositorio.BuscarPorMensaje(ctx, evento.Datos.EmailID)
	} else {
		envioID, err = h.repositorio.ActualizarPorMensaje(ctx, evento.Datos.EmailID, estado)
	}

	if err != nil {
		return err
	}

	if estado == dispatches.EstadoRebotado {
		if err := h.exclusiones.SuprimirPorEnvio(
			ctx,
			envioID,
			suppression.MotivoRebote,
			evento.Tipo,
		); err != nil {
			slog.Error("no se pudo excluir el correo rebotado", "envio_id", envioID, "error", err)
		} else {
			slog.Info("correo excluido por rebote", "envio_id", envioID, "tipo", evento.Tipo)
		}
	}

	fecha := evento.CreadoEn
	if fecha.IsZero() {
		fecha = time.Now()
	}

	return h.repositorio.RegistrarEvento(ctx, envioID, evento.Tipo, fecha, cuerpo)
}
