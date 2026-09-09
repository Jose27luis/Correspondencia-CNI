package correspondence

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
)

var patronVariable = regexp.MustCompile(`\{([a-zA-Z0-9_\- áéíóúÁÉÍÓÚñÑ]{1,60})\}|«([^»]{1,60})»`)

var alias = map[string]string{
	"nombres":            "nombre",
	"nombre_contacto":    "nombre",
	"contacto":           "nombre",
	"razon_social":       "empresa",
	"empresas":           "empresa",
	"compania":           "empresa",
	"email":              "correo",
	"correo_electronico": "correo",
	"paises":             "pais",
}

type Resultado struct {
	Texto    string
	SinValor []string
}

func Variables(texto string) []string {
	encontradas := patronVariable.FindAllStringSubmatch(texto, -1)

	unicas := make(map[string]struct{}, len(encontradas))
	for _, coincidencia := range encontradas {
		if nombre := nombreVariable(coincidencia); nombre != "" {
			unicas[nombre] = struct{}{}
		}
	}

	nombres := make([]string, 0, len(unicas))
	for nombre := range unicas {
		nombres = append(nombres, nombre)
	}
	sort.Strings(nombres)

	return nombres
}

func Renderizar(texto string, contacto contacts.Contacto) Resultado {
	valores := valoresDeContacto(contacto)
	faltantes := make(map[string]struct{})

	renderizado := patronVariable.ReplaceAllStringFunc(texto, func(coincidencia string) string {
		nombre := nombreVariable(patronVariable.FindStringSubmatch(coincidencia))
		if nombre == "" {
			return coincidencia
		}

		if valor, existe := valores[nombre]; existe && valor != "" {
			return valor
		}

		faltantes[nombre] = struct{}{}
		return coincidencia
	})

	sinValor := make([]string, 0, len(faltantes))
	for nombre := range faltantes {
		sinValor = append(sinValor, nombre)
	}
	sort.Strings(sinValor)

	return Resultado{Texto: renderizado, SinValor: sinValor}
}

func nombreVariable(coincidencia []string) string {
	if len(coincidencia) < 3 {
		return ""
	}

	crudo := coincidencia[1]
	if crudo == "" {
		crudo = coincidencia[2]
	}

	return NormalizarNombreVariable(crudo)
}

func NormalizarNombreVariable(crudo string) string {
	var construida strings.Builder

	for _, caracter := range strings.ToLower(strings.TrimSpace(crudo)) {
		switch {
		case caracter >= 'a' && caracter <= 'z', caracter >= '0' && caracter <= '9':
			construida.WriteRune(caracter)
		case caracter == ' ', caracter == '-', caracter == '_':
			construida.WriteRune('_')
		case caracter == 'á':
			construida.WriteRune('a')
		case caracter == 'é':
			construida.WriteRune('e')
		case caracter == 'í':
			construida.WriteRune('i')
		case caracter == 'ó':
			construida.WriteRune('o')
		case caracter == 'ú':
			construida.WriteRune('u')
		case caracter == 'ñ':
			construida.WriteRune('n')
		}
	}

	nombre := strings.Trim(construida.String(), "_")

	if equivalente, existe := alias[nombre]; existe {
		return equivalente
	}

	return nombre
}

func valoresDeContacto(contacto contacts.Contacto) map[string]string {
	valores := map[string]string{
		"nombre":  contacto.Nombre,
		"empresa": contacto.Empresa,
		"correo":  contacto.Correo,
	}

	if contacto.Pais != nil {
		valores["pais"] = *contacto.Pais
	}

	if len(contacto.CamposExtra) > 0 {
		var extras map[string]json.RawMessage
		if err := json.Unmarshal(contacto.CamposExtra, &extras); err == nil {
			for clave, crudo := range extras {
				valores[NormalizarNombreVariable(clave)] = textoDeValor(crudo)
			}
		}
	}

	return valores
}

func textoDeValor(crudo json.RawMessage) string {
	var texto string
	if err := json.Unmarshal(crudo, &texto); err == nil {
		return texto
	}
	return strings.Trim(string(crudo), `"`)
}
