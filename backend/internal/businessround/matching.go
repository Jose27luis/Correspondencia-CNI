package businessround

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/ai"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/mailer"
)

const tiempoBusqueda = 90 * time.Second

func (s *Servicio) BuscarEnSegundoPlano(id uuid.UUID) {
	if !s.asistente.Disponible() {
		return
	}

	go func() {
		ctx, cancelar := context.WithTimeout(context.Background(), tiempoBusqueda)
		defer cancelar()

		if _, err := s.BuscarCoincidencias(ctx, id); err != nil {
			slog.Warn("no se pudieron buscar coincidencias", "publicacion", id, "error", err)
		}
	}()
}

func (s *Servicio) BuscarCoincidencias(ctx context.Context, id uuid.UUID) (ResultadoBusqueda, error) {
	if !s.asistente.Disponible() {
		return ResultadoBusqueda{}, ai.ErrNoConfigurado
	}

	referencia, err := s.consultas.ObtenerPublicacion(ctx, id)
	if err != nil {
		return ResultadoBusqueda{}, traducir(err)
	}
	if referencia.Estado != EstadoAprobado || referencia.EmpresaEstado != EstadoAprobado {
		return ResultadoBusqueda{}, ErrNoAprobada
	}

	tipoContrario := TipoOferta
	if referencia.Tipo == TipoOferta {
		tipoContrario = TipoDemanda
	}

	filas, err := s.consultas.ListarCandidatas(ctx, sqlcgen.ListarCandidatasParams{
		Tipo:      tipoContrario,
		EmpresaID: referencia.EmpresaID,
	})
	if err != nil {
		return ResultadoBusqueda{}, fmt.Errorf("no se pudieron listar las candidatas: %w", err)
	}

	candidatas := make([]ai.PublicacionResumen, 0, len(filas))
	for _, fila := range filas {
		candidatas = append(candidatas, ai.PublicacionResumen{
			ID:          fila.ID.String(),
			Titulo:      fila.Titulo,
			Descripcion: fila.Descripcion,
			Empresa:     fila.RazonSocial,
			Ciudad:      valor(fila.EmpresaCiudad),
			Pais:        fila.EmpresaPais,
		})
	}

	sugeridas, err := s.asistente.SugerirCoincidencias(ctx, ai.PublicacionResumen{
		ID:          referencia.ID.String(),
		Titulo:      referencia.Titulo,
		Descripcion: referencia.Descripcion,
		Empresa:     referencia.RazonSocial,
		Ciudad:      valor(referencia.EmpresaCiudad),
		Pais:        referencia.EmpresaPais,
	}, referencia.Tipo, candidatas)
	if err != nil {
		return ResultadoBusqueda{}, err
	}

	for _, sugerida := range sugeridas {
		candidataID, err := uuid.Parse(sugerida.ID)
		if err != nil {
			continue
		}

		parametros := sqlcgen.GuardarCoincidenciaParams{
			DemandaID: referencia.ID,
			OfertaID:  candidataID,
			Puntaje:   int32(sugerida.Puntaje),
			Motivo:    sugerida.Motivo,
		}
		if referencia.Tipo == TipoOferta {
			parametros.DemandaID, parametros.OfertaID = candidataID, referencia.ID
		}

		if err := s.consultas.GuardarCoincidencia(ctx, parametros); err != nil {
			return ResultadoBusqueda{}, fmt.Errorf("no se pudo guardar la coincidencia: %w", err)
		}
	}

	return ResultadoBusqueda{Encontradas: len(sugeridas)}, nil
}

func (s *Servicio) ListarCoincidencias(ctx context.Context, estado *string) ([]Coincidencia, error) {
	filas, err := s.consultas.ListarCoincidencias(ctx, estado)
	if err != nil {
		return nil, fmt.Errorf("no se pudieron listar las coincidencias: %w", err)
	}

	datos := make([]Coincidencia, 0, len(filas))
	for _, fila := range filas {
		datos = append(datos, Coincidencia(fila))
	}
	return datos, nil
}

func (s *Servicio) Descartar(ctx context.Context, id uuid.UUID) error {
	if _, err := s.consultas.MarcarCoincidencia(ctx, sqlcgen.MarcarCoincidenciaParams{
		Estado: CoincidenciaDescartada,
		ID:     id,
	}); err != nil {
		return traducir(err)
	}
	return nil
}

func (s *Servicio) Notificar(ctx context.Context, id uuid.UUID) error {
	coincidencia, err := s.consultas.ObtenerCoincidencia(ctx, id)
	if err != nil {
		return traducir(err)
	}
	if coincidencia.Estado == CoincidenciaNotificada {
		return ErrYaNotificada
	}

	mensajes := []mailer.Mensaje{
		{
			Para:   coincidencia.DemandaCorreo,
			Asunto: fmt.Sprintf("Rueda de Negocios CNI: %s puede atender su requerimiento", coincidencia.OfertaEmpresa),
			Cuerpo: presentacion(
				coincidencia.DemandaEncargado,
				coincidencia.DemandaTitulo,
				coincidencia.OfertaEmpresa,
				coincidencia.OfertaTitulo,
				coincidencia.OfertaDescripcion,
				coincidencia.OfertaCorreo,
				coincidencia.Motivo,
			),
		},
		{
			Para:   coincidencia.OfertaCorreo,
			Asunto: fmt.Sprintf("Rueda de Negocios CNI: %s busca lo que usted ofrece", coincidencia.DemandaEmpresa),
			Cuerpo: presentacion(
				coincidencia.OfertaEncargado,
				coincidencia.OfertaTitulo,
				coincidencia.DemandaEmpresa,
				coincidencia.DemandaTitulo,
				coincidencia.DemandaDescripcion,
				coincidencia.DemandaCorreo,
				coincidencia.Motivo,
			),
		},
	}

	for _, mensaje := range mensajes {
		if _, err := s.proveedor.Enviar(ctx, mensaje); err != nil {
			return fmt.Errorf("no se pudo notificar a %s: %w", mensaje.Para, err)
		}
	}

	if _, err := s.consultas.MarcarCoincidencia(ctx, sqlcgen.MarcarCoincidenciaParams{
		Estado: CoincidenciaNotificada,
		ID:     id,
	}); err != nil {
		return fmt.Errorf("no se pudo marcar la coincidencia: %w", err)
	}
	return nil
}

func presentacion(
	encargado string,
	publicacionPropia string,
	otraEmpresa string,
	otraPublicacion string,
	otraDescripcion string,
	otroCorreo string,
	motivo string,
) string {
	var cuerpo strings.Builder
	fmt.Fprintf(&cuerpo, "Estimado(a) %s:\n\n", encargado)
	fmt.Fprintf(&cuerpo, "En la Rueda de Negocios de CNI identificamos una oportunidad para su publicación \"%s\".\n\n", publicacionPropia)
	fmt.Fprintf(&cuerpo, "%s publicó: \"%s\".\n%s\n\n", otraEmpresa, otraPublicacion, otraDescripcion)
	fmt.Fprintf(&cuerpo, "Por qué creemos que encajan: %s\n\n", motivo)
	fmt.Fprintf(&cuerpo, "Puede escribirles directamente a %s.\n\n", otroCorreo)
	cuerpo.WriteString("Atentamente,\nCorporación de Negocios Interoceánicos")
	return cuerpo.String()
}
