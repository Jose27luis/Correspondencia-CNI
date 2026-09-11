package correspondence

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"regexp"
	"sort"
	"strings"
)

const configuracionWord = "word/settings.xml"

var patronCombinacion = regexp.MustCompile(`(?s)<w:mailMerge>.*?</w:mailMerge>|<w:mailMerge\s*/>`)

type tramo struct {
	inicio int
	fin    int
}

type ejecucion struct {
	inicio      int
	tipoCampo   string
	instruccion string
}

type campo struct {
	instruccion string
	controles   []tramo
}

func quitarCombinacionCorrespondencia(contenido []byte) []byte {
	return patronCombinacion.ReplaceAll(contenido, nil)
}

func aplanarCamposCombinacion(contenido []byte) []byte {
	decodificador := xml.NewDecoder(bytes.NewReader(contenido))

	var pila []*ejecucion
	var abiertos []*campo
	var eliminar []tramo
	enInstruccion := false

	for {
		previo := int(decodificador.InputOffset())

		elemento, err := decodificador.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return contenido
		}

		switch valor := elemento.(type) {
		case xml.StartElement:
			switch valor.Name.Local {
			case "r":
				pila = append(pila, &ejecucion{inicio: previo})
			case "fldChar":
				if len(pila) > 0 {
					for _, atributo := range valor.Attr {
						if atributo.Name.Local == "fldCharType" {
							pila[len(pila)-1].tipoCampo = atributo.Value
						}
					}
				}
			case "instrText":
				enInstruccion = true
			}
		case xml.CharData:
			if enInstruccion && len(pila) > 0 {
				pila[len(pila)-1].instruccion += string(valor)
			}
		case xml.EndElement:
			switch valor.Name.Local {
			case "instrText":
				enInstruccion = false
			case "r":
				if len(pila) == 0 {
					continue
				}

				actual := pila[len(pila)-1]
				pila = pila[:len(pila)-1]
				rango := tramo{inicio: actual.inicio, fin: int(decodificador.InputOffset())}

				switch {
				case actual.tipoCampo == "begin":
					abiertos = append(abiertos, &campo{controles: []tramo{rango}})
				case len(abiertos) == 0:
				case actual.tipoCampo == "separate":
					abierto := abiertos[len(abiertos)-1]
					abierto.controles = append(abierto.controles, rango)
				case actual.tipoCampo == "end":
					abierto := abiertos[len(abiertos)-1]
					abiertos = abiertos[:len(abiertos)-1]
					abierto.controles = append(abierto.controles, rango)
					if esCampoDeCombinacion(abierto.instruccion) {
						eliminar = append(eliminar, abierto.controles...)
					}
				case actual.instruccion != "":
					abierto := abiertos[len(abiertos)-1]
					abierto.instruccion += actual.instruccion
					abierto.controles = append(abierto.controles, rango)
				}
			}
		}
	}

	if len(eliminar) == 0 {
		return contenido
	}

	sort.Slice(eliminar, func(i, j int) bool { return eliminar[i].inicio < eliminar[j].inicio })

	var salida bytes.Buffer
	ultimo := 0
	for _, rango := range eliminar {
		if rango.inicio < ultimo {
			continue
		}
		salida.Write(contenido[ultimo:rango.inicio])
		ultimo = rango.fin
	}
	salida.Write(contenido[ultimo:])

	return salida.Bytes()
}

func esCampoDeCombinacion(instruccion string) bool {
	campos := strings.Fields(strings.ToUpper(instruccion))
	return len(campos) > 0 && campos[0] == "MERGEFIELD"
}
