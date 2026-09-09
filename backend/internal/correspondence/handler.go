package correspondence

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/auth"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
)

type Handler struct {
	servicio *Servicio
}

func NuevoHandler(servicio *Servicio) *Handler {
	return &Handler{servicio: servicio}
}

func (h *Handler) Registrar(r chi.Router) {
	r.Route("/correspondencia", func(rc chi.Router) {
		rc.Get("/", h.listar)
		rc.Post("/", h.crear)
		rc.Get("/{id}", h.obtener)
		rc.Put("/{id}", h.actualizar)
		rc.Delete("/{id}", h.eliminar)
		rc.Get("/{id}/previsualizacion", h.previsualizar)
		rc.Post("/{id}/adjuntos", h.agregarAdjunto)
		rc.Delete("/{id}/adjuntos/{adjuntoId}", h.eliminarAdjunto)
	})
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	listado, err := h.servicio.Listar(
		r.Context(),
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

func (h *Handler) crear(w http.ResponseWriter, r *http.Request) {
	var entrada EntradaCorrespondencia
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	correspondencia, err := h.servicio.Crear(r.Context(), entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, correspondencia)
}

func (h *Handler) obtener(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	correspondencia, err := h.servicio.Obtener(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, correspondencia)
}

func (h *Handler) actualizar(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	var entrada EntradaCorrespondencia
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	correspondencia, err := h.servicio.Actualizar(r.Context(), id, entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, correspondencia)
}

func (h *Handler) eliminar(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	if err := h.servicio.Eliminar(r.Context(), id); err != nil {
		responderError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) previsualizar(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	previsualizacion, err := h.servicio.Previsualizar(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, previsualizacion)
}

func (h *Handler) agregarAdjunto(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	var entrada EntradaAdjunto
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	adjunto, err := h.servicio.AgregarAdjunto(r.Context(), id, entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, adjunto)
}

func (h *Handler) eliminarAdjunto(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	adjuntoID, err := uuid.Parse(chi.URLParam(r, "adjuntoId"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador de adjunto inválido")
		return
	}

	if err := h.servicio.EliminarAdjunto(r.Context(), id, adjuntoID); err != nil {
		responderError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNoEncontrada):
		httpx.Error(w, http.StatusNotFound, "correspondencia no encontrada")
	case errors.Is(err, ErrAdjuntoInvalido):
		httpx.Error(w, http.StatusNotFound, "adjunto no encontrado")
	case errors.Is(err, ErrNoEsBorrador):
		httpx.Error(w, http.StatusConflict, "solo se puede modificar una correspondencia en borrador")
	case errors.Is(err, ErrListaInvalida):
		httpx.Error(w, http.StatusUnprocessableEntity, "la lista indicada no existe")
	case errors.Is(err, ErrSinDestinatario):
		httpx.Error(w, http.StatusUnprocessableEntity, "la correspondencia necesita una lista con contactos para previsualizar")
	case errors.Is(err, ErrLimiteAdjuntos):
		httpx.Error(w, http.StatusRequestEntityTooLarge, "los adjuntos superan los 25 MB permitidos")
	case errors.Is(err, auth.ErrSinToken):
		httpx.Error(w, http.StatusUnauthorized, "no se envió el token de acceso")
	default:
		var erroresValidacion validator.ValidationErrors
		if errors.As(err, &erroresValidacion) {
			httpx.ErrorConDetalle(w, http.StatusUnprocessableEntity, "datos inválidos", httpx.DetallarValidacion(erroresValidacion))
			return
		}
		slog.Error("error no controlado en correspondencia", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
