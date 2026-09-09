package correspondence

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
)

var patronVariable = regexp.MustCompile(`\{([a-zA-Z0-9_]{1,50})\}`)

type Resultado struct {
	Texto    string
	SinValor []string
}

func Variables(texto string) []string {
	encontradas := patronVariable.FindAllStringSubmatch(texto, -1)

	unicas := make(map[string]struct{}, len(encontradas))
	for _, coincidencia := range encontradas {
		unicas[strings.ToLower(coincidencia[1])] = struct{}{}
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
		nombre := strings.ToLower(strings.Trim(coincidencia, "{}"))

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
				valores[strings.ToLower(clave)] = textoDeValor(crudo)
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
