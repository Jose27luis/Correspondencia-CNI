package correspondence

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/auth"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/mailer"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/storage"
)

const (
	limitePorDefecto     = 50
	limiteMaximo         = 200
	tamanoMaximoAdjuntos = 25 << 20
)

type Servicio struct {
	repositorio *Repositorio
	almacen     *storage.Almacen
	proveedor   mailer.Proveedor
	validador   *validator.Validate
}

func NuevoServicio(repositorio *Repositorio, almacen *storage.Almacen, proveedor mailer.Proveedor) *Servicio {
	return &Servicio{
		repositorio: repositorio,
		almacen:     almacen,
		proveedor:   proveedor,
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

func (s *Servicio) SubirAdjunto(ctx context.Context, correspondenciaID uuid.UUID, archivo multipart.File, cabecera *multipart.FileHeader) (Adjunto, error) {
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

	if acumulado+cabecera.Size > tamanoMaximoAdjuntos {
		return Adjunto{}, ErrLimiteAdjuntos
	}

	guardado, err := s.almacen.Guardar(archivo, cabecera)
	if err != nil {
		return Adjunto{}, err
	}

	adjunto, err := s.repositorio.CrearAdjunto(ctx, correspondenciaID, EntradaAdjunto{
		NombreArchivo: guardado.NombreArchivo,
		UrlArchivo:    guardado.UrlArchivo,
		Tipo:          guardado.Tipo,
		TamanoBytes:   guardado.TamanoBytes,
	})
	if err != nil {
		if errEliminar := s.almacen.Eliminar(guardado.UrlArchivo); errEliminar != nil {
			slog.Error("quedó un archivo huérfano en el almacén", "url", guardado.UrlArchivo, "error", errEliminar)
		}
		return Adjunto{}, err
	}

	return adjunto, nil
}

func (s *Servicio) SubirPlantilla(ctx context.Context, correspondenciaID uuid.UUID, archivo multipart.File, cabecera *multipart.FileHeader) (Correspondencia, error) {
	pieza, err := s.repositorio.Obtener(ctx, correspondenciaID)
	if err != nil {
		return Correspondencia{}, err
	}

	if pieza.Estado != EstadoBorrador {
		return Correspondencia{}, ErrNoEsBorrador
	}

	contenido, err := io.ReadAll(io.LimitReader(archivo, tamanoMaximoWord))
	if err != nil {
		return Correspondencia{}, fmt.Errorf("no se pudo leer el documento: %w", err)
	}

	if _, err := LeerWord(bytes.NewReader(contenido), int64(len(contenido))); err != nil {
		return Correspondencia{}, err
	}

	if _, err := archivo.Seek(0, io.SeekStart); err != nil {
		return Correspondencia{}, fmt.Errorf("no se pudo procesar el documento: %w", err)
	}

	guardado, err := s.almacen.Guardar(archivo, cabecera)
	if err != nil {
		return Correspondencia{}, err
	}

	actualizada, err := s.repositorio.AsignarPlantilla(
		ctx,
		correspondenciaID,
		&guardado.UrlArchivo,
		&guardado.NombreArchivo,
	)
	if err != nil {
		if errEliminar := s.almacen.Eliminar(guardado.UrlArchivo); errEliminar != nil {
			slog.Error("quedó una plantilla huérfana", "url", guardado.UrlArchivo, "error", errEliminar)
		}
		return Correspondencia{}, err
	}

	if pieza.PlantillaURL != nil {
		if err := s.almacen.Eliminar(*pieza.PlantillaURL); err != nil {
			slog.Error("no se pudo borrar la plantilla anterior", "error", err)
		}
	}

	return actualizada, nil
}

func (s *Servicio) QuitarPlantilla(ctx context.Context, correspondenciaID uuid.UUID) (Correspondencia, error) {
	pieza, err := s.repositorio.Obtener(ctx, correspondenciaID)
	if err != nil {
		return Correspondencia{}, err
	}

	if pieza.Estado != EstadoBorrador {
		return Correspondencia{}, ErrNoEsBorrador
	}

	actualizada, err := s.repositorio.AsignarPlantilla(ctx, correspondenciaID, nil, nil)
	if err != nil {
		return Correspondencia{}, err
	}

	if pieza.PlantillaURL != nil {
		if err := s.almacen.Eliminar(*pieza.PlantillaURL); err != nil {
			slog.Error("no se pudo borrar la plantilla", "error", err)
		}
	}

	return actualizada, nil
}

func (s *Servicio) EliminarAdjunto(ctx context.Context, correspondenciaID uuid.UUID, adjuntoID uuid.UUID) error {
	correspondencia, err := s.repositorio.Obtener(ctx, correspondenciaID)
	if err != nil {
		return err
	}

	if correspondencia.Estado != EstadoBorrador {
		return ErrNoEsBorrador
	}

	var urlArchivo string
	for _, adjunto := range correspondencia.Adjuntos {
		if adjunto.ID == adjuntoID {
			urlArchivo = adjunto.UrlArchivo
			break
		}
	}

	if err := s.repositorio.EliminarAdjunto(ctx, correspondenciaID, adjuntoID); err != nil {
		return err
	}

	if urlArchivo != "" {
		if err := s.almacen.Eliminar(urlArchivo); err != nil {
			slog.Error("no se pudo borrar el archivo del almacén", "url", urlArchivo, "error", err)
		}
	}

	return nil
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
