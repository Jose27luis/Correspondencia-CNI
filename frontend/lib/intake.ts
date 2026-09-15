import { peticion } from "./api-client";
import type { Paginado } from "./tipos";

export type EstadoRevision = "pendiente" | "aprobado" | "rechazado";
export type TipoPublicacion = "oferta" | "demanda";
export type EstadoCoincidencia = "sugerida" | "notificada" | "descartada";

export interface EmpresaDirectorio {
  id: string;
  nombre: string;
  ruc: string;
  correo: string;
  direccion: string;
  ciudad: string;
  telefono: string | null;
  celular: string | null;
  facebook: string | null;
  pagina_web: string | null;
  descripcion: string;
  logo_url: string | null;
  estado: EstadoRevision;
  motivo_rechazo: string | null;
  creado_en: string;
}

export type EmpresaDirectorioPublica = Omit<
  EmpresaDirectorio,
  "ruc" | "estado" | "motivo_rechazo" | "creado_en"
>;

export interface EmpresaRueda {
  id: string;
  razon_social: string;
  ruc: string;
  persona_encargada: string;
  cargo_encargado: string | null;
  correo: string;
  ciudad: string | null;
  region: string | null;
  pais: string;
  telefono: string | null;
  celular: string | null;
  pagina_web: string | null;
  estado: EstadoRevision;
  motivo_rechazo: string | null;
  creado_en: string;
  publicaciones: number;
}

export interface Publicacion {
  id: string;
  empresa_id: string;
  tipo: TipoPublicacion;
  titulo: string;
  descripcion: string;
  imagen_url: string | null;
  estado: EstadoRevision;
  motivo_rechazo: string | null;
  creado_en: string;
  razon_social: string;
  empresa_ciudad: string | null;
  empresa_pais: string;
  empresa_correo: string;
  empresa_estado: EstadoRevision;
}

export interface PublicacionPublica {
  id: string;
  tipo: TipoPublicacion;
  titulo: string;
  descripcion: string;
  imagen_url: string | null;
  creado_en: string;
  razon_social: string;
  empresa_ciudad: string | null;
  empresa_pais: string;
}

export interface Coincidencia {
  id: string;
  puntaje: number;
  motivo: string;
  estado: EstadoCoincidencia;
  creado_en: string;
  notificada_en: string | null;
  demanda_titulo: string;
  demanda_empresa: string;
  demanda_correo: string;
  oferta_titulo: string;
  oferta_empresa: string;
  oferta_correo: string;
}

interface Registro {
  recibido: boolean;
}

export const etiquetaRevision: Record<EstadoRevision, string> = {
  pendiente: "Pendiente",
  aprobado: "Aprobado",
  rechazado: "Rechazado",
};

export const colorRevision: Record<EstadoRevision, string> = {
  pendiente: "bg-amber-100 text-amber-800 hover:bg-amber-100",
  aprobado: "bg-emerald-100 text-emerald-800 hover:bg-emerald-100",
  rechazado: "bg-red-100 text-red-800 hover:bg-red-100",
};

export const etiquetaTipo: Record<TipoPublicacion, string> = {
  oferta: "Oferta",
  demanda: "Demanda",
};

export const campoTrampa = "confirmacion_correo";

export const apiCaptacion = {
  directorioPublico(): Promise<EmpresaDirectorioPublica[]> {
    return peticion<EmpresaDirectorioPublica[]>("/publico/directorio");
  },

  registrarEnDirectorio(formulario: FormData): Promise<Registro> {
    return peticion<Registro>("/publico/directorio", { metodo: "POST", formulario });
  },

  ruedaPublica(tipo?: TipoPublicacion): Promise<PublicacionPublica[]> {
    return peticion<PublicacionPublica[]>("/publico/rueda", { parametros: { tipo } });
  },

  registrarEnRueda(formulario: FormData): Promise<Registro> {
    return peticion<Registro>("/publico/rueda", { metodo: "POST", formulario });
  },

  listarDirectorio(estado?: EstadoRevision): Promise<Paginado<EmpresaDirectorio>> {
    return peticion<Paginado<EmpresaDirectorio>>("/directorio", {
      parametros: { estado, limite: 200 },
    });
  },

  aprobarDirectorio(id: string): Promise<EmpresaDirectorio> {
    return peticion<EmpresaDirectorio>(`/directorio/${id}/aprobar`, { metodo: "POST" });
  },

  rechazarDirectorio(id: string, motivo: string): Promise<EmpresaDirectorio> {
    return peticion<EmpresaDirectorio>(`/directorio/${id}/rechazar`, {
      metodo: "POST",
      cuerpo: { motivo },
    });
  },

  listarEmpresasRueda(estado?: EstadoRevision): Promise<Paginado<EmpresaRueda>> {
    return peticion<Paginado<EmpresaRueda>>("/rueda/empresas", {
      parametros: { estado, limite: 200 },
    });
  },

  aprobarEmpresaRueda(id: string): Promise<EmpresaRueda> {
    return peticion<EmpresaRueda>(`/rueda/empresas/${id}/aprobar`, { metodo: "POST" });
  },

  rechazarEmpresaRueda(id: string, motivo: string): Promise<EmpresaRueda> {
    return peticion<EmpresaRueda>(`/rueda/empresas/${id}/rechazar`, {
      metodo: "POST",
      cuerpo: { motivo },
    });
  },

  listarPublicaciones(estado?: EstadoRevision): Promise<Paginado<Publicacion>> {
    return peticion<Paginado<Publicacion>>("/rueda/publicaciones", {
      parametros: { estado, limite: 200 },
    });
  },

  aprobarPublicacion(id: string): Promise<Publicacion> {
    return peticion<Publicacion>(`/rueda/publicaciones/${id}/aprobar`, { metodo: "POST" });
  },

  rechazarPublicacion(id: string, motivo: string): Promise<Publicacion> {
    return peticion<Publicacion>(`/rueda/publicaciones/${id}/rechazar`, {
      metodo: "POST",
      cuerpo: { motivo },
    });
  },

  buscarCoincidencias(id: string): Promise<{ encontradas: number }> {
    return peticion<{ encontradas: number }>(`/rueda/publicaciones/${id}/coincidencias`, {
      metodo: "POST",
    });
  },

  listarCoincidencias(estado?: EstadoCoincidencia): Promise<Coincidencia[]> {
    return peticion<Coincidencia[]>("/rueda/coincidencias", { parametros: { estado } });
  },

  notificarCoincidencia(id: string): Promise<void> {
    return peticion<void>(`/rueda/coincidencias/${id}/notificar`, { metodo: "POST" });
  },

  descartarCoincidencia(id: string): Promise<void> {
    return peticion<void>(`/rueda/coincidencias/${id}/descartar`, { metodo: "POST" });
  },
};
