package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	tokensCoincidencias = 4000
	puntajeMinimo       = 60
	maximoCandidatas    = 60
)

type PublicacionResumen struct {
	ID          string
	Titulo      string
	Descripcion string
	Empresa     string
	Ciudad      string
	Pais        string
}

type CoincidenciaSugerida struct {
	ID      string `json:"id"`
	Puntaje int    `json:"puntaje"`
	Motivo  string `json:"motivo"`
}

type respuestaCoincidencias struct {
	Coincidencias []CoincidenciaSugerida `json:"coincidencias"`
}

const instruccionesCoincidencias = `Eres el analista de la rueda de negocios de CNI (Corporación de Negocios
Interoceánicos), que conecta empresas de Perú y Brasil que ofrecen productos o servicios con empresas
que los necesitan.

Recibirás una publicación de referencia (una oferta o una demanda) y una lista de publicaciones del
tipo contrario. Indica cuáles encajan de verdad como posible negocio entre ambas empresas.

Reglas:
- Solo cuentan coincidencias reales de producto o servicio: lo que una ofrece debe responder a lo que
  la otra necesita. Parecerse en palabras no basta.
- No inventes capacidades, volúmenes, precios ni certificaciones que no estén en los textos.
- Puntúa de 0 a 100 qué tan buena es la coincidencia. Incluye solo las de 60 o más.
- Es correcto y esperado devolver una lista vacía si ninguna encaja.
- El motivo es una o dos frases concretas en español que explican por qué encajan, útiles para
  presentarlas entre sí.
- Usa exactamente los identificadores recibidos.

Responde únicamente con un objeto JSON con esta forma exacta, sin texto adicional ni bloques de código:
{"coincidencias": [{"id": "...", "puntaje": 0, "motivo": "..."}]}`

func (a *Asistente) SugerirCoincidencias(
	ctx context.Context,
	referencia PublicacionResumen,
	tipoReferencia string,
	candidatas []PublicacionResumen,
) ([]CoincidenciaSugerida, error) {
	if !a.Disponible() {
		return nil, ErrNoConfigurado
	}

	if len(candidatas) == 0 {
		return []CoincidenciaSugerida{}, nil
	}

	if len(candidatas) > maximoCandidatas {
		candidatas = candidatas[:maximoCandidatas]
	}

	var peticion strings.Builder
	fmt.Fprintf(&peticion, "Publicación de referencia (%s):\n", tipoReferencia)
	escribirPublicacion(&peticion, referencia)
	peticion.WriteString("\nPublicaciones candidatas:\n")
	for _, candidata := range candidatas {
		escribirPublicacion(&peticion, candidata)
	}

	texto, err := a.pedir(ctx, instruccionesCoincidencias, peticion.String(), tokensCoincidencias)
	if err != nil {
		return nil, err
	}

	var respuesta respuestaCoincidencias
	if err := json.Unmarshal([]byte(texto), &respuesta); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSinRespuesta, err)
	}

	return filtrarCoincidencias(respuesta.Coincidencias, candidatas), nil
}

func filtrarCoincidencias(
	sugeridas []CoincidenciaSugerida,
	candidatas []PublicacionResumen,
) []CoincidenciaSugerida {
	respuesta := respuestaCoincidencias{Coincidencias: sugeridas}

	validas := make(map[string]struct{}, len(candidatas))
	for _, candidata := range candidatas {
		validas[candidata.ID] = struct{}{}
	}

	resultado := make([]CoincidenciaSugerida, 0, len(respuesta.Coincidencias))
	vistas := make(map[string]struct{}, len(respuesta.Coincidencias))

	for _, coincidencia := range respuesta.Coincidencias {
		if _, existe := validas[coincidencia.ID]; !existe {
			continue
		}
		if _, repetida := vistas[coincidencia.ID]; repetida {
			continue
		}
		if coincidencia.Puntaje < puntajeMinimo || coincidencia.Puntaje > 100 {
			continue
		}
		if strings.TrimSpace(coincidencia.Motivo) == "" {
			continue
		}

		vistas[coincidencia.ID] = struct{}{}
		coincidencia.Motivo = strings.TrimSpace(coincidencia.Motivo)
		resultado = append(resultado, coincidencia)
	}

	return resultado
}

func escribirPublicacion(destino *strings.Builder, publicacion PublicacionResumen) {
	fmt.Fprintf(
		destino,
		"- id: %s | empresa: %s (%s, %s) | título: %s | descripción: %s\n",
		publicacion.ID,
		publicacion.Empresa,
		publicacion.Ciudad,
		publicacion.Pais,
		publicacion.Titulo,
		publicacion.Descripcion,
	)
}
