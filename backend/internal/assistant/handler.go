package assistant

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/ai"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
)

const (
	longitudMinimaInstruccion = 10
	longitudMaximaInstruccion = 2000
	longitudMaximaCarta       = 100000
	longitudMaximaTextoPegado = 40000
)

type Handler struct {
	asistente *ai.Asistente
}

func NuevoHandler(asistente *ai.Asistente) *Handler {
	return &Handler{asistente: asistente}
}

type EstadoAsistente struct {
	Disponible bool `json:"disponible"`
}

type PeticionRedaccion struct {
	Instruccion string   `json:"instruccion"`
	Asunto      string   `json:"asunto"`
	Cuerpo      string   `json:"cuerpo"`
	Variables   []string `json:"variables"`
}

type PeticionRevision struct {
	Asunto    string   `json:"asunto"`
	Cuerpo    string   `json:"cuerpo"`
	Variables []string `json:"variables"`
}

func (h *Handler) Registrar(r chi.Router) {
	r.Route("/asistente", func(ra chi.Router) {
		ra.Get("/estado", h.estado)
		ra.Post("/redactar", h.redactar)
		ra.Post("/revisar", h.revisar)
		ra.Post("/contactos/extraer", h.extraerContactos)
		ra.Post("/contactos/buscar", h.interpretarBusqueda)
		ra.Post("/sugerir-cuerpo", h.sugerirCuerpo)
	})
}

func (h *Handler) estado(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusOK, EstadoAsistente{Disponible: h.asistente.Disponible()})
}

func (h *Handler) redactar(w http.ResponseWriter, r *http.Request) {
	var peticion PeticionRedaccion
	if err := httpx.Decodificar(w, r, &peticion); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	peticion.Instruccion = strings.TrimSpace(peticion.Instruccion)

	if len(peticion.Instruccion) < longitudMinimaInstruccion {
		httpx.Error(w, http.StatusUnprocessableEntity, "describa con más detalle lo que quiere comunicar")
		return
	}

	if len(peticion.Instruccion) > longitudMaximaInstruccion || len(peticion.Cuerpo) > longitudMaximaCarta {
		httpx.Error(w, http.StatusRequestEntityTooLarge, "el texto enviado es demasiado largo")
		return
	}

	borrador, err := h.asistente.Redactar(r.Context(), ai.EntradaRedaccion{
		Instruccion: peticion.Instruccion,
		Asunto:      peticion.Asunto,
		Cuerpo:      peticion.Cuerpo,
		Variables:   peticion.Variables,
	})
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, borrador)
}

func (h *Handler) revisar(w http.ResponseWriter, r *http.Request) {
	var peticion PeticionRevision
	if err := httpx.Decodificar(w, r, &peticion); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	if strings.TrimSpace(peticion.Cuerpo) == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, "no hay texto que revisar")
		return
	}

	if len(peticion.Cuerpo) > longitudMaximaCarta {
		httpx.Error(w, http.StatusRequestEntityTooLarge, "la carta es demasiado larga para revisarla")
		return
	}

	revision, err := h.asistente.Revisar(r.Context(), peticion.Asunto, peticion.Cuerpo, peticion.Variables)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, revision)
}

func (h *Handler) sugerirCuerpo(w http.ResponseWriter, r *http.Request) {
	var peticion ai.EntradaSugerencia
	if err := httpx.Decodificar(w, r, &peticion); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	if len(strings.TrimSpace(peticion.Documento)) < longitudMinimaInstruccion {
		httpx.Error(w, http.StatusUnprocessableEntity, "suba primero el documento de Word para analizarlo")
		return
	}

	if len(peticion.Documento) > longitudMaximaCarta {
		httpx.Error(w, http.StatusRequestEntityTooLarge, "el documento es demasiado largo para analizarlo")
		return
	}

	sugerencia, err := h.asistente.SugerirCuerpo(r.Context(), peticion)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, sugerencia)
}

type PeticionExtraccion struct {
	Texto string `json:"texto"`
}

type PeticionBusqueda struct {
	Consulta string `json:"consulta"`
}

func (h *Handler) extraerContactos(w http.ResponseWriter, r *http.Request) {
	var peticion PeticionExtraccion
	if err := httpx.Decodificar(w, r, &peticion); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	peticion.Texto = strings.TrimSpace(peticion.Texto)

	if peticion.Texto == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, "pegue el texto del que quiere extraer contactos")
		return
	}

	if len(peticion.Texto) > longitudMaximaTextoPegado {
		httpx.Error(w, http.StatusRequestEntityTooLarge, "el texto pegado es demasiado largo")
		return
	}

	extraccion, err := h.asistente.ExtraerContactos(r.Context(), peticion.Texto)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, extraccion)
}

func (h *Handler) interpretarBusqueda(w http.ResponseWriter, r *http.Request) {
	var peticion PeticionBusqueda
	if err := httpx.Decodificar(w, r, &peticion); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	peticion.Consulta = strings.TrimSpace(peticion.Consulta)

	if peticion.Consulta == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, "escriba qué contactos está buscando")
		return
	}

	if len(peticion.Consulta) > longitudMaximaInstruccion {
		httpx.Error(w, http.StatusRequestEntityTooLarge, "la búsqueda es demasiado larga")
		return
	}

	criterios, err := h.asistente.InterpretarBusqueda(r.Context(), peticion.Consulta)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, criterios)
}

func responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ai.ErrNoConfigurado):
		httpx.Error(
			w,
			http.StatusServiceUnavailable,
			"el asistente de redacción no está configurado en este servidor",
		)
	case errors.Is(err, ai.ErrSinRespuesta):
		httpx.Error(w, http.StatusBadGateway, "el asistente no pudo completar la solicitud, intente de nuevo")
	default:
		slog.Error("error no controlado en el asistente", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
