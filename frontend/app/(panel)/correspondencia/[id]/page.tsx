"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { useState } from "react";

import { api } from "@/lib/api";
import { ErrorPeticion } from "@/lib/api-client";
import type { EstadoEnvio } from "@/lib/tipos";

const etiquetaEstado: Record<EstadoEnvio, string> = {
  pendiente: "Pendiente",
  enviado: "Enviado",
  entregado: "Entregado",
  rebotado: "Rebotado",
  fallido: "Fallido",
};

const colorEstado: Record<EstadoEnvio, string> = {
  pendiente: "bg-slate-100 text-slate-700",
  enviado: "bg-sky-100 text-sky-800",
  entregado: "bg-emerald-100 text-emerald-800",
  rebotado: "bg-red-100 text-red-800",
  fallido: "bg-red-100 text-red-800",
};

export default function PaginaDetalleCorrespondencia() {
  const parametros = useParams<{ id: string }>();
  const id = parametros.id;
  const clienteConsultas = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const [confirmando, setConfirmando] = useState(false);

  const pieza = useQuery({
    queryKey: ["correspondencia", id],
    queryFn: () => api.obtenerCorrespondencia(id),
  });

  const previsualizacion = useQuery({
    queryKey: ["previsualizacion", id],
    queryFn: () => api.previsualizar(id),
    enabled: pieza.data?.lista_id !== null && pieza.data !== undefined,
    retry: false,
  });

  const esBorrador = pieza.data?.estado === "borrador";

  const resumen = useQuery({
    queryKey: ["resumen", id],
    queryFn: () => api.resumenEnvios(id),
    enabled: pieza.data !== undefined && !esBorrador,
    refetchInterval: pieza.data && !esBorrador ? 5000 : false,
  });

  const envios = useQuery({
    queryKey: ["envios", id],
    queryFn: () => api.listarEnvios(id),
    enabled: pieza.data !== undefined && !esBorrador,
    refetchInterval: pieza.data && !esBorrador ? 5000 : false,
  });

  const envio = useMutation({
    mutationFn: () => api.enviar(id),
    onSuccess: () => {
      setError(null);
      setConfirmando(false);
      void clienteConsultas.invalidateQueries({ queryKey: ["correspondencia", id] });
    },
    onError: (fallo: unknown) => {
      setConfirmando(false);
      setError(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo iniciar el envío");
    },
  });

  if (pieza.isLoading) {
    return <p className="text-sm text-slate-500">Cargando</p>;
  }

  if (!pieza.data) {
    return <p className="text-sm text-red-600">No se encontró la correspondencia</p>;
  }

  const variablesSinValor = previsualizacion.data?.variables_sin_valor ?? [];

  return (
    <section className="space-y-8">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-lg font-semibold">{pieza.data.asunto}</h1>
          <p className="text-sm text-slate-500">Estado: {pieza.data.estado}</p>
        </div>

        {esBorrador && (
          <div className="text-right">
            {confirmando ? (
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => envio.mutate()}
                  disabled={envio.isPending}
                  className="rounded bg-red-600 px-4 py-2 text-sm font-medium text-white disabled:opacity-60"
                >
                  Confirmar envío
                </button>
                <button
                  type="button"
                  onClick={() => setConfirmando(false)}
                  className="rounded border border-slate-300 px-4 py-2 text-sm"
                >
                  Cancelar
                </button>
              </div>
            ) : (
              <button
                type="button"
                onClick={() => setConfirmando(true)}
                disabled={pieza.data.lista_id === null}
                className="rounded bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:opacity-60"
              >
                Enviar correspondencia
              </button>
            )}
            {pieza.data.lista_id === null && (
              <p className="mt-2 text-xs text-slate-500">Asigne una lista para poder enviar</p>
            )}
          </div>
        )}
      </div>

      {error && <p className="rounded bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p>}

      {variablesSinValor.length > 0 && (
        <p className="rounded bg-amber-50 px-3 py-2 text-sm text-amber-800">
          Estas variables quedarían sin reemplazar: {variablesSinValor.join(", ")}
        </p>
      )}

      {previsualizacion.data && (
        <div className="rounded border border-slate-200 bg-white p-6">
          <h2 className="text-sm font-medium">
            Previsualización para {previsualizacion.data.destinatario}
          </h2>
          <p className="mt-3 text-sm font-medium">{previsualizacion.data.asunto}</p>
          <pre className="mt-2 whitespace-pre-wrap font-sans text-sm text-slate-700">
            {previsualizacion.data.cuerpo}
          </pre>

          {pieza.data.adjuntos.length > 0 && (
            <ul className="mt-4 border-t border-slate-100 pt-3 text-xs text-slate-600">
              {pieza.data.adjuntos.map((adjunto) => (
                <li key={adjunto.id}>
                  {adjunto.nombre_archivo} ({Math.round(adjunto.tamano_bytes / 1024)} KB)
                </li>
              ))}
            </ul>
          )}
        </div>
      )}

      {!esBorrador && resumen.data && (
        <div>
          <h2 className="text-sm font-medium">Resultados del envío</h2>
          <div className="mt-3 flex flex-wrap gap-3">
            <span className="rounded border border-slate-200 bg-white px-3 py-2 text-sm">
              Total: {resumen.data.total}
            </span>
            {(Object.keys(etiquetaEstado) as EstadoEnvio[])
              .filter((estado) => resumen.data.por_estado[estado])
              .map((estado) => (
                <span
                  key={estado}
                  className={`rounded px-3 py-2 text-sm font-medium ${colorEstado[estado]}`}
                >
                  {etiquetaEstado[estado]}: {resumen.data.por_estado[estado]}
                </span>
              ))}
          </div>

          <div className="mt-4 overflow-x-auto rounded border border-slate-200 bg-white">
            <table className="w-full text-sm">
              <thead className="border-b border-slate-200 bg-slate-50 text-left">
                <tr>
                  <th className="px-4 py-2 font-medium">Empresa</th>
                  <th className="px-4 py-2 font-medium">Correo</th>
                  <th className="px-4 py-2 font-medium">Estado</th>
                  <th className="px-4 py-2 font-medium">Fecha</th>
                </tr>
              </thead>
              <tbody>
                {envios.data?.datos.map((registro) => (
                  <tr key={registro.id} className="border-b border-slate-100 last:border-0">
                    <td className="px-4 py-2">{registro.empresa}</td>
                    <td className="px-4 py-2 text-slate-600">{registro.correo}</td>
                    <td className="px-4 py-2">
                      <span
                        className={`rounded px-2 py-1 text-xs font-medium ${colorEstado[registro.estado]}`}
                      >
                        {etiquetaEstado[registro.estado]}
                      </span>
                    </td>
                    <td className="px-4 py-2 text-slate-600">
                      {registro.fecha_envio
                        ? new Date(registro.fecha_envio).toLocaleString("es-PE")
                        : "—"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </section>
  );
}
