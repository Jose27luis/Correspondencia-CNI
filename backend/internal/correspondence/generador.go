package correspondence

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
)

type parrafo struct {
	inicio    int
	fin       int
	fragmento []posicionTexto
}

type posicionTexto struct {
	inicio int
	fin    int
	texto  string
}

func GenerarDocumento(plantilla []byte, contacto contacts.Contacto) ([]byte, error) {
	lector, err := zip.NewReader(bytes.NewReader(plantilla), int64(len(plantilla)))
	if err != nil {
		return nil, ErrWordInvalido
	}

	var salida bytes.Buffer
	escritor := zip.NewWriter(&salida)

	for _, entrada := range lector.File {
		if err := copiarEntrada(escritor, entrada, contacto); err != nil {
			return nil, err
		}
	}

	if err := escritor.Close(); err != nil {
		return nil, fmt.Errorf("no se pudo cerrar el documento generado: %w", err)
	}

	return salida.Bytes(), nil
}

func copiarEntrada(escritor *zip.Writer, entrada *zip.File, contacto contacts.Contacto) error {
	abierto, err := entrada.Open()
	if err != nil {
		return ErrWordInvalido
	}
	defer abierto.Close()

	contenido, err := io.ReadAll(abierto)
	if err != nil {
		return ErrWordInvalido
	}

	if debePersonalizarse(entrada.Name) {
		contenido = reemplazarEnXML(contenido, contacto)
	}

	destino, err := escritor.Create(entrada.Name)
	if err != nil {
		return fmt.Errorf("no se pudo escribir el documento generado: %w", err)
	}

	if _, err := destino.Write(contenido); err != nil {
		return fmt.Errorf("no se pudo escribir el documento generado: %w", err)
	}

	return nil
}

func debePersonalizarse(nombre string) bool {
	if !strings.HasSuffix(nombre, ".xml") {
		return false
	}

	return nombre == documentoPrincipal ||
		strings.HasPrefix(nombre, "word/header") ||
		strings.HasPrefix(nombre, "word/footer")
}

func reemplazarEnXML(contenido []byte, contacto contacts.Contacto) []byte {
	parrafos, err := ubicarParrafos(contenido)
	if err != nil {
		return contenido
	}

	var salida bytes.Buffer
	ultimo := 0

	for _, bloque := range parrafos {
		original := unirFragmentos(bloque.fragmento)
		if !strings.Contains(original, "{") {
			continue
		}

		resultado := Renderizar(original, contacto)
		if resultado.Texto == original {
			continue
		}

		salida.Write(contenido[ultimo:bloque.fragmento[0].inicio])
		xml.EscapeText(&salida, []byte(resultado.Texto))

		for posicion := 1; posicion < len(bloque.fragmento); posicion++ {
			salida.Write(contenido[bloque.fragmento[posicion-1].fin:bloque.fragmento[posicion].inicio])
		}

		ultimo = bloque.fragmento[len(bloque.fragmento)-1].fin
	}

	salida.Write(contenido[ultimo:])

	return salida.Bytes()
}

func unirFragmentos(fragmentos []posicionTexto) string {
	var construido strings.Builder
	for _, fragmento := range fragmentos {
		construido.WriteString(fragmento.texto)
	}
	return construido.String()
}

func ubicarParrafos(contenido []byte) ([]parrafo, error) {
	decodificador := xml.NewDecoder(bytes.NewReader(contenido))

	var parrafos []parrafo
	var actual *parrafo
	var enTexto bool

	for {
		desplazamientoPrevio := decodificador.InputOffset()

		elemento, err := decodificador.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, ErrWordInvalido
		}

		switch valor := elemento.(type) {
		case xml.StartElement:
			switch valor.Name.Local {
			case "p":
				actual = &parrafo{inicio: int(desplazamientoPrevio)}
			case "t":
				enTexto = true
			}
		case xml.EndElement:
			switch valor.Name.Local {
			case "p":
				if actual != nil && len(actual.fragmento) > 0 {
					actual.fin = int(decodificador.InputOffset())
					parrafos = append(parrafos, *actual)
				}
				actual = nil
			case "t":
				enTexto = false
			}
		case xml.CharData:
			if enTexto && actual != nil {
				actual.fragmento = append(actual.fragmento, posicionTexto{
					inicio: int(desplazamientoPrevio),
					fin:    int(decodificador.InputOffset()),
					texto:  string(valor),
				})
			}
		}
	}

	return parrafos, nil
}
