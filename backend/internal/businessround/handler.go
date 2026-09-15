package businessround

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/intake"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/ai"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/middleware"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/storage"
)

type Handler struct {
	servicio  *Servicio
	almacen   *storage.Almacen
	limitador *middleware.LimitadorIntentos
}

func NuevoHandler(servicio *Servicio, almacen *storage.Almacen, limitador *middleware.LimitadorIntentos) *Handler {
	return &Handler{servicio: servicio, almacen: almacen, limitador: limitador}
}

func (h *Handler) RegistrarPublicas(r chi.Router) {
	r.Get("/publico/rueda", h.listarPublicas)
	r.With(h.limitador.Middleware).Post("/publico/rueda", h.registrar)
}

func (h *Handler) RegistrarProtegidas(r chi.Router) {
	r.Route("/rueda", func(rr chi.Router) {
		rr.Get("/empresas", h.listarEmpresas)
		rr.Post("/empresas/{id}/aprobar", h.aprobarEmpresa)
		rr.Post("/empresas/{id}/rechazar", h.rechazarEmpresa)
		rr.Get("/publicaciones", h.listarPublicaciones)
		rr.Post("/publicaciones/{id}/aprobar", h.aprobarPublicacion)
		rr.Post("/publicaciones/{id}/rechazar", h.rechazarPublicacion)
		rr.Post("/publicaciones/{id}/coincidencias", h.buscarCoincidencias)
		rr.Get("/coincidencias", h.listarCoincidencias)
		rr.Post("/coincidencias/{id}/notificar", h.notificar)
		rr.Post("/coincidencias/{id}/descartar", h.descartar)
	})
}

func (h *Handler) listarPublicas(w http.ResponseWriter, r *http.Request) {
	datos, err := h.servicio.ListarPublicas(r.Context(), httpx.LeerTexto(r, "tipo"))
	if err != nil {
		responderError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, datos)
}

func (h *Handler) registrar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, intake.TamanoMaximoFormulario)
	if err := r.ParseMultipartForm(intake.TamanoMaximoFormulario); err != nil {
		httpx.Error(w, http.StatusBadRequest, intake.ErrFormulario.Error())
		return
	}

	if intake.EsRobot(r) {
		httpx.JSON(w, http.StatusCreated, Registro{Recibido: true})
		return
	}

	entrada := EntradaRegistro{
		RazonSocial:      intake.Texto(r, "razon_social"),
		Ruc:              intake.Texto(r, "ruc"),
		PersonaEncargada: intake.Texto(r, "persona_encargada"),
		CargoEncargado:   intake.TextoOpcional(r, "cargo_encargado"),
		Correo:           intake.Texto(r, "correo"),
		Direccion:        intake.TextoOpcional(r, "direccion"),
		Ciudad:           intake.TextoOpcional(r, "ciudad"),
		Region:           intake.TextoOpcional(r, "region"),
		Pais:             intake.Texto(r, "pais"),
		CodigoPostal:     intake.TextoOpcional(r, "codigo_postal"),
		Telefono:         intake.TextoOpcional(r, "telefono"),
		Celular:          intake.TextoOpcional(r, "celular"),
		PaginaWeb:        intake.TextoOpcional(r, "pagina_web"),
		Tipo:             intake.Texto(r, "tipo"),
		Titulo:           intake.Texto(r, "titulo"),
		Descripcion:      intake.Texto(r, "descripcion"),
	}
	if err := h.servicio.Validar(&entrada); err != nil {
		responderError(w, err)
		return
	}

	imagen, err := intake.GuardarImagen(h.almacen, r, "imagen")
	if err != nil {
		responderError(w, err)
		return
	}
	entrada.ImagenUrl = imagen

	if err := h.servicio.Registrar(r.Context(), entrada); err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, Registro{Recibido: true})
}

func (h *Handler) listarEmpresas(w http.ResponseWriter, r *http.Request) {
	listado, err := h.servicio.ListarEmpresas(
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

func (h *Handler) aprobarEmpresa(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	empresa, err := h.servicio.AprobarEmpresa(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, empresa)
}

func (h *Handler) rechazarEmpresa(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	entrada, ok := leerRechazo(w, r)
	if !ok {
		return
	}

	empresa, err := h.servicio.RechazarEmpresa(r.Context(), id, entrada)
	if err != nil {
		responderError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, empresa)
}

func (h *Handler) listarPublicaciones(w http.ResponseWriter, r *http.Request) {
	listado, err := h.servicio.ListarPublicaciones(
		r.Context(),
		httpx.LeerTexto(r, "estado"),
		httpx.LeerTexto(r, "tipo"),
		int32(httpx.LeerEntero(r, "limite", limitePorDefecto)),
		int32(httpx.LeerEntero(r, "desfase", 0)),
	)
	if err != nil {
		responderError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, listado)
}

func (h *Handler) aprobarPublicacion(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	publicacion, err := h.servicio.AprobarPublicacion(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}

	h.servicio.BuscarEnSegundoPlano(id)
	httpx.JSON(w, http.StatusOK, publicacion)
}

func (h *Handler) rechazarPublicacion(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	entrada, ok := leerRechazo(w, r)
	if !ok {
		return
	}

	publicacion, err := h.servicio.RechazarPublicacion(r.Context(), id, entrada)
	if err != nil {
		responderError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, publicacion)
}

func (h *Handler) buscarCoincidencias(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	resultado, err := h.servicio.BuscarCoincidencias(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, resultado)
}

func (h *Handler) listarCoincidencias(w http.ResponseWriter, r *http.Request) {
	datos, err := h.servicio.ListarCoincidencias(r.Context(), httpx.LeerTexto(r, "estado"))
	if err != nil {
		responderError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, datos)
}

func (h *Handler) notificar(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	if err := h.servicio.Notificar(r.Context(), id); err != nil {
		responderError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) descartar(w http.ResponseWriter, r *http.Request) {
	id, ok := leerID(w, r)
	if !ok {
		return
	}

	if err := h.servicio.Descartar(r.Context(), id); err != nil {
		responderError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func leerID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return uuid.Nil, false
	}
	return id, true
}

func leerRechazo(w http.ResponseWriter, r *http.Request) (EntradaRechazo, bool) {
	var entrada EntradaRechazo
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return EntradaRechazo{}, false
	}
	return entrada, true
}

func responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNoEncontrada):
		httpx.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrYaRevisada), errors.Is(err, ErrYaNotificada),
		errors.Is(err, ErrEmpresaPendiente), errors.Is(err, ErrNoAprobada):
		httpx.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrTipoInvalido), errors.Is(err, intake.ErrImagenInvalida), errors.Is(err, intake.ErrFormulario):
		httpx.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ai.ErrNoConfigurado):
		httpx.Error(w, http.StatusServiceUnavailable, "la búsqueda con IA no está configurada")
	default:
		var erroresValidacion validator.ValidationErrors
		if errors.As(err, &erroresValidacion) {
			httpx.ErrorConDetalle(w, http.StatusUnprocessableEntity, "datos inválidos", httpx.DetallarValidacion(erroresValidacion))
			return
		}
		slog.Error("error no controlado en la rueda de negocios", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
