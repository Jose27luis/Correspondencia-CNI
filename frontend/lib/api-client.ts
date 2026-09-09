import type { ErrorApi } from "./tipos";

const URL_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://127.0.0.1:8080/api";
const CLAVE_TOKEN = "correspondencia_cni_token";

export class ErrorPeticion extends Error {
  readonly codigo: number;
  readonly detalle: Record<string, string>;

  constructor(codigo: number, mensaje: string, detalle: Record<string, string> = {}) {
    super(mensaje);
    this.name = "ErrorPeticion";
    this.codigo = codigo;
    this.detalle = detalle;
  }

  get esNoAutorizado(): boolean {
    return this.codigo === 401;
  }
}

export function guardarToken(token: string): void {
  window.localStorage.setItem(CLAVE_TOKEN, token);
}

export function leerToken(): string | null {
  if (typeof window === "undefined") {
    return null;
  }
  return window.localStorage.getItem(CLAVE_TOKEN);
}

export function borrarToken(): void {
  window.localStorage.removeItem(CLAVE_TOKEN);
}

interface OpcionesPeticion {
  metodo?: "GET" | "POST" | "PUT" | "DELETE";
  cuerpo?: unknown;
  formulario?: FormData;
  parametros?: Record<string, string | number | undefined>;
}

export async function peticion<T>(ruta: string, opciones: OpcionesPeticion = {}): Promise<T> {
  const { metodo = "GET", cuerpo, formulario, parametros } = opciones;

  const url = new URL(`${URL_BASE}${ruta}`);
  if (parametros) {
    for (const [clave, valor] of Object.entries(parametros)) {
      if (valor !== undefined && valor !== "") {
        url.searchParams.set(clave, String(valor));
      }
    }
  }

  const cabeceras: Record<string, string> = {};
  const token = leerToken();
  if (token) {
    cabeceras.Authorization = `Bearer ${token}`;
  }
  if (cuerpo !== undefined) {
    cabeceras["Content-Type"] = "application/json";
  }

  const respuesta = await fetch(url.toString(), {
    method: metodo,
    headers: cabeceras,
    body: formulario ?? (cuerpo !== undefined ? JSON.stringify(cuerpo) : undefined),
  });

  if (respuesta.status === 204) {
    return undefined as T;
  }

  const texto = await respuesta.text();
  const contenido: unknown = texto ? JSON.parse(texto) : null;

  if (!respuesta.ok) {
    const error = contenido as ErrorApi | null;
    throw new ErrorPeticion(
      respuesta.status,
      error?.error ?? "No se pudo completar la operación",
      error?.detalle ?? {},
    );
  }

  return contenido as T;
}
