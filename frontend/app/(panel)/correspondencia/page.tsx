"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";

import { api } from "@/lib/api";
import type { EstadoCorrespondencia } from "@/lib/tipos";

const colorPorEstado: Record<EstadoCorrespondencia, string> = {
  borrador: "bg-slate-100 text-slate-700",
  encolada: "bg-amber-100 text-amber-800",
  enviando: "bg-amber-100 text-amber-800",
  enviada: "bg-emerald-100 text-emerald-800",
  fallida: "bg-red-100 text-red-800",
};

export default function PaginaCorrespondencia() {
  const consulta = useQuery({
    queryKey: ["correspondencia"],
    queryFn: () => api.listarCorrespondencia(),
  });

  return (
    <section>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold">Correspondencia</h1>
          <p className="text-sm text-slate-500">
            {consulta.data ? `${consulta.data.total} piezas registradas` : "Cargando"}
          </p>
        </div>

        <Link
          href="/correspondencia/nueva"
          className="rounded bg-slate-900 px-4 py-2 text-sm font-medium text-white"
        >
          Nueva correspondencia
        </Link>
      </div>

      <ul className="mt-6 divide-y divide-slate-100 rounded border border-slate-200 bg-white">
        {consulta.data?.datos.map((pieza) => (
          <li key={pieza.id}>
            <Link
              href={`/correspondencia/${pieza.id}`}
              className="flex items-center justify-between px-4 py-3 hover:bg-slate-50"
            >
              <div>
                <span className="block text-sm font-medium">{pieza.asunto}</span>
                <span className="block text-xs text-slate-500">
                  Actualizada el {new Date(pieza.actualizado_en).toLocaleString("es-PE")}
                </span>
              </div>
              <span
                className={`rounded px-2 py-1 text-xs font-medium ${colorPorEstado[pieza.estado]}`}
              >
                {pieza.estado}
              </span>
            </Link>
          </li>
        ))}
        {consulta.data?.datos.length === 0 && (
          <li className="px-4 py-6 text-center text-sm text-slate-500">
            Aún no hay correspondencia registrada
          </li>
        )}
      </ul>
    </section>
  );
}
