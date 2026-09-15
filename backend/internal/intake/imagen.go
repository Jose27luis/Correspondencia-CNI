package intake

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/storage"
)

const (
	TamanoMaximoFormulario = 6 << 20
	tamanoMaximoImagen     = 5 << 20
	CampoTrampa            = "confirmacion_correo"
)

var (
	ErrImagenInvalida = errors.New("la imagen debe ser PNG o JPG de hasta 5 MB")
	ErrFormulario     = errors.New("el formulario no es válido")
)

var tiposImagen = map[string]struct{}{
	"image/png":  {},
	"image/jpeg": {},
}

func EsRobot(r *http.Request) bool {
	return strings.TrimSpace(r.FormValue(CampoTrampa)) != ""
}

func Texto(r *http.Request, campo string) string {
	return strings.TrimSpace(r.FormValue(campo))
}

func TextoOpcional(r *http.Request, campo string) *string {
	valor := Texto(r, campo)
	if valor == "" {
		return nil
	}
	return &valor
}

func GuardarImagen(almacen *storage.Almacen, r *http.Request, campo string) (*string, error) {
	archivo, cabecera, err := r.FormFile(campo)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return nil, nil
		}
		return nil, ErrFormulario
	}
	defer archivo.Close()

	if cabecera.Size <= 0 || cabecera.Size > tamanoMaximoImagen {
		return nil, ErrImagenInvalida
	}

	inicio := make([]byte, 512)
	leidos, err := io.ReadFull(archivo, inicio)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, ErrImagenInvalida
	}
	if _, permitido := tiposImagen[http.DetectContentType(inicio[:leidos])]; !permitido {
		return nil, ErrImagenInvalida
	}
	if _, err := archivo.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("no se pudo leer la imagen: %w", err)
	}

	guardado, err := almacen.Guardar(archivo, cabecera)
	if err != nil {
		if errors.Is(err, storage.ErrTipoNoPermitido) || errors.Is(err, storage.ErrArchivoVacio) {
			return nil, ErrImagenInvalida
		}
		return nil, err
	}

	return &guardado.UrlArchivo, nil
}
