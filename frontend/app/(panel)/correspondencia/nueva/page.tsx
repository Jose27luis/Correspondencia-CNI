"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { api } from "@/lib/api";
import { ErrorPeticion } from "@/lib/api-client";

export default function PaginaNuevaCorrespondencia() {
  const router = useRouter();
  const [asunto, setAsunto] = useState("");
  const [cuerpo, setCuerpo] = useState("");
  const [listaId, setListaId] = useState("");
  const [error, setError] = useState<string | null>(null);

  const listas = useQuery({ queryKey: ["listas"], queryFn: () => api.listarListas() });

  const creacion = useMutation({
    mutationFn: () =>
      api.crearCorrespondencia({
        asunto,
        cuerpo,
        lista_id: listaId || null,
      }),
    onSuccess: (pieza) => {
      router.push(`/correspondencia/${pieza.id}`);
    },
    onError: (fallo: unknown) => {
      setError(
        fallo instanceof ErrorPeticion ? fallo.message : "No se pudo guardar la correspondencia",
      );
    },
  });

  return (
    <section className="max-w-3xl">
      <h1 className="text-lg font-semibold">Nueva correspondencia</h1>
      <p className="mt-1 text-sm text-slate-500">
        Use variables como {"{nombre}"}, {"{empresa}"} o {"{pais}"} para personalizar cada carta.
      </p>

      <form
        onSubmit={(evento) => {
          evento.preventDefault();
          setError(null);
          creacion.mutate();
        }}
        className="mt-6 space-y-4 rounded border border-slate-200 bg-white p-6"
      >
        <div>
          <label htmlFor="lista" className="block text-sm font-medium">
            Lista de destinatarios
          </label>
          <select
            id="lista"
            value={listaId}
            onChange={(evento) => setListaId(evento.target.value)}
            className="mt-1 w-full rounded border border-slate-300 px-3 py-2 text-sm outline-none focus:border-slate-900"
          >
            <option value="">Sin lista asignada</option>
            {listas.data?.datos.map((lista) => (
              <option key={lista.id} value={lista.id}>
                {lista.nombre} ({lista.total_contactos} contactos)
              </option>
            ))}
          </select>
        </div>

        <div>
          <label htmlFor="asunto" className="block text-sm font-medium">
            Asunto
          </label>
          <input
            id="asunto"
            value={asunto}
            onChange={(evento) => setAsunto(evento.target.value)}
            className="mt-1 w-full rounded border border-slate-300 px-3 py-2 text-sm outline-none focus:border-slate-900"
          />
        </div>

        <div>
          <label htmlFor="cuerpo" className="block text-sm font-medium">
            Cuerpo del mensaje
          </label>
          <textarea
            id="cuerpo"
            rows={14}
            value={cuerpo}
            onChange={(evento) => setCuerpo(evento.target.value)}
            className="mt-1 w-full rounded border border-slate-300 px-3 py-2 font-mono text-sm outline-none focus:border-slate-900"
          />
        </div>

        {error && <p className="rounded bg-red-50 px-3 py-2 text-xs text-red-700">{error}</p>}

        <div className="flex gap-3">
          <button
            type="submit"
            disabled={creacion.isPending || asunto.trim().length < 3 || cuerpo.trim().length < 10}
            className="rounded bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:opacity-60"
          >
            Guardar borrador
          </button>
          <button
            type="button"
            onClick={() => router.back()}
            className="rounded border border-slate-300 px-4 py-2 text-sm"
          >
            Cancelar
          </button>
        </div>
      </form>
    </section>
  );
}
