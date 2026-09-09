package correspondence

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	documentoPrincipal = "word/document.xml"
	tamanoMaximoWord   = 10 << 20
)

var (
	ErrWordInvalido = errors.New("el archivo no es un documento de Word válido")
	ErrWordVacio    = errors.New("el documento no tiene texto")
)

type ContenidoWord struct {
	Cuerpo    string   `json:"cuerpo"`
	Variables []string `json:"variables"`
}

func LeerWord(lector io.Reader, tamano int64) (ContenidoWord, error) {
	if tamano > tamanoMaximoWord {
		return ContenidoWord{}, fmt.Errorf("el documento supera los %d MB", tamanoMaximoWord>>20)
	}

	contenido, err := io.ReadAll(io.LimitReader(lector, tamanoMaximoWord))
	if err != nil {
		return ContenidoWord{}, fmt.Errorf("no se pudo leer el archivo: %w", err)
	}

	archivo, err := zip.NewReader(strings.NewReader(string(contenido)), int64(len(contenido)))
	if err != nil {
		return ContenidoWord{}, ErrWordInvalido
	}

	documento, err := abrirDocumento(archivo)
	if err != nil {
		return ContenidoWord{}, err
	}

	texto, err := extraerTexto(documento)
	if err != nil {
		return ContenidoWord{}, err
	}

	texto = strings.TrimSpace(texto)
	if texto == "" {
		return ContenidoWord{}, ErrWordVacio
	}

	return ContenidoWord{Cuerpo: texto, Variables: Variables(texto)}, nil
}

func abrirDocumento(archivo *zip.Reader) (io.ReadCloser, error) {
	for _, entrada := range archivo.File {
		if entrada.Name != documentoPrincipal {
			continue
		}

		abierto, err := entrada.Open()
		if err != nil {
			return nil, ErrWordInvalido
		}

		return abierto, nil
	}

	return nil, ErrWordInvalido
}

func extraerTexto(documento io.ReadCloser) (string, error) {
	defer documento.Close()

	decodificador := xml.NewDecoder(documento)
	var construido strings.Builder
	var enTexto bool

	for {
		elemento, err := decodificador.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", ErrWordInvalido
		}

		switch valor := elemento.(type) {
		case xml.StartElement:
			switch valor.Name.Local {
			case "t":
				enTexto = true
			case "tab":
				construido.WriteString("\t")
			case "br", "cr":
				construido.WriteString("\n")
			}
		case xml.EndElement:
			switch valor.Name.Local {
			case "t":
				enTexto = false
			case "p":
				construido.WriteString("\n")
			}
		case xml.CharData:
			if enTexto {
				construido.Write(valor)
			}
		}
	}

	return normalizarSaltos(construido.String()), nil
}

func normalizarSaltos(texto string) string {
	lineas := strings.Split(texto, "\n")
	limpias := make([]string, 0, len(lineas))
	vaciasSeguidas := 0

	for _, linea := range lineas {
		recortada := strings.TrimRight(linea, " \t")

		if recortada == "" {
			vaciasSeguidas++
			if vaciasSeguidas > 1 {
				continue
			}
		} else {
			vaciasSeguidas = 0
		}

		limpias = append(limpias, recortada)
	}

	return strings.Join(limpias, "\n")
}
