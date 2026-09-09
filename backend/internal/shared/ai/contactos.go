package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	tokensExtraccion  = 8000
	tokensBusqueda    = 1000
	maximoTextoPegado = 40000
)

type ContactoExtraido struct {
	Nombre  string `json:"nombre"`
	Empresa string `json:"empresa"`
	Correo  string `json:"correo"`
	Pais    string `json:"pais"`
	Cargo   string `json:"cargo"`
}

type Extraccion struct {
	Contactos []ContactoExtraido `json:"contactos"`
	Aviso     string             `json:"aviso"`
}

type CriteriosBusqueda struct {
	Terminos    []string `json:"terminos"`
	Explicacion string   `json:"explicacion"`
}

const instruccionesExtraccion = `Extraes datos de contacto de empresas a partir de texto pegado por el usuario:
correos recibidos, firmas, directorios, listados copiados de una web o de un documento.

Reglas estrictas:
- Extrae solo lo que está literalmente en el texto. Nunca inventes ni deduzcas correos,
  nombres, cargos ni países que no aparezcan.
- Si un dato falta, deja la cadena vacía. Es correcto y esperado devolver campos vacíos.
- No incluyas entradas sin correo electrónico: sin correo el contacto no sirve.
- Descarta correos genéricos de sistemas (noreply, mailer-daemon, postmaster).
- Corrige solo el formato evidente: espacios sobrantes, mayúsculas de más en el correo.
- "empresa" es la razón social. Si el texto solo trae el dominio del correo, y no el nombre
  de la empresa, deja "empresa" vacío en lugar de inventarlo a partir del dominio.
- Si el texto no contiene ningún contacto, devuelve la lista vacía y explica por qué en "aviso".

Responde únicamente con un objeto JSON con esta forma exacta, sin texto adicional ni bloques de código:
{"contactos": [{"nombre": "...", "empresa": "...", "correo": "...", "pais": "...", "cargo": "..."}],
 "aviso": "..."}

En "aviso" señala brevemente lo que el usuario deba verificar: datos incompletos, entradas dudosas
o duplicados aparentes. Si no hay nada que advertir, deja "aviso" vacío.`

const instruccionesBusqueda = `Traduces una búsqueda escrita en lenguaje natural a términos de búsqueda
para una base de contactos de empresas de CNI, que exporta granos y ofrece transporte entre Perú y Brasil.

La base guarda por cada contacto: nombre de la persona, empresa, correo, país y campos adicionales
como el cargo o el rubro.

Devuelve los términos que deben buscarse como texto. Reglas:
- Usa palabras raíz que aparezcan literalmente en los datos: para "molinos" usa "molino",
  para "transportistas" usa "transport".
- Incluye el país como término aparte cuando la consulta lo mencione.
- Entre 1 y 4 términos. Menos términos es mejor que más.
- No inventes nombres de empresas concretas que la consulta no mencione.

Responde únicamente con un objeto JSON con esta forma exacta, sin texto adicional ni bloques de código:
{"terminos": ["...", "..."], "explicacion": "..."}

En "explicacion" describe en una frase corta qué se va a buscar, para que el usuario lo confirme.`

func (a *Asistente) ExtraerContactos(ctx context.Context, texto string) (Extraccion, error) {
	if !a.Disponible() {
		return Extraccion{}, ErrNoConfigurado
	}

	recortado := strings.TrimSpace(texto)
	if recortado == "" {
		return Extraccion{}, fmt.Errorf("%w: no se envió texto", ErrSinRespuesta)
	}

	if len(recortado) > maximoTextoPegado {
		recortado = recortado[:maximoTextoPegado]
	}

	respuesta, err := a.pedir(ctx, instruccionesExtraccion, recortado, tokensExtraccion)
	if err != nil {
		return Extraccion{}, err
	}

	var extraccion Extraccion
	if err := json.Unmarshal([]byte(respuesta), &extraccion); err != nil {
		return Extraccion{}, fmt.Errorf("%w: %v", ErrSinRespuesta, err)
	}

	extraccion.Contactos = limpiarExtraidos(extraccion.Contactos)

	return extraccion, nil
}

func (a *Asistente) InterpretarBusqueda(ctx context.Context, consulta string) (CriteriosBusqueda, error) {
	if !a.Disponible() {
		return CriteriosBusqueda{}, ErrNoConfigurado
	}

	respuesta, err := a.pedir(ctx, instruccionesBusqueda, strings.TrimSpace(consulta), tokensBusqueda)
	if err != nil {
		return CriteriosBusqueda{}, err
	}

	var criterios CriteriosBusqueda
	if err := json.Unmarshal([]byte(respuesta), &criterios); err != nil {
		return CriteriosBusqueda{}, fmt.Errorf("%w: %v", ErrSinRespuesta, err)
	}

	if criterios.Terminos == nil {
		criterios.Terminos = []string{}
	}

	return criterios, nil
}

func limpiarExtraidos(contactos []ContactoExtraido) []ContactoExtraido {
	limpios := make([]ContactoExtraido, 0, len(contactos))
	vistos := make(map[string]struct{}, len(contactos))

	for _, contacto := range contactos {
		correo := strings.ToLower(strings.TrimSpace(contacto.Correo))

		if correo == "" || !strings.Contains(correo, "@") {
			continue
		}

		if _, repetido := vistos[correo]; repetido {
			continue
		}
		vistos[correo] = struct{}{}

		limpios = append(limpios, ContactoExtraido{
			Nombre:  strings.TrimSpace(contacto.Nombre),
			Empresa: strings.TrimSpace(contacto.Empresa),
			Correo:  correo,
			Pais:    strings.TrimSpace(contacto.Pais),
			Cargo:   strings.TrimSpace(contacto.Cargo),
		})
	}

	return limpios
}
