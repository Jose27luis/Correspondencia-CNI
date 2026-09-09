package contacts

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
)

const tamanoMaximoCSV = 10 << 20

type Handler struct {
	servicio *Servicio
}

func NuevoHandler(servicio *Servicio) *Handler {
	return &Handler{servicio: servicio}
}

func (h *Handler) Registrar(r chi.Router) {
	r.Route("/contactos", func(rc chi.Router) {
		rc.Get("/", h.listar)
		rc.Post("/", h.crear)
		rc.Post("/importar", h.importar)
		rc.Get("/{id}", h.obtener)
		rc.Put("/{id}", h.actualizar)
		rc.Delete("/{id}", h.eliminar)
	})
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	filtro := FiltroListado{
		Limite:  int32(leerEntero(r, "limite", limitePorDefecto)),
		Desfase: int32(leerEntero(r, "desfase", 0)),
	}

	if busqueda := strings.TrimSpace(r.URL.Query().Get("busqueda")); busqueda != "" {
		filtro.Busqueda = &busqueda
	}

	listado, err := h.servicio.Listar(r.Context(), filtro)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, listado)
}

func (h *Handler) crear(w http.ResponseWriter, r *http.Request) {
	var entrada EntradaContacto
	if err := decodificar(r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	contacto, err := h.servicio.Crear(r.Context(), entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, contacto)
}

func (h *Handler) obtener(w http.ResponseWriter, r *http.Request) {
	id, err := leerID(r)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	contacto, err := h.servicio.Obtener(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, contacto)
}

func (h *Handler) actualizar(w http.ResponseWriter, r *http.Request) {
	id, err := leerID(r)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	var entrada EntradaContacto
	if err := decodificar(r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	contacto, err := h.servicio.Actualizar(r.Context(), id, entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, contacto)
}

func (h *Handler) eliminar(w http.ResponseWriter, r *http.Request) {
	id, err := leerID(r)
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

func (h *Handler) importar(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(tamanoMaximoCSV); err != nil {
		httpx.Error(w, http.StatusBadRequest, "no se pudo leer el archivo enviado")
		return
	}

	archivo, cabecera, err := r.FormFile("archivo")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "falta el archivo en el campo archivo")
		return
	}
	defer archivo.Close()

	if cabecera.Size > tamanoMaximoCSV {
		httpx.Error(w, http.StatusRequestEntityTooLarge, "el archivo supera los 10 MB")
		return
	}

	resumen, err := h.servicio.ImportarCSV(r.Context(), archivo)
	if err != nil {
		if errors.Is(err, ErrCabeceraCSVInvalida) {
			httpx.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, resumen)
}

func decodificar[T any](r *http.Request, destino *T) error {
	decodificador := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	decodificador.DisallowUnknownFields()
	return decodificador.Decode(destino)
}

func leerID(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "id"))
}

func leerEntero(r *http.Request, clave string, porDefecto int) int {
	valor := r.URL.Query().Get(clave)
	if valor == "" {
		return porDefecto
	}
	numero, err := strconv.Atoi(valor)
	if err != nil {
		return porDefecto
	}
	return numero
}

func responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNoEncontrado):
		httpx.Error(w, http.StatusNotFound, "contacto no encontrado")
	case errors.Is(err, ErrCorreoEnUso):
		httpx.Error(w, http.StatusConflict, "ya existe un contacto con ese correo")
	default:
		var erroresValidacion validator.ValidationErrors
		if errors.As(err, &erroresValidacion) {
			httpx.ErrorConDetalle(w, http.StatusUnprocessableEntity, "datos inválidos", detallarValidacion(erroresValidacion))
			return
		}
		slog.Error("error no controlado en contactos", "error", err)
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
