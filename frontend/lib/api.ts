import { descargar, peticion } from "./api-client";
import type {
  Adjunto,
  Borrador,
  Contacto,
  ContenidoWord,
  CriteriosBusqueda,
  EntradaContacto,
  Correspondencia,
  Envio,
  EstadoAsistente,
  Extraccion,
  Lista,
  Paginado,
  Previsualizacion,
  ResultadoEncolado,
  ResumenEnvios,
  ResultadoPrueba,
  ResumenImportacion,
  Revision,
  Sesion,
  Sugerencia,
  Supresion,
  Usuario,
} from "./tipos";

export const api = {
  acceder(correo: string, contrasena: string): Promise<Sesion> {
    return peticion<Sesion>("/auth/acceso", {
      metodo: "POST",
      cuerpo: { correo, contrasena },
    });
  },

  perfil(): Promise<Usuario> {
    return peticion<Usuario>("/auth/perfil");
  },

  listarContactos(busqueda?: string, limite = 50, desfase = 0): Promise<Paginado<Contacto>> {
    return peticion<Paginado<Contacto>>("/contactos", {
      parametros: { busqueda, limite, desfase },
    });
  },

  crearContacto(datos: EntradaContacto): Promise<Contacto> {
    return peticion<Contacto>("/contactos", { metodo: "POST", cuerpo: datos });
  },

  actualizarContacto(id: string, datos: EntradaContacto): Promise<Contacto> {
    return peticion<Contacto>(`/contactos/${id}`, { metodo: "PUT", cuerpo: datos });
  },

  eliminarContacto(id: string): Promise<void> {
    return peticion<void>(`/contactos/${id}`, { metodo: "DELETE" });
  },

  importarContactos(archivo: File): Promise<ResumenImportacion> {
    const formulario = new FormData();
    formulario.append("archivo", archivo);
    return peticion<ResumenImportacion>("/contactos/importar", {
      metodo: "POST",
      formulario,
    });
  },

  descargarPlantilla(): Promise<void> {
    return descargar("/contactos/plantilla", "plantilla-contactos.xlsx");
  },

  listarListas(): Promise<Paginado<Lista>> {
    return peticion<Paginado<Lista>>("/listas");
  },

  crearLista(datos: { nombre: string; descripcion?: string | null }): Promise<Lista> {
    return peticion<Lista>("/listas", { metodo: "POST", cuerpo: datos });
  },

  eliminarLista(id: string): Promise<void> {
    return peticion<void>(`/listas/${id}`, { metodo: "DELETE" });
  },

  miembrosDeLista(id: string): Promise<Paginado<Contacto>> {
    return peticion<Paginado<Contacto>>(`/listas/${id}/contactos`);
  },

  agregarALista(id: string, contactoIds: string[]): Promise<{ agregados: number }> {
    return peticion<{ agregados: number }>(`/listas/${id}/contactos`, {
      metodo: "POST",
      cuerpo: { contacto_ids: contactoIds },
    });
  },

  quitarDeLista(listaId: string, contactoId: string): Promise<void> {
    return peticion<void>(`/listas/${listaId}/contactos/${contactoId}`, {
      metodo: "DELETE",
    });
  },

  listarCorrespondencia(estado?: string): Promise<Paginado<Correspondencia>> {
    return peticion<Paginado<Correspondencia>>("/correspondencia", {
      parametros: { estado },
    });
  },

  obtenerCorrespondencia(id: string): Promise<Correspondencia> {
    return peticion<Correspondencia>(`/correspondencia/${id}`);
  },

  sugerirCuerpo(datos: {
    documento: string;
    asunto?: string;
    variables?: string[];
  }): Promise<Sugerencia> {
    return peticion<Sugerencia>("/asistente/sugerir-cuerpo", { metodo: "POST", cuerpo: datos });
  },

  estadoAsistente(): Promise<EstadoAsistente> {
    return peticion<EstadoAsistente>("/asistente/estado");
  },

  redactarConAsistente(datos: {
    instruccion: string;
    asunto?: string;
    cuerpo?: string;
    variables?: string[];
  }): Promise<Borrador> {
    return peticion<Borrador>("/asistente/redactar", { metodo: "POST", cuerpo: datos });
  },

  revisarConAsistente(datos: {
    asunto: string;
    cuerpo: string;
    variables?: string[];
  }): Promise<Revision> {
    return peticion<Revision>("/asistente/revisar", { metodo: "POST", cuerpo: datos });
  },

  extraerContactos(texto: string): Promise<Extraccion> {
    return peticion<Extraccion>("/asistente/contactos/extraer", {
      metodo: "POST",
      cuerpo: { texto },
    });
  },

  interpretarBusqueda(consulta: string): Promise<CriteriosBusqueda> {
    return peticion<CriteriosBusqueda>("/asistente/contactos/buscar", {
      metodo: "POST",
      cuerpo: { consulta },
    });
  },

  leerDocumento(archivo: File): Promise<ContenidoWord> {
    const formulario = new FormData();
    formulario.append("archivo", archivo);
    return peticion<ContenidoWord>("/correspondencia/leer-documento", {
      metodo: "POST",
      formulario,
    });
  },

  crearCorrespondencia(datos: {
    lista_id?: string | null;
    asunto: string;
    cuerpo: string;
  }): Promise<Correspondencia> {
    return peticion<Correspondencia>("/correspondencia", {
      metodo: "POST",
      cuerpo: datos,
    });
  },

  actualizarCorrespondencia(
    id: string,
    datos: { lista_id?: string | null; asunto: string; cuerpo: string },
  ): Promise<Correspondencia> {
    return peticion<Correspondencia>(`/correspondencia/${id}`, {
      metodo: "PUT",
      cuerpo: datos,
    });
  },

  eliminarCorrespondencia(id: string): Promise<void> {
    return peticion<void>(`/correspondencia/${id}`, { metodo: "DELETE" });
  },

  subirPlantilla(id: string, archivo: File): Promise<Correspondencia> {
    const formulario = new FormData();
    formulario.append("archivo", archivo);
    return peticion<Correspondencia>(`/correspondencia/${id}/plantilla`, {
      metodo: "POST",
      formulario,
    });
  },

  quitarPlantilla(id: string): Promise<Correspondencia> {
    return peticion<Correspondencia>(`/correspondencia/${id}/plantilla`, {
      metodo: "DELETE",
    });
  },

  subirAdjunto(id: string, archivo: File): Promise<Adjunto> {
    const formulario = new FormData();
    formulario.append("archivo", archivo);
    return peticion<Adjunto>(`/correspondencia/${id}/adjuntos`, {
      metodo: "POST",
      formulario,
    });
  },

  eliminarAdjunto(id: string, adjuntoId: string): Promise<void> {
    return peticion<void>(`/correspondencia/${id}/adjuntos/${adjuntoId}`, {
      metodo: "DELETE",
    });
  },

  previsualizar(id: string): Promise<Previsualizacion> {
    return peticion<Previsualizacion>(`/correspondencia/${id}/previsualizacion`);
  },

  enviarPrueba(id: string): Promise<ResultadoPrueba> {
    return peticion<ResultadoPrueba>(`/correspondencia/${id}/prueba`, { metodo: "POST" });
  },

  listarExclusiones(): Promise<Paginado<Supresion>> {
    return peticion<Paginado<Supresion>>("/exclusiones");
  },

  excluirCorreo(correo: string, detalle: string): Promise<void> {
    return peticion<void>("/exclusiones", { metodo: "POST", cuerpo: { correo, detalle } });
  },

  quitarExclusion(correo: string): Promise<void> {
    return peticion<void>(`/exclusiones/${encodeURIComponent(correo)}`, { metodo: "DELETE" });
  },

  enviar(id: string): Promise<ResultadoEncolado> {
    return peticion<ResultadoEncolado>(`/correspondencia/${id}/enviar`, {
      metodo: "POST",
    });
  },

  listarEnvios(id: string, estado?: string): Promise<Paginado<Envio>> {
    return peticion<Paginado<Envio>>(`/correspondencia/${id}/envios`, {
      parametros: { estado },
    });
  },

  resumenEnvios(id: string): Promise<ResumenEnvios> {
    return peticion<ResumenEnvios>(`/correspondencia/${id}/envios/resumen`);
  },
};
