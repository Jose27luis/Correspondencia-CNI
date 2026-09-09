package dispatches

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/correspondence"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
)

type Handler struct {
	servicio *Servicio
}

func NuevoHandler(servicio *Servicio) *Handler {
	return &Handler{servicio: servicio}
}

func (h *Handler) Registrar(r chi.Router) {
	r.Post("/correspondencia/{id}/enviar", h.enviar)
	r.Get("/correspondencia/{id}/envios", h.listar)
	r.Get("/correspondencia/{id}/envios/resumen", h.resumen)
}

func (h *Handler) enviar(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	resultado, err := h.servicio.Encolar(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusAccepted, resultado)
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	listado, err := h.servicio.Listar(
		r.Context(),
		id,
		httpx.LeerTexto(r, "estado"),
		int32(httpx.LeerEntero(r, "limite", limitePorDefecto)),
		int32(httpx.LeerEntero(r, "desfase", 0)),
	)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, listado)
}

func (h *Handler) resumen(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	resumen, err := h.servicio.Resumen(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, resumen)
}

func responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, correspondence.ErrNoEncontrada), errors.Is(err, ErrNoEncontrado):
		httpx.Error(w, http.StatusNotFound, "correspondencia no encontrada")
	case errors.Is(err, ErrYaProcesada):
		httpx.Error(w, http.StatusConflict, "la correspondencia ya fue encolada o enviada")
	case errors.Is(err, ErrSinLista):
		httpx.Error(w, http.StatusUnprocessableEntity, "la correspondencia no tiene una lista asignada")
	case errors.Is(err, ErrListaVacia):
		httpx.Error(w, http.StatusUnprocessableEntity, "la lista no tiene contactos")
	default:
		slog.Error("error no controlado en envíos", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
