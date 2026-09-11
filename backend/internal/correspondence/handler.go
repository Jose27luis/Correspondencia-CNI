package correspondence

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/auth"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/pdf"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/storage"
)

const tamanoMaximoSubida = 25 << 20

type Handler struct {
	servicio    *Servicio
	convertidor *pdf.Convertidor
}

func NuevoHandler(servicio *Servicio, convertidor *pdf.Convertidor) *Handler {
	return &Handler{servicio: servicio, convertidor: convertidor}
}

func (h *Handler) Registrar(r chi.Router) {
	r.Route("/correspondencia", func(rc chi.Router) {
		rc.Get("/", h.listar)
		rc.Post("/", h.crear)
		rc.Get("/{id}", h.obtener)
		rc.Put("/{id}", h.actualizar)
		rc.Delete("/{id}", h.eliminar)
		rc.Post("/leer-documento", h.leerDocumento)
		rc.Get("/{id}/previsualizacion", h.previsualizar)
		rc.Post("/{id}/prueba", h.enviarPrueba)
		rc.Get("/{id}/plantilla/vista-previa", h.vistaPreviaPlantilla)
		rc.Post("/{id}/plantilla", h.subirPlantilla)
		rc.Delete("/{id}/plantilla", h.quitarPlantilla)
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

func (h *Handler) leerDocumento(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(tamanoMaximoSubida); err != nil {
		httpx.Error(w, http.StatusBadRequest, "no se pudo leer el archivo enviado")
		return
	}

	archivo, cabecera, err := r.FormFile("archivo")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "falta el archivo en el campo archivo")
		return
	}
	defer archivo.Close()

	if !strings.HasSuffix(strings.ToLower(cabecera.Filename), ".docx") {
		httpx.Error(w, http.StatusUnsupportedMediaType, "el documento debe ser un archivo .docx")
		return
	}

	contenido, err := LeerWord(archivo, cabecera.Size)
	if err != nil {
		if errors.Is(err, ErrWordInvalido) || errors.Is(err, ErrWordVacio) {
			httpx.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, contenido)
}

func (h *Handler) enviarPrueba(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	resultado, err := h.servicio.EnviarPrueba(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, resultado)
}

func (h *Handler) vistaPreviaPlantilla(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	var contactoID *uuid.UUID
	if valor := r.URL.Query().Get("contacto"); valor != "" {
		interpretado, err := uuid.Parse(valor)
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, "identificador de contacto inválido")
			return
		}
		contactoID = &interpretado
	}

	vista, err := h.servicio.GenerarVistaPrevia(r.Context(), id, contactoID)
	if err != nil {
		responderError(w, err)
		return
	}

	if r.URL.Query().Get("formato") == "pdf" {
		documento, err := h.convertidor.ConvertirDOCX(r.Context(), vista.Documento)
		if err != nil {
			if errors.Is(err, pdf.ErrNoDisponible) {
				httpx.Error(w, http.StatusServiceUnavailable, "la vista en PDF no está disponible en este servidor")
				return
			}
			slog.Error("no se pudo convertir la vista previa a PDF", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "no se pudo generar la vista en PDF")
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", `inline; filename="vista-previa.pdf"`)
		w.Header().Set("Cache-Control", "no-store")

		if _, err := w.Write(documento); err != nil {
			slog.Error("no se pudo enviar la vista previa en PDF", "error", err)
		}
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", vista.NombreArchivo))
	w.Header().Set("Cache-Control", "no-store")

	if _, err := w.Write(vista.Documento); err != nil {
		slog.Error("no se pudo enviar la vista previa", "error", err)
	}
}

func (h *Handler) subirPlantilla(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	if err := r.ParseMultipartForm(tamanoMaximoSubida); err != nil {
		httpx.Error(w, http.StatusBadRequest, "no se pudo leer el archivo enviado")
		return
	}

	archivo, cabecera, err := r.FormFile("archivo")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "falta el archivo en el campo archivo")
		return
	}
	defer archivo.Close()

	if !strings.HasSuffix(strings.ToLower(cabecera.Filename), ".docx") {
		httpx.Error(w, http.StatusUnsupportedMediaType, "la plantilla debe ser un archivo .docx")
		return
	}

	pieza, err := h.servicio.SubirPlantilla(r.Context(), id, archivo, cabecera)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, pieza)
}

func (h *Handler) quitarPlantilla(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	pieza, err := h.servicio.QuitarPlantilla(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, pieza)
}

func (h *Handler) agregarAdjunto(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	if err := r.ParseMultipartForm(tamanoMaximoSubida); err != nil {
		httpx.Error(w, http.StatusBadRequest, "no se pudo leer el archivo enviado")
		return
	}

	archivo, cabecera, err := r.FormFile("archivo")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "falta el archivo en el campo archivo")
		return
	}
	defer archivo.Close()

	if cabecera.Size > tamanoMaximoSubida {
		httpx.Error(w, http.StatusRequestEntityTooLarge, "el archivo supera los 25 MB")
		return
	}

	adjunto, err := h.servicio.SubirAdjunto(r.Context(), id, archivo, cabecera)
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
	case errors.Is(err, storage.ErrTipoNoPermitido):
		httpx.Error(w, http.StatusUnsupportedMediaType, "solo se aceptan archivos PDF, Word, Excel o imágenes")
	case errors.Is(err, storage.ErrArchivoVacio):
		httpx.Error(w, http.StatusUnprocessableEntity, "el archivo está vacío")
	case errors.Is(err, ErrSinPlantilla):
		httpx.Error(w, http.StatusUnprocessableEntity, "esta carta no tiene una carta en Word personalizada")
	case errors.Is(err, ErrWordInvalido), errors.Is(err, ErrWordVacio):
		httpx.Error(w, http.StatusUnprocessableEntity, err.Error())
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
