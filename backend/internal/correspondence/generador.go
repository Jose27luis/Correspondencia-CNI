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

	if entrada.Name == configuracionWord {
		contenido = quitarCombinacionCorrespondencia(contenido)
	}

	if debePersonalizarse(entrada.Name) {
		contenido = reemplazarEnXML(aplanarCamposCombinacion(contenido), contacto)
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
		if !strings.ContainsAny(original, "{«<") {
			continue
		}

		nuevos, cambio := redistribuir(original, bloque.fragmento, contacto)
		if !cambio {
			continue
		}

		salida.Write(contenido[ultimo:bloque.fragmento[0].inicio])

		for posicion, fragmento := range bloque.fragmento {
			if posicion > 0 {
				salida.Write(contenido[bloque.fragmento[posicion-1].fin:fragmento.inicio])
			}
			xml.EscapeText(&salida, []byte(nuevos[posicion]))
		}

		ultimo = bloque.fragmento[len(bloque.fragmento)-1].fin
	}

	salida.Write(contenido[ultimo:])

	return salida.Bytes()
}

func redistribuir(original string, fragmentos []posicionTexto, contacto contacts.Contacto) ([]string, bool) {
	coincidencias := patronVariable.FindAllStringIndex(original, -1)
	if len(coincidencias) == 0 {
		return nil, false
	}

	limites := make([]int, len(fragmentos)+1)
	for posicion, fragmento := range fragmentos {
		limites[posicion+1] = limites[posicion] + len(fragmento.texto)
	}

	nuevos := make([]strings.Builder, len(fragmentos))

	copiar := func(desde int, hasta int) {
		for posicion := range fragmentos {
			inicio := max(desde, limites[posicion])
			fin := min(hasta, limites[posicion+1])
			if inicio < fin {
				nuevos[posicion].WriteString(original[inicio:fin])
			}
		}
	}

	fragmentoDe := func(indice int) int {
		for posicion := range fragmentos {
			if indice < limites[posicion+1] {
				return posicion
			}
		}
		return len(fragmentos) - 1
	}

	cambio := false
	cursor := 0

	for _, coincidencia := range coincidencias {
		inicio, fin := coincidencia[0], coincidencia[1]
		copiar(cursor, inicio)

		variable := original[inicio:fin]
		reemplazo := Renderizar(variable, contacto).Texto
		if reemplazo != variable {
			cambio = true
		}

		nuevos[fragmentoDe(inicio)].WriteString(reemplazo)
		cursor = fin
	}

	copiar(cursor, len(original))

	resultado := make([]string, len(fragmentos))
	for posicion := range nuevos {
		resultado[posicion] = nuevos[posicion].String()
	}

	return resultado, cambio
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
