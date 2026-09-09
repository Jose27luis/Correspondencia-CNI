package users

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/auth"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/middleware"
)

type Handler struct {
	servicio  *Servicio
	limitador *middleware.LimitadorIntentos
}

func NuevoHandler(servicio *Servicio, limitador *middleware.LimitadorIntentos) *Handler {
	return &Handler{servicio: servicio, limitador: limitador}
}

func (h *Handler) RegistrarPublicas(r chi.Router) {
	r.Route("/auth", func(ra chi.Router) {
		ra.With(h.limitador.Middleware).Post("/acceso", h.acceder)
		ra.With(h.limitador.Middleware).Post("/registro", h.registrar)
	})
}

func (h *Handler) RegistrarProtegidas(r chi.Router) {
	r.Get("/auth/perfil", h.perfil)
}

func (h *Handler) acceder(w http.ResponseWriter, r *http.Request) {
	var entrada EntradaAcceso
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	sesion, err := h.servicio.Acceder(r.Context(), entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	h.limitador.Perdonar(r)

	httpx.JSON(w, http.StatusOK, sesion)
}

func (h *Handler) registrar(w http.ResponseWriter, r *http.Request) {
	var entrada EntradaRegistro
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	ctx := r.Context()
	if cabecera := r.Header.Get("Authorization"); strings.HasPrefix(cabecera, "Bearer ") {
		if identidad, err := h.servicio.emisor.Verificar(strings.TrimPrefix(cabecera, "Bearer ")); err == nil {
			ctx = auth.ConIdentidad(ctx, identidad)
		}
	}

	usuario, err := h.servicio.Registrar(ctx, entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, usuario)
}

func (h *Handler) perfil(w http.ResponseWriter, r *http.Request) {
	usuario, err := h.servicio.Perfil(r.Context())
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, usuario)
}

func responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrCredencialesInvalidas):
		httpx.Error(w, http.StatusUnauthorized, "correo o contraseña incorrectos")
	case errors.Is(err, ErrRegistroNoPermitido):
		httpx.Error(w, http.StatusForbidden, "solo un usuario autenticado puede registrar nuevas cuentas")
	case errors.Is(err, ErrCorreoEnUso):
		httpx.Error(w, http.StatusConflict, "ya existe un usuario con ese correo")
	case errors.Is(err, ErrNoEncontrado):
		httpx.Error(w, http.StatusNotFound, "usuario no encontrado")
	case errors.Is(err, auth.ErrSinToken):
		httpx.Error(w, http.StatusUnauthorized, "no se envió el token de acceso")
	default:
		var erroresValidacion validator.ValidationErrors
		if errors.As(err, &erroresValidacion) {
			httpx.ErrorConDetalle(w, http.StatusUnprocessableEntity, "datos inválidos", httpx.DetallarValidacion(erroresValidacion))
			return
		}
		slog.Error("error no controlado en usuarios", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
