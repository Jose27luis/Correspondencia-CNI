package users

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/auth"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
)

type Handler struct {
	servicio *Servicio
}

func NuevoHandler(servicio *Servicio) *Handler {
	return &Handler{servicio: servicio}
}

func (h *Handler) RegistrarPublicas(r chi.Router) {
	r.Route("/auth", func(ra chi.Router) {
		ra.Post("/acceso", h.acceder)
		ra.Post("/registro", h.registrar)
	})
}

func (h *Handler) RegistrarProtegidas(r chi.Router) {
	r.Get("/auth/perfil", h.perfil)
}

func (h *Handler) acceder(w http.ResponseWriter, r *http.Request) {
	var entrada EntradaAcceso
	if err := decodificar(r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	sesion, err := h.servicio.Acceder(r.Context(), entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, sesion)
}

func (h *Handler) registrar(w http.ResponseWriter, r *http.Request) {
	var entrada EntradaRegistro
	if err := decodificar(r, &entrada); err != nil {
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

func decodificar[T any](r *http.Request, destino *T) error {
	decodificador := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	decodificador.DisallowUnknownFields()
	return decodificador.Decode(destino)
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
			httpx.ErrorConDetalle(w, http.StatusUnprocessableEntity, "datos inválidos", detallarValidacion(erroresValidacion))
			return
		}
		slog.Error("error no controlado en usuarios", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "error interno del servidor")
	}
}

func detallarValidacion(erroresValidacion validator.ValidationErrors) map[string]string {
	detalle := make(map[string]string, len(erroresValidacion))
	for _, errorCampo := range erroresValidacion {
		detalle[strings.ToLower(errorCampo.Field())] = errorCampo.Tag()
	}
	return detalle
}
