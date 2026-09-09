import { peticion } from "./api-client";
import type {
  Contacto,
  Correspondencia,
  Envio,
  Lista,
  Paginado,
  Previsualizacion,
  ResultadoEncolado,
  ResumenEnvios,
  ResumenImportacion,
  Sesion,
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

  crearContacto(datos: {
    nombre: string;
    empresa: string;
    correo: string;
    pais?: string | null;
  }): Promise<Contacto> {
    return peticion<Contacto>("/contactos", { metodo: "POST", cuerpo: datos });
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

  previsualizar(id: string): Promise<Previsualizacion> {
    return peticion<Previsualizacion>(`/correspondencia/${id}/previsualizacion`);
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
