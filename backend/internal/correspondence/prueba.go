package correspondence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/auth"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/mailer"
)

type ResultadoPrueba struct {
	Destinatario string   `json:"destinatario"`
	Asunto       string   `json:"asunto"`
	Simulada     bool     `json:"simulada"`
	Variables    []string `json:"variables_sin_valor"`
}

func contactoDeEjemplo() contacts.Contacto {
	pais := "Perú"

	return contacts.Contacto{
		Nombre:      "Nombre de ejemplo",
		Empresa:     "Empresa de ejemplo",
		Correo:      "ejemplo@empresa.pe",
		Pais:        &pais,
		CamposExtra: json.RawMessage(`{"cargo":"Cargo de ejemplo"}`),
	}
}

func (s *Servicio) EnviarPrueba(ctx context.Context, id uuid.UUID) (ResultadoPrueba, error) {
	identidad, autenticado := auth.DesdeContexto(ctx)
	if !autenticado {
		return ResultadoPrueba{}, auth.ErrSinToken
	}

	pieza, err := s.repositorio.Obtener(ctx, id)
	if err != nil {
		return ResultadoPrueba{}, err
	}

	contacto := contactoDeEjemplo()
	if pieza.ListaID != nil {
		if real, err := s.repositorio.PrimerContactoDeLista(ctx, *pieza.ListaID); err == nil {
			contacto = real
		}
	}

	asunto := Renderizar(pieza.Asunto, contacto)
	cuerpo := Renderizar(pieza.Cuerpo, contacto)

	adjuntos, err := s.adjuntosDePrueba(pieza, contacto)
	if err != nil {
		return ResultadoPrueba{}, err
	}

	aviso := fmt.Sprintf(
		"Esta es una prueba dirigida a %s (%s). El envío real usará los datos de cada empresa.\n\n---\n\n",
		contacto.Empresa,
		contacto.Correo,
	)

	_, err = s.proveedor.Enviar(ctx, mailer.Mensaje{
		Para:     identidad.Correo,
		Asunto:   "[PRUEBA] " + asunto.Texto,
		Cuerpo:   aviso + cuerpo.Texto,
		Adjuntos: adjuntos,
	})
	if err != nil {
		return ResultadoPrueba{}, err
	}

	return ResultadoPrueba{
		Destinatario: identidad.Correo,
		Asunto:       asunto.Texto,
		Simulada:     s.proveedor.Nombre() == "bitacora",
		Variables:    unir(asunto.SinValor, cuerpo.SinValor),
	}, nil
}

func (s *Servicio) adjuntosDePrueba(pieza Correspondencia, contacto contacts.Contacto) ([]mailer.Adjunto, error) {
	adjuntos := make([]mailer.Adjunto, 0, len(pieza.Adjuntos)+1)

	for _, adjunto := range pieza.Adjuntos {
		adjuntos = append(adjuntos, mailer.Adjunto{
			NombreArchivo: adjunto.NombreArchivo,
			UrlArchivo:    adjunto.UrlArchivo,
		})
	}

	if pieza.PlantillaURL == nil {
		return adjuntos, nil
	}

	plantilla, err := s.almacen.Leer(*pieza.PlantillaURL)
	if err != nil {
		return nil, err
	}

	documento, err := GenerarDocumento(plantilla, contacto)
	if err != nil {
		return nil, err
	}

	nombre := "carta-prueba.docx"
	if pieza.PlantillaNombre != nil {
		nombre = *pieza.PlantillaNombre
	}

	return append(adjuntos, mailer.Adjunto{
		NombreArchivo: nombre,
		Contenido:     documento,
	}), nil
}
