package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const permisosDirectorio = 0o755

var (
	ErrTipoNoPermitido = errors.New("el tipo de archivo no está permitido")
	ErrArchivoVacio    = errors.New("el archivo está vacío")
)

var extensionesPermitidas = map[string]string{
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
}

type ArchivoGuardado struct {
	NombreArchivo string
	UrlArchivo    string
	Tipo          string
	TamanoBytes   int64
}

type Almacen struct {
	directorio string
	urlPublica string
}

func NuevoAlmacen(directorio string, urlPublica string) (*Almacen, error) {
	if strings.TrimSpace(directorio) == "" {
		return nil, errors.New("falta la ruta del almacén de archivos")
	}

	if err := os.MkdirAll(directorio, permisosDirectorio); err != nil {
		return nil, fmt.Errorf("no se pudo preparar el almacén: %w", err)
	}

	return &Almacen{
		directorio: directorio,
		urlPublica: strings.TrimSuffix(urlPublica, "/"),
	}, nil
}

func (a *Almacen) Guardar(archivo multipart.File, cabecera *multipart.FileHeader) (ArchivoGuardado, error) {
	if cabecera.Size <= 0 {
		return ArchivoGuardado{}, ErrArchivoVacio
	}

	nombreOriginal := filepath.Base(cabecera.Filename)
	extension := strings.ToLower(filepath.Ext(nombreOriginal))

	tipo, permitido := extensionesPermitidas[extension]
	if !permitido {
		return ArchivoGuardado{}, ErrTipoNoPermitido
	}

	nombreEnDisco := uuid.NewString() + extension
	rutaDestino := filepath.Join(a.directorio, nombreEnDisco)

	destino, err := os.Create(rutaDestino)
	if err != nil {
		return ArchivoGuardado{}, fmt.Errorf("no se pudo crear el archivo: %w", err)
	}
	defer destino.Close()

	escritos, err := io.Copy(destino, archivo)
	if err != nil {
		os.Remove(rutaDestino)
		return ArchivoGuardado{}, fmt.Errorf("no se pudo guardar el archivo: %w", err)
	}

	return ArchivoGuardado{
		NombreArchivo: nombreOriginal,
		UrlArchivo:    fmt.Sprintf("%s/%s", a.urlPublica, nombreEnDisco),
		Tipo:          tipo,
		TamanoBytes:   escritos,
	}, nil
}

func (a *Almacen) Eliminar(urlArchivo string) error {
	nombre := filepath.Base(urlArchivo)
	if nombre == "" || nombre == "." || nombre == "/" || strings.Contains(nombre, "..") {
		return nil
	}

	if err := os.Remove(filepath.Join(a.directorio, nombre)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("no se pudo eliminar el archivo: %w", err)
	}

	return nil
}
