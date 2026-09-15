package intake

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/lists"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
)

type Incorporador struct {
	consultas *sqlcgen.Queries
	contactos *contacts.Repositorio
	listas    *lists.Repositorio
}

type Solicitud struct {
	Nombre           string
	Empresa          string
	Correo           string
	Pais             *string
	Datos            map[string]string
	NombreLista      string
	DescripcionLista string
}

func NuevoIncorporador(pool *pgxpool.Pool) *Incorporador {
	return &Incorporador{
		consultas: sqlcgen.New(pool),
		contactos: contacts.NuevoRepositorio(pool),
		listas:    lists.NuevoRepositorio(pool),
	}
}

func (i *Incorporador) Incorporar(ctx context.Context, solicitud Solicitud) (uuid.UUID, error) {
	datos := make(map[string]string, len(solicitud.Datos))
	for clave, valor := range solicitud.Datos {
		if valor = strings.TrimSpace(valor); valor != "" {
			datos[clave] = valor
		}
	}

	camposExtra, err := json.Marshal(datos)
	if err != nil {
		return uuid.Nil, fmt.Errorf("no se pudieron preparar los datos del contacto: %w", err)
	}

	entrada := contacts.EntradaContacto{
		Nombre:      solicitud.Nombre,
		Empresa:     solicitud.Empresa,
		Correo:      solicitud.Correo,
		Pais:        solicitud.Pais,
		CamposExtra: camposExtra,
	}
	entrada.Normalizar()

	contacto, _, err := i.contactos.Importar(ctx, entrada)
	if err != nil {
		return uuid.Nil, fmt.Errorf("no se pudo registrar el contacto: %w", err)
	}

	listaID, err := i.asegurarLista(ctx, solicitud.NombreLista, solicitud.DescripcionLista)
	if err != nil {
		return uuid.Nil, err
	}

	if _, err := i.listas.AgregarContactos(ctx, listaID, []uuid.UUID{contacto.ID}); err != nil {
		return uuid.Nil, fmt.Errorf("no se pudo agregar el contacto a la lista: %w", err)
	}

	return contacto.ID, nil
}

func (i *Incorporador) asegurarLista(ctx context.Context, nombre string, descripcion string) (uuid.UUID, error) {
	existente, err := i.consultas.ObtenerListaPorNombre(ctx, nombre)
	if err == nil {
		return existente.ID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("no se pudo buscar la lista %s: %w", nombre, err)
	}

	creada, err := i.listas.Crear(ctx, lists.EntradaLista{Nombre: nombre, Descripcion: &descripcion})
	if err != nil {
		return uuid.Nil, fmt.Errorf("no se pudo crear la lista %s: %w", nombre, err)
	}

	return creada.ID, nil
}
