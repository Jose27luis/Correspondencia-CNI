package businessround

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/intake"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/ai"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/mailer"
)

const (
	limitePorDefecto = 50
	limiteMaximo     = 200
)

type Servicio struct {
	consultas    *sqlcgen.Queries
	incorporador *intake.Incorporador
	asistente    *ai.Asistente
	proveedor    mailer.Proveedor
	validador    *validator.Validate
}

func NuevoServicio(
	pool *pgxpool.Pool,
	incorporador *intake.Incorporador,
	asistente *ai.Asistente,
	proveedor mailer.Proveedor,
) *Servicio {
	return &Servicio{
		consultas:    sqlcgen.New(pool),
		incorporador: incorporador,
		asistente:    asistente,
		proveedor:    proveedor,
		validador:    validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (s *Servicio) Validar(entrada *EntradaRegistro) error {
	entrada.Correo = strings.ToLower(strings.TrimSpace(entrada.Correo))
	return s.validador.Struct(entrada)
}

func (s *Servicio) Registrar(ctx context.Context, entrada EntradaRegistro) error {
	empresa, err := s.consultas.RegistrarRuedaEmpresa(ctx, sqlcgen.RegistrarRuedaEmpresaParams{
		RazonSocial:      entrada.RazonSocial,
		Ruc:              entrada.Ruc,
		PersonaEncargada: entrada.PersonaEncargada,
		CargoEncargado:   entrada.CargoEncargado,
		Correo:           entrada.Correo,
		Direccion:        entrada.Direccion,
		Ciudad:           entrada.Ciudad,
		Region:           entrada.Region,
		Pais:             entrada.Pais,
		CodigoPostal:     entrada.CodigoPostal,
		Telefono:         entrada.Telefono,
		Celular:          entrada.Celular,
		PaginaWeb:        entrada.PaginaWeb,
	})
	if err != nil {
		return fmt.Errorf("no se pudo registrar la empresa: %w", err)
	}

	if _, err := s.consultas.CrearPublicacion(ctx, sqlcgen.CrearPublicacionParams{
		EmpresaID:   empresa.ID,
		Tipo:        entrada.Tipo,
		Titulo:      entrada.Titulo,
		Descripcion: entrada.Descripcion,
		ImagenUrl:   entrada.ImagenUrl,
	}); err != nil {
		return fmt.Errorf("no se pudo registrar la publicación: %w", err)
	}
	return nil
}

func (s *Servicio) ListarPublicas(ctx context.Context, tipo *string) ([]PublicacionPublica, error) {
	if tipo != nil && *tipo != TipoOferta && *tipo != TipoDemanda {
		return nil, ErrTipoInvalido
	}

	filas, err := s.consultas.ListarPublicacionesPublicas(ctx, tipo)
	if err != nil {
		return nil, fmt.Errorf("no se pudieron listar las publicaciones: %w", err)
	}

	datos := make([]PublicacionPublica, 0, len(filas))
	for _, fila := range filas {
		datos = append(datos, PublicacionPublica{
			ID:            fila.ID,
			Tipo:          fila.Tipo,
			Titulo:        fila.Titulo,
			Descripcion:   fila.Descripcion,
			ImagenUrl:     fila.ImagenUrl,
			CreadoEn:      fila.CreadoEn,
			RazonSocial:   fila.RazonSocial,
			EmpresaCiudad: fila.EmpresaCiudad,
			EmpresaPais:   fila.EmpresaPais,
		})
	}
	return datos, nil
}

func (s *Servicio) ListarEmpresas(ctx context.Context, estado *string, limite int32, desfase int32) (ListadoEmpresas, error) {
	limite, desfase = acotar(limite, desfase)

	filas, err := s.consultas.ListarRuedaEmpresas(ctx, sqlcgen.ListarRuedaEmpresasParams{
		Limit:  limite,
		Offset: desfase,
		Estado: estado,
	})
	if err != nil {
		return ListadoEmpresas{}, fmt.Errorf("no se pudieron listar las empresas: %w", err)
	}

	total, err := s.consultas.ContarRuedaEmpresas(ctx, estado)
	if err != nil {
		return ListadoEmpresas{}, fmt.Errorf("no se pudieron contar las empresas: %w", err)
	}

	datos := make([]Empresa, 0, len(filas))
	for _, fila := range filas {
		datos = append(datos, empresaDesdeFila(sqlcgen.RuedaEmpresa{
			ID:               fila.ID,
			RazonSocial:      fila.RazonSocial,
			Ruc:              fila.Ruc,
			PersonaEncargada: fila.PersonaEncargada,
			CargoEncargado:   fila.CargoEncargado,
			Correo:           fila.Correo,
			Direccion:        fila.Direccion,
			Ciudad:           fila.Ciudad,
			Region:           fila.Region,
			Pais:             fila.Pais,
			CodigoPostal:     fila.CodigoPostal,
			Telefono:         fila.Telefono,
			Celular:          fila.Celular,
			PaginaWeb:        fila.PaginaWeb,
			Estado:           fila.Estado,
			MotivoRechazo:    fila.MotivoRechazo,
			ContactoID:       fila.ContactoID,
			CreadoEn:         fila.CreadoEn,
			RevisadoEn:       fila.RevisadoEn,
		}, fila.Publicaciones))
	}
	return ListadoEmpresas{Datos: datos, Total: total}, nil
}

func (s *Servicio) AprobarEmpresa(ctx context.Context, id uuid.UUID) (Empresa, error) {
	fila, err := s.consultas.ObtenerRuedaEmpresa(ctx, id)
	if err != nil {
		return Empresa{}, traducir(err)
	}
	if fila.Estado == EstadoAprobado {
		return Empresa{}, ErrYaRevisada
	}

	pais := fila.Pais
	contactoID, err := s.incorporador.Incorporar(ctx, intake.Solicitud{
		Nombre:  fila.PersonaEncargada,
		Empresa: fila.RazonSocial,
		Correo:  fila.Correo,
		Pais:    &pais,
		Datos: map[string]string{
			"ruc":           fila.Ruc,
			"cargo":         valor(fila.CargoEncargado),
			"direccion":     valor(fila.Direccion),
			"ciudad":        valor(fila.Ciudad),
			"region":        valor(fila.Region),
			"codigo_postal": valor(fila.CodigoPostal),
			"telefono":      valor(fila.Telefono),
			"celular":       valor(fila.Celular),
			"pagina_web":    valor(fila.PaginaWeb),
			"origen":        NombreLista,
		},
		NombreLista:      NombreLista,
		DescripcionLista: DescripcionLista,
	})
	if err != nil {
		return Empresa{}, err
	}

	aprobada, err := s.consultas.AprobarRuedaEmpresa(ctx, sqlcgen.AprobarRuedaEmpresaParams{
		ID:         id,
		ContactoID: &contactoID,
	})
	if err != nil {
		return Empresa{}, fmt.Errorf("no se pudo aprobar la empresa: %w", err)
	}
	return empresaDesdeFila(aprobada, 0), nil
}

func (s *Servicio) RechazarEmpresa(ctx context.Context, id uuid.UUID, entrada EntradaRechazo) (Empresa, error) {
	if err := s.validarRechazo(&entrada); err != nil {
		return Empresa{}, err
	}

	fila, err := s.consultas.ObtenerRuedaEmpresa(ctx, id)
	if err != nil {
		return Empresa{}, traducir(err)
	}
	if fila.Estado != EstadoPendiente {
		return Empresa{}, ErrYaRevisada
	}

	rechazada, err := s.consultas.RechazarRuedaEmpresa(ctx, sqlcgen.RechazarRuedaEmpresaParams{
		ID:            id,
		MotivoRechazo: &entrada.Motivo,
	})
	if err != nil {
		return Empresa{}, fmt.Errorf("no se pudo rechazar la empresa: %w", err)
	}
	return empresaDesdeFila(rechazada, 0), nil
}

func (s *Servicio) ListarPublicaciones(
	ctx context.Context,
	estado *string,
	tipo *string,
	limite int32,
	desfase int32,
) (ListadoPublicaciones, error) {
	limite, desfase = acotar(limite, desfase)

	filas, err := s.consultas.ListarPublicaciones(ctx, sqlcgen.ListarPublicacionesParams{
		Limit:  limite,
		Offset: desfase,
		Estado: estado,
		Tipo:   tipo,
	})
	if err != nil {
		return ListadoPublicaciones{}, fmt.Errorf("no se pudieron listar las publicaciones: %w", err)
	}

	total, err := s.consultas.ContarPublicaciones(ctx, sqlcgen.ContarPublicacionesParams{Estado: estado, Tipo: tipo})
	if err != nil {
		return ListadoPublicaciones{}, fmt.Errorf("no se pudieron contar las publicaciones: %w", err)
	}

	datos := make([]Publicacion, 0, len(filas))
	for _, fila := range filas {
		datos = append(datos, publicacionDesdeFila(sqlcgen.ObtenerPublicacionRow(fila)))
	}
	return ListadoPublicaciones{Datos: datos, Total: total}, nil
}

func (s *Servicio) AprobarPublicacion(ctx context.Context, id uuid.UUID) (Publicacion, error) {
	fila, err := s.consultas.ObtenerPublicacion(ctx, id)
	if err != nil {
		return Publicacion{}, traducir(err)
	}
	if fila.Estado == EstadoAprobado {
		return Publicacion{}, ErrYaRevisada
	}
	if fila.EmpresaEstado != EstadoAprobado {
		return Publicacion{}, ErrEmpresaPendiente
	}

	if _, err := s.consultas.AprobarPublicacion(ctx, id); err != nil {
		return Publicacion{}, fmt.Errorf("no se pudo aprobar la publicación: %w", err)
	}

	fila.Estado = EstadoAprobado
	return publicacionDesdeFila(fila), nil
}

func (s *Servicio) RechazarPublicacion(ctx context.Context, id uuid.UUID, entrada EntradaRechazo) (Publicacion, error) {
	if err := s.validarRechazo(&entrada); err != nil {
		return Publicacion{}, err
	}

	fila, err := s.consultas.ObtenerPublicacion(ctx, id)
	if err != nil {
		return Publicacion{}, traducir(err)
	}
	if fila.Estado != EstadoPendiente {
		return Publicacion{}, ErrYaRevisada
	}

	if _, err := s.consultas.RechazarPublicacion(ctx, sqlcgen.RechazarPublicacionParams{
		ID:            id,
		MotivoRechazo: &entrada.Motivo,
	}); err != nil {
		return Publicacion{}, fmt.Errorf("no se pudo rechazar la publicación: %w", err)
	}

	fila.Estado = EstadoRechazado
	fila.MotivoRechazo = &entrada.Motivo
	return publicacionDesdeFila(fila), nil
}

func (s *Servicio) validarRechazo(entrada *EntradaRechazo) error {
	entrada.Motivo = strings.TrimSpace(entrada.Motivo)
	return s.validador.Struct(entrada)
}

func acotar(limite int32, desfase int32) (int32, int32) {
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

func traducir(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoEncontrada
	}
	return fmt.Errorf("error de base de datos: %w", err)
}

func valor(texto *string) string {
	if texto == nil {
		return ""
	}
	return *texto
}
