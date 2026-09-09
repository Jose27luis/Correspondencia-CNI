package suppression

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
)

type Handler struct {
	servicio *Servicio
}

func NuevoHandler(servicio *Servicio) *Handler {
	return &Handler{servicio: servicio}
}

func (h *Handler) RegistrarPublicas(r chi.Router) {
	r.Post("/baja", h.darDeBaja)
}

func (h *Handler) RegistrarProtegidas(r chi.Router) {
	r.Route("/exclusiones", func(re chi.Router) {
		re.Get("/", h.listar)
		re.Post("/", h.agregar)
		re.Delete("/{correo}", h.quitar)
	})
}

func (h *Handler) darDeBaja(w http.ResponseWriter, r *http.Request) {
	contactoID, err := uuid.Parse(r.URL.Query().Get("c"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "el enlace de baja no es válido")
		return
	}

	resultado, err := h.servicio.DarDeBaja(r.Context(), contactoID, r.URL.Query().Get("t"))
	if err != nil {
		if errors.Is(err, ErrTokenInvalido) {
			httpx.Error(w, http.StatusBadRequest, "el enlace de baja no es válido")
			return
		}
		slog.Error("no se pudo procesar la baja", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "no se pudo procesar la baja")
		return
	}

	httpx.JSON(w, http.StatusOK, resultado)
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	listado, err := h.servicio.Listar(
		r.Context(),
		int32(httpx.LeerEntero(r, "limite", limitePorDefecto)),
		int32(httpx.LeerEntero(r, "desfase", 0)),
	)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, listado)
}

func (h *Handler) agregar(w http.ResponseWriter, r *http.Request) {
	var entrada EntradaSupresion
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	if err := h.servicio.Agregar(r.Context(), entrada); err != nil {
		responderError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) quitar(w http.ResponseWriter, r *http.Request) {
	if err := h.servicio.Quitar(r.Context(), chi.URLParam(r, "correo")); err != nil {
		responderError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNoEncontrado):
		httpx.Error(w, http.StatusNotFound, "el correo no está en la lista de exclusión")
	default:
		var erroresValidacion validator.ValidationErrors
		if errors.As(err, &erroresValidacion) {
			httpx.ErrorConDetalle(w, http.StatusUnprocessableEntity, "datos inválidos", httpx.DetallarValidacion(erroresValidacion))
			return
		}
		slog.Error("error no controlado en exclusiones", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
