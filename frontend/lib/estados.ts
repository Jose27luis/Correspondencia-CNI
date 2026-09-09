import type { EstadoCorrespondencia, EstadoEnvio } from "./tipos";

export const etiquetaEstadoCorrespondencia: Record<EstadoCorrespondencia, string> = {
  borrador: "Borrador",
  encolada: "En cola",
  enviando: "Enviando",
  enviada: "Enviada",
  fallida: "Fallida",
};

export const colorEstadoCorrespondencia: Record<EstadoCorrespondencia, string> = {
  borrador: "bg-zinc-100 text-zinc-700 hover:bg-zinc-100",
  encolada: "bg-amber-100 text-amber-800 hover:bg-amber-100",
  enviando: "bg-amber-100 text-amber-800 hover:bg-amber-100",
  enviada: "bg-emerald-100 text-emerald-800 hover:bg-emerald-100",
  fallida: "bg-red-100 text-red-800 hover:bg-red-100",
};

export const etiquetaEstadoEnvio: Record<EstadoEnvio, string> = {
  pendiente: "Pendiente",
  enviado: "Enviado",
  entregado: "Entregado",
  rebotado: "Rebotado",
  fallido: "Fallido",
};

export const colorEstadoEnvio: Record<EstadoEnvio, string> = {
  pendiente: "bg-zinc-100 text-zinc-700 hover:bg-zinc-100",
  enviado: "bg-sky-100 text-sky-800 hover:bg-sky-100",
  entregado: "bg-emerald-100 text-emerald-800 hover:bg-emerald-100",
  rebotado: "bg-red-100 text-red-800 hover:bg-red-100",
  fallido: "bg-red-100 text-red-800 hover:bg-red-100",
};
