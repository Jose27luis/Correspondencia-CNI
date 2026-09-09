export type EstadoCorrespondencia =
  | "borrador"
  | "encolada"
  | "enviando"
  | "enviada"
  | "fallida";

export type EstadoEnvio =
  | "pendiente"
  | "enviado"
  | "entregado"
  | "rebotado"
  | "fallido";

export interface Usuario {
  id: string;
  nombre: string;
  correo: string;
  creado_en: string;
}

export interface Sesion {
  token: string;
  expira_en: string;
  usuario: Usuario;
}

export interface Contacto {
  id: string;
  nombre: string;
  empresa: string;
  correo: string;
  pais: string | null;
  campos_extra: Record<string, unknown>;
  creado_en: string;
  actualizado_en: string;
}

export interface Lista {
  id: string;
  nombre: string;
  descripcion: string | null;
  total_contactos: number;
  creado_en: string;
}

export interface Adjunto {
  id: string;
  nombre_archivo: string;
  url_archivo: string;
  tipo: string;
  tamano_bytes: number;
  creado_en: string;
}

export interface Correspondencia {
  id: string;
  usuario_id: string;
  lista_id: string | null;
  asunto: string;
  cuerpo: string;
  estado: EstadoCorrespondencia;
  plantilla_url: string | null;
  plantilla_nombre: string | null;
  adjuntos: Adjunto[];
  creado_en: string;
  actualizado_en: string;
}

export interface ContenidoWord {
  cuerpo: string;
  variables: string[];
}

export interface Previsualizacion {
  asunto: string;
  cuerpo: string;
  destinatario: string;
  variables_sin_valor: string[];
}

export interface Supresion {
  correo: string;
  motivo: "rebote" | "baja" | "manual";
  detalle: string | null;
  creado_en: string;
}

export interface ResultadoPrueba {
  destinatario: string;
  asunto: string;
  simulada: boolean;
  variables_sin_valor: string[];
}

export interface EstadoAsistente {
  disponible: boolean;
}

export interface Borrador {
  asunto: string;
  cuerpo: string;
}

export interface Hallazgo {
  gravedad: "alta" | "media" | "baja";
  titulo: string;
  detalle: string;
  sugerencia: string;
}

export interface Revision {
  veredicto: "lista" | "revisar" | "no_enviar";
  resumen: string;
  hallazgos: Hallazgo[];
}

export interface Envio {
  id: string;
  estado: EstadoEnvio;
  proveedor_mensaje_id: string | null;
  fecha_envio: string | null;
  nombre: string;
  empresa: string;
  correo: string;
}

export interface ResumenEnvios {
  total: number;
  por_estado: Partial<Record<EstadoEnvio, number>>;
}

export interface ResumenImportacion {
  creados: number;
  actualizados: number;
  omitidos: number;
  errores: string[];
}

export interface ResultadoEncolado {
  correspondencia_id: string;
  encolados: number;
  estado: EstadoCorrespondencia;
}

export interface Paginado<T> {
  datos: T[];
  total: number;
}

export interface ErrorApi {
  error: string;
  detalle?: Record<string, string>;
}
