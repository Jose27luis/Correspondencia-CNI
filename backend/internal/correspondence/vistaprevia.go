package correspondence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
)

var ErrSinPlantilla = errors.New("esta carta no tiene una carta en Word personalizada")

type VistaPrevia struct {
	Documento     []byte
	NombreArchivo string
}

func (r *Repositorio) ContactoPorID(ctx context.Context, id uuid.UUID) (contacts.Contacto, error) {
	fila, err := r.consultas.ObtenerContactoPorID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return contacts.Contacto{}, ErrSinDestinatario
		}
		return contacts.Contacto{}, fmt.Errorf("no se pudo obtener el contacto: %w", err)
	}
	return contacts.DesdeFila(fila), nil
}

func (s *Servicio) GenerarVistaPrevia(ctx context.Context, id uuid.UUID, contactoID *uuid.UUID) (VistaPrevia, error) {
	pieza, err := s.repositorio.Obtener(ctx, id)
	if err != nil {
		return VistaPrevia{}, err
	}

	if pieza.PlantillaURL == nil {
		return VistaPrevia{}, ErrSinPlantilla
	}

	contacto, err := s.contactoParaVistaPrevia(ctx, pieza, contactoID)
	if err != nil {
		return VistaPrevia{}, err
	}

	plantilla, err := s.almacen.Leer(*pieza.PlantillaURL)
	if err != nil {
		return VistaPrevia{}, err
	}

	documento, err := GenerarDocumento(plantilla, contacto)
	if err != nil {
		return VistaPrevia{}, err
	}

	nombre := "carta.docx"
	if pieza.PlantillaNombre != nil {
		nombre = *pieza.PlantillaNombre
	}

	return VistaPrevia{Documento: documento, NombreArchivo: nombre}, nil
}

func (s *Servicio) contactoParaVistaPrevia(
	ctx context.Context,
	pieza Correspondencia,
	contactoID *uuid.UUID,
) (contacts.Contacto, error) {
	if contactoID != nil {
		return s.repositorio.ContactoPorID(ctx, *contactoID)
	}

	if pieza.ListaID != nil {
		contacto, err := s.repositorio.PrimerContactoDeLista(ctx, *pieza.ListaID)
		if err == nil {
			return contacto, nil
		}
		if !errors.Is(err, ErrSinDestinatario) {
			return contacts.Contacto{}, err
		}
	}

	return contactoDeEjemplo(), nil
}
