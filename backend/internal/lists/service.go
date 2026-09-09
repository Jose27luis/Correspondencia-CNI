package lists

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const (
	limitePorDefecto = 50
	limiteMaximo     = 200
)

type Servicio struct {
	repositorio *Repositorio
	validador   *validator.Validate
}

func NuevoServicio(repositorio *Repositorio) *Servicio {
	return &Servicio{
		repositorio: repositorio,
		validador:   validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (s *Servicio) Crear(ctx context.Context, entrada EntradaLista) (Lista, error) {
	entrada.Normalizar()
	if err := s.validador.Struct(entrada); err != nil {
		return Lista{}, err
	}
	return s.repositorio.Crear(ctx, entrada)
}

func (s *Servicio) Obtener(ctx context.Context, id uuid.UUID) (Lista, error) {
	return s.repositorio.Obtener(ctx, id)
}

func (s *Servicio) Listar(ctx context.Context, limite int32, desfase int32) (ListadoListas, error) {
	limite, desfase = ajustarPaginado(limite, desfase)
	return s.repositorio.Listar(ctx, limite, desfase)
}

func (s *Servicio) Actualizar(ctx context.Context, id uuid.UUID, entrada EntradaLista) (Lista, error) {
	entrada.Normalizar()
	if err := s.validador.Struct(entrada); err != nil {
		return Lista{}, err
	}
	return s.repositorio.Actualizar(ctx, id, entrada)
}

func (s *Servicio) Eliminar(ctx context.Context, id uuid.UUID) error {
	return s.repositorio.Eliminar(ctx, id)
}

func (s *Servicio) AgregarContactos(ctx context.Context, listaID uuid.UUID, entrada EntradaMiembros) (ResultadoMiembros, error) {
	if err := s.validador.Struct(entrada); err != nil {
		return ResultadoMiembros{}, err
	}

	agregados, err := s.repositorio.AgregarContactos(ctx, listaID, entrada.ContactoIDs)
	if err != nil {
		return ResultadoMiembros{}, err
	}

	return ResultadoMiembros{Agregados: agregados}, nil
}

func (s *Servicio) QuitarContacto(ctx context.Context, listaID uuid.UUID, contactoID uuid.UUID) error {
	return s.repositorio.QuitarContacto(ctx, listaID, contactoID)
}

func (s *Servicio) ListarMiembros(ctx context.Context, listaID uuid.UUID, limite int32, desfase int32) (ListadoMiembros, error) {
	existe, err := s.repositorio.Existe(ctx, listaID)
	if err != nil {
		return ListadoMiembros{}, err
	}
	if !existe {
		return ListadoMiembros{}, ErrNoEncontrada
	}

	limite, desfase = ajustarPaginado(limite, desfase)
	return s.repositorio.ListarMiembros(ctx, listaID, limite, desfase)
}

func ajustarPaginado(limite int32, desfase int32) (int32, int32) {
	if limite <= 0 {
		limite = limitePorDefecto
	}
	if limite > limiteMaximo {
		limite = limiteMaximo
	}
	if desfase < 0 {
		desfase = 0
	}
	return limite, desfase
}
