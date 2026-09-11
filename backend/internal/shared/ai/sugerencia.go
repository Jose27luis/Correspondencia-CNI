package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	tokensSugerencia     = 8000
	maximoTextoDocumento = 60000
)

type EntradaSugerencia struct {
	Documento string   `json:"documento"`
	Asunto    string   `json:"asunto"`
	Variables []string `json:"variables"`
}

type Sugerencia struct {
	Asunto  string `json:"asunto"`
	Cuerpo  string `json:"cuerpo"`
	Resumen string `json:"resumen"`
}

const instruccionesSugerencia = `Eres el asistente comercial de CNI (Corporación de Negocios Interoceánicos),
empresa peruana que exporta granos y ofrece transporte y logística entre Perú y Brasil.

Recibirás el texto de una carta comercial redactada en Word. Esa carta viajará como documento
adjunto, personalizado para cada empresa. Tu tarea es redactar el cuerpo del correo electrónico
que acompaña a esa carta: un mensaje breve y profesional que presente la oferta e invite a abrir
el documento adjunto.

Reglas:
- Analiza la carta y resume su propuesta principal en dos o tres párrafos cortos.
- Trato de usted, formal y directo, sin lenguaje publicitario ni superlativos vacíos.
- Usa solo cifras, productos, volúmenes, plazos y certificaciones que aparezcan en la carta.
  Nunca inventes datos. Si la carta no trae un dato, no lo menciones.
- Conserva exactamente las variables de personalización que encuentres en la carta, tanto las
  escritas entre llaves como {empresa}, como las de combinación de Word como «NOMBRES».
  Puedes usar además las variables disponibles que se te indiquen.
- Menciona que el detalle completo va en el documento adjunto.
- Cierra con una invitación concreta a responder o coordinar.
- No incluyas firma, membrete ni datos de contacto: se añaden aparte.
- No uses emojis, viñetas decorativas ni markdown.

Responde únicamente con un objeto JSON con esta forma exacta, sin texto adicional ni bloques de código:
{"asunto": "...", "cuerpo": "...", "resumen": "..."}

"asunto": concreto, que no parezca correo masivo; si se te da un asunto actual y es adecuado, puedes
mantenerlo o mejorarlo. "cuerpo": el mensaje con saltos de línea reales. "resumen": una frase que
explique al usuario qué entendiste de la carta, para que confirme que el análisis es correcto.`

func (a *Asistente) SugerirCuerpo(ctx context.Context, entrada EntradaSugerencia) (Sugerencia, error) {
	if !a.Disponible() {
		return Sugerencia{}, ErrNoConfigurado
	}

	documento := strings.TrimSpace(entrada.Documento)
	if documento == "" {
		return Sugerencia{}, fmt.Errorf("%w: no hay documento que analizar", ErrSinRespuesta)
	}

	if len(documento) > maximoTextoDocumento {
		documento = documento[:maximoTextoDocumento]
	}

	var peticion strings.Builder
	peticion.WriteString("Texto de la carta en Word:\n\n")
	peticion.WriteString(documento)

	if asunto := strings.TrimSpace(entrada.Asunto); asunto != "" {
		peticion.WriteString("\n\nAsunto actual: ")
		peticion.WriteString(asunto)
	}

	if len(entrada.Variables) > 0 {
		peticion.WriteString("\n\nVariables disponibles de cada empresa: ")
		peticion.WriteString(strings.Join(entrada.Variables, ", "))
	}

	texto, err := a.pedir(ctx, instruccionesSugerencia, peticion.String(), tokensSugerencia)
	if err != nil {
		return Sugerencia{}, err
	}

	var sugerencia Sugerencia
	if err := json.Unmarshal([]byte(texto), &sugerencia); err != nil {
		return Sugerencia{}, fmt.Errorf("%w: %v", ErrSinRespuesta, err)
	}

	if strings.TrimSpace(sugerencia.Cuerpo) == "" {
		return Sugerencia{}, ErrSinRespuesta
	}

	return sugerencia, nil
}
