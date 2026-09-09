package correspondence

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/auth"
)

const (
	limitePorDefecto     = 50
	limiteMaximo         = 200
	tamanoMaximoAdjuntos = 25 << 20
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

func (s *Servicio) Crear(ctx context.Context, entrada EntradaCorrespondencia) (Correspondencia, error) {
	identidad, autenticado := auth.DesdeContexto(ctx)
	if !autenticado {
		return Correspondencia{}, auth.ErrSinToken
	}

	entrada.Normalizar()
	if err := s.validador.Struct(entrada); err != nil {
		return Correspondencia{}, err
	}

	return s.repositorio.Crear(ctx, identidad.UsuarioID, entrada)
}

func (s *Servicio) Obtener(ctx context.Context, id uuid.UUID) (Correspondencia, error) {
	return s.repositorio.Obtener(ctx, id)
}

func (s *Servicio) Listar(ctx context.Context, estado *string, limite int32, desfase int32) (ListadoCorrespondencia, error) {
	if limite <= 0 {
		limite = limitePorDefecto
	}
	if limite > limiteMaximo {
		limite = limiteMaximo
	}
	if desfase < 0 {
		desfase = 0
	}
	return s.repositorio.Listar(ctx, estado, limite, desfase)
}

func (s *Servicio) Actualizar(ctx context.Context, id uuid.UUID, entrada EntradaCorrespondencia) (Correspondencia, error) {
	entrada.Normalizar()
	if err := s.validador.Struct(entrada); err != nil {
		return Correspondencia{}, err
	}
	return s.repositorio.Actualizar(ctx, id, entrada)
}

func (s *Servicio) Eliminar(ctx context.Context, id uuid.UUID) error {
	return s.repositorio.Eliminar(ctx, id)
}

func (s *Servicio) AgregarAdjunto(ctx context.Context, correspondenciaID uuid.UUID, entrada EntradaAdjunto) (Adjunto, error) {
	entrada.Normalizar()
	if err := s.validador.Struct(entrada); err != nil {
		return Adjunto{}, err
	}

	correspondencia, err := s.repositorio.Obtener(ctx, correspondenciaID)
	if err != nil {
		return Adjunto{}, err
	}

	if correspondencia.Estado != EstadoBorrador {
		return Adjunto{}, ErrNoEsBorrador
	}

	acumulado, err := s.repositorio.TamanoAdjuntos(ctx, correspondenciaID)
	if err != nil {
		return Adjunto{}, err
	}

	if acumulado+entrada.TamanoBytes > tamanoMaximoAdjuntos {
		return Adjunto{}, ErrLimiteAdjuntos
	}

	return s.repositorio.CrearAdjunto(ctx, correspondenciaID, entrada)
}

func (s *Servicio) EliminarAdjunto(ctx context.Context, correspondenciaID uuid.UUID, adjuntoID uuid.UUID) error {
	correspondencia, err := s.repositorio.Obtener(ctx, correspondenciaID)
	if err != nil {
		return err
	}

	if correspondencia.Estado != EstadoBorrador {
		return ErrNoEsBorrador
	}

	return s.repositorio.EliminarAdjunto(ctx, correspondenciaID, adjuntoID)
}

func (s *Servicio) Previsualizar(ctx context.Context, id uuid.UUID) (Previsualizacion, error) {
	correspondencia, err := s.repositorio.Obtener(ctx, id)
	if err != nil {
		return Previsualizacion{}, err
	}

	if correspondencia.ListaID == nil {
		return Previsualizacion{}, ErrSinDestinatario
	}

	contacto, err := s.repositorio.PrimerContactoDeLista(ctx, *correspondencia.ListaID)
	if err != nil {
		return Previsualizacion{}, err
	}

	asunto := Renderizar(correspondencia.Asunto, contacto)
	cuerpo := Renderizar(correspondencia.Cuerpo, contacto)

	sinValor := unir(asunto.SinValor, cuerpo.SinValor)

	return Previsualizacion{
		Asunto:       asunto.Texto,
		Cuerpo:       cuerpo.Texto,
		Destinatario: contacto.Correo,
		Variables:    sinValor,
	}, nil
}

func unir(primera []string, segunda []string) []string {
	unicas := make(map[string]struct{}, len(primera)+len(segunda))
	resultado := make([]string, 0, len(primera)+len(segunda))

	for _, grupo := range [][]string{primera, segunda} {
		for _, nombre := range grupo {
			if _, repetida := unicas[nombre]; repetida {
				continue
			}
			unicas[nombre] = struct{}{}
			resultado = append(resultado, nombre)
		}
	}

	return resultado
}
