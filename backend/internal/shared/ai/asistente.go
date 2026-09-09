package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	modelo          = "claude-opus-5"
	tokensRedaccion = 8000
	tokensRevision  = 4000
)

var (
	ErrNoConfigurado = errors.New("el asistente de redacción no está configurado")
	ErrSinRespuesta  = errors.New("el asistente no devolvió una respuesta utilizable")
)

type Asistente struct {
	cliente *anthropic.Client
}

func NuevoAsistente(claveAPI string) *Asistente {
	if strings.TrimSpace(claveAPI) == "" {
		return &Asistente{}
	}

	cliente := anthropic.NewClient(option.WithAPIKey(claveAPI))

	return &Asistente{cliente: &cliente}
}

func (a *Asistente) Disponible() bool {
	return a.cliente != nil
}

type EntradaRedaccion struct {
	Instruccion string   `json:"instruccion"`
	Asunto      string   `json:"asunto"`
	Cuerpo      string   `json:"cuerpo"`
	Variables   []string `json:"variables"`
}

type Borrador struct {
	Asunto string `json:"asunto"`
	Cuerpo string `json:"cuerpo"`
}

type Hallazgo struct {
	Gravedad   string `json:"gravedad"`
	Titulo     string `json:"titulo"`
	Detalle    string `json:"detalle"`
	Sugerencia string `json:"sugerencia"`
}

type Revision struct {
	Veredicto string     `json:"veredicto"`
	Resumen   string     `json:"resumen"`
	Hallazgos []Hallazgo `json:"hallazgos"`
}

const instruccionesRedaccion = `Eres el asistente de redacción comercial de CNI (Corporación de Negocios Interoceánicos),
una empresa peruana que exporta granos y ofrece servicios de transporte y logística entre Perú y Brasil.

Escribes cartas comerciales dirigidas de empresa a empresa. Reglas de estilo:

- Trato de usted, formal pero directo. Nada de lenguaje publicitario ni superlativos vacíos.
- Estructura de carta comercial: saludo, propósito, oferta concreta, cierre con llamada a la acción.
- Nunca inventes cifras, precios, volúmenes, plazos ni certificaciones que no estén en la instrucción.
  Si falta un dato necesario, deja una variable entre llaves para que el usuario la complete.
- Usa las variables de personalización indicadas, escritas entre llaves.
- No incluyas la firma ni el membrete: el sistema los añade aparte.
- No uses emojis ni viñetas decorativas.

Responde únicamente con un objeto JSON con esta forma exacta, sin texto adicional ni bloques de código:
{"asunto": "...", "cuerpo": "..."}

El asunto debe ser concreto y no parecer correo masivo. El cuerpo usa saltos de línea reales.`

const instruccionesRevision = `Eres el revisor de correspondencia comercial de CNI. Vas a revisar una carta
que está a punto de enviarse a decenas o cientos de empresas. Un error aquí no se puede recoger.

Revisa y reporta solo problemas reales:
- Ortografía, gramática y puntuación.
- Variables mal escritas, sin cerrar, o que no existen en la lista de variables disponibles.
- Cifras, plazos o compromisos ambiguos que puedan interpretarse mal por el cliente.
- Tono inadecuado para correspondencia comercial entre empresas.
- Elementos que aumenten la probabilidad de caer en spam: asunto en mayúsculas, exceso de
  signos de exclamación, promesas exageradas.
- Datos que parezcan de marcador de posición y se hayan quedado sin reemplazar.

No inventes problemas. Si la carta está bien, dilo. No comentes el estilo por preferencia personal.

Responde únicamente con un objeto JSON con esta forma exacta, sin texto adicional ni bloques de código:
{"veredicto": "lista|revisar|no_enviar", "resumen": "...", "hallazgos": [
  {"gravedad": "alta|media|baja", "titulo": "...", "detalle": "...", "sugerencia": "..."}
]}

Usa "lista" si puede enviarse tal cual, "revisar" si hay detalles menores, y "no_enviar" solo si
hay un error que dañaría la imagen de CNI ante sus clientes.`

func (a *Asistente) Redactar(ctx context.Context, entrada EntradaRedaccion) (Borrador, error) {
	if !a.Disponible() {
		return Borrador{}, ErrNoConfigurado
	}

	var peticion strings.Builder
	peticion.WriteString("Lo que se quiere comunicar:\n")
	peticion.WriteString(entrada.Instruccion)

	if len(entrada.Variables) > 0 {
		peticion.WriteString("\n\nVariables disponibles de cada empresa destinataria: ")
		peticion.WriteString(strings.Join(entrada.Variables, ", "))
	}

	if strings.TrimSpace(entrada.Cuerpo) != "" {
		peticion.WriteString("\n\nCarta actual que debes mejorar o reescribir:\n")
		peticion.WriteString("Asunto: " + entrada.Asunto + "\n\n")
		peticion.WriteString(entrada.Cuerpo)
	}

	texto, err := a.pedir(ctx, instruccionesRedaccion, peticion.String(), tokensRedaccion)
	if err != nil {
		return Borrador{}, err
	}

	var borrador Borrador
	if err := json.Unmarshal([]byte(texto), &borrador); err != nil {
		return Borrador{}, fmt.Errorf("%w: %v", ErrSinRespuesta, err)
	}

	if strings.TrimSpace(borrador.Cuerpo) == "" {
		return Borrador{}, ErrSinRespuesta
	}

	return borrador, nil
}

func (a *Asistente) Revisar(ctx context.Context, asunto string, cuerpo string, variables []string) (Revision, error) {
	if !a.Disponible() {
		return Revision{}, ErrNoConfigurado
	}

	var peticion strings.Builder
	peticion.WriteString("Asunto: ")
	peticion.WriteString(asunto)
	peticion.WriteString("\n\nCuerpo:\n")
	peticion.WriteString(cuerpo)

	if len(variables) > 0 {
		peticion.WriteString("\n\nVariables disponibles de cada empresa: ")
		peticion.WriteString(strings.Join(variables, ", "))
	}

	texto, err := a.pedir(ctx, instruccionesRevision, peticion.String(), tokensRevision)
	if err != nil {
		return Revision{}, err
	}

	var revision Revision
	if err := json.Unmarshal([]byte(texto), &revision); err != nil {
		return Revision{}, fmt.Errorf("%w: %v", ErrSinRespuesta, err)
	}

	if revision.Hallazgos == nil {
		revision.Hallazgos = []Hallazgo{}
	}

	return revision, nil
}

func (a *Asistente) pedir(ctx context.Context, sistema string, peticion string, tokens int64) (string, error) {
	adaptativo := anthropic.ThinkingConfigAdaptiveParam{}

	respuesta, err := a.cliente.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     modelo,
		MaxTokens: tokens,
		Thinking:  anthropic.ThinkingConfigParamUnion{OfAdaptive: &adaptativo},
		System: []anthropic.TextBlockParam{{
			Text:         sistema,
			CacheControl: anthropic.NewCacheControlEphemeralParam(),
		}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(peticion)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("no se pudo consultar al asistente: %w", err)
	}

	if respuesta.StopReason == anthropic.StopReasonRefusal {
		return "", fmt.Errorf("%w: el asistente declinó la solicitud", ErrSinRespuesta)
	}

	var construido strings.Builder
	for _, bloque := range respuesta.Content {
		if texto, ok := bloque.AsAny().(anthropic.TextBlock); ok {
			construido.WriteString(texto.Text)
		}
	}

	return extraerJSON(construido.String()), nil
}

func extraerJSON(texto string) string {
	recortado := strings.TrimSpace(texto)

	if despues, encontrado := strings.CutPrefix(recortado, "```json"); encontrado {
		recortado = despues
	} else if despues, encontrado := strings.CutPrefix(recortado, "```"); encontrado {
		recortado = despues
	}

	recortado = strings.TrimSuffix(strings.TrimSpace(recortado), "```")

	inicio := strings.Index(recortado, "{")
	fin := strings.LastIndex(recortado, "}")

	if inicio >= 0 && fin > inicio {
		return recortado[inicio : fin+1]
	}

	return strings.TrimSpace(recortado)
}
