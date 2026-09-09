"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import { api } from "@/lib/api";
import { ErrorPeticion } from "@/lib/api-client";

export default function PaginaListas() {
  const [nombre, setNombre] = useState("");
  const [descripcion, setDescripcion] = useState("");
  const [seleccionada, setSeleccionada] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const clienteConsultas = useQueryClient();

  const listas = useQuery({ queryKey: ["listas"], queryFn: () => api.listarListas() });

  const miembros = useQuery({
    queryKey: ["miembros", seleccionada],
    queryFn: () => api.miembrosDeLista(seleccionada as string),
    enabled: seleccionada !== null,
  });

  const creacion = useMutation({
    mutationFn: () => api.crearLista({ nombre, descripcion: descripcion || null }),
    onSuccess: () => {
      setNombre("");
      setDescripcion("");
      setError(null);
      void clienteConsultas.invalidateQueries({ queryKey: ["listas"] });
    },
    onError: (fallo: unknown) => {
      setError(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo crear la lista");
    },
  });

  const eliminacion = useMutation({
    mutationFn: (id: string) => api.eliminarLista(id),
    onSuccess: (_datos, id) => {
      if (seleccionada === id) {
        setSeleccionada(null);
      }
      void clienteConsultas.invalidateQueries({ queryKey: ["listas"] });
    },
  });

  return (
    <section className="grid gap-8 lg:grid-cols-2">
      <div>
        <h1 className="text-lg font-semibold">Listas de contactos</h1>

        <form
          onSubmit={(evento) => {
            evento.preventDefault();
            creacion.mutate();
          }}
          className="mt-4 space-y-3 rounded border border-slate-200 bg-white p-4"
        >
          <div>
            <label htmlFor="nombre" className="block text-sm font-medium">
              Nombre de la lista
            </label>
            <input
              id="nombre"
              value={nombre}
              onChange={(evento) => setNombre(evento.target.value)}
              className="mt-1 w-full rounded border border-slate-300 px-3 py-2 text-sm outline-none focus:border-slate-900"
            />
          </div>

          <div>
            <label htmlFor="descripcion" className="block text-sm font-medium">
              Descripción
            </label>
            <input
              id="descripcion"
              value={descripcion}
              onChange={(evento) => setDescripcion(evento.target.value)}
              className="mt-1 w-full rounded border border-slate-300 px-3 py-2 text-sm outline-none focus:border-slate-900"
            />
          </div>

          {error && <p className="text-xs text-red-600">{error}</p>}

          <button
            type="submit"
            disabled={creacion.isPending || nombre.trim().length < 2}
            className="rounded bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:opacity-60"
          >
            Crear lista
          </button>
        </form>

        <ul className="mt-4 divide-y divide-slate-100 rounded border border-slate-200 bg-white">
          {listas.data?.datos.map((lista) => (
            <li key={lista.id} className="flex items-center justify-between px-4 py-3">
              <button
                type="button"
                onClick={() => setSeleccionada(lista.id)}
                className="text-left"
              >
                <span className="block text-sm font-medium">{lista.nombre}</span>
                <span className="block text-xs text-slate-500">
                  {lista.total_contactos} contactos
                  {lista.descripcion ? ` · ${lista.descripcion}` : ""}
                </span>
              </button>
              <button
                type="button"
                onClick={() => eliminacion.mutate(lista.id)}
                className="text-xs text-red-600 hover:underline"
              >
                Eliminar
              </button>
            </li>
          ))}
          {listas.data?.datos.length === 0 && (
            <li className="px-4 py-6 text-center text-sm text-slate-500">
              Aún no hay listas creadas
            </li>
          )}
        </ul>
      </div>

      <div>
        <h2 className="text-lg font-semibold">Contactos de la lista</h2>
        {seleccionada === null ? (
          <p className="mt-4 text-sm text-slate-500">Seleccione una lista para ver sus contactos</p>
        ) : (
          <ul className="mt-4 divide-y divide-slate-100 rounded border border-slate-200 bg-white">
            {miembros.data?.datos.map((contacto) => (
              <li key={contacto.id} className="px-4 py-3">
                <span className="block text-sm">{contacto.empresa}</span>
                <span className="block text-xs text-slate-500">
                  {contacto.nombre} · {contacto.correo}
                </span>
              </li>
            ))}
            {miembros.data?.datos.length === 0 && (
              <li className="px-4 py-6 text-center text-sm text-slate-500">
                Esta lista todavía no tiene contactos
              </li>
            )}
          </ul>
        )}
      </div>
    </section>
  );
}
