"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRef, useState } from "react";

import { api } from "@/lib/api";
import { ErrorPeticion } from "@/lib/api-client";
import type { ResumenImportacion } from "@/lib/tipos";

export default function PaginaContactos() {
  const [busqueda, setBusqueda] = useState("");
  const [resumen, setResumen] = useState<ResumenImportacion | null>(null);
  const [error, setError] = useState<string | null>(null);
  const referenciaArchivo = useRef<HTMLInputElement>(null);
  const clienteConsultas = useQueryClient();

  const consulta = useQuery({
    queryKey: ["contactos", busqueda],
    queryFn: () => api.listarContactos(busqueda || undefined),
  });

  const importacion = useMutation({
    mutationFn: (archivo: File) => api.importarContactos(archivo),
    onSuccess: (datos) => {
      setResumen(datos);
      setError(null);
      void clienteConsultas.invalidateQueries({ queryKey: ["contactos"] });
    },
    onError: (fallo: unknown) => {
      setResumen(null);
      setError(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo importar el archivo");
    },
  });

  const eliminacion = useMutation({
    mutationFn: (id: string) => api.eliminarContacto(id),
    onSuccess: () => {
      void clienteConsultas.invalidateQueries({ queryKey: ["contactos"] });
    },
  });

  function seleccionarArchivo(evento: React.ChangeEvent<HTMLInputElement>) {
    const archivo = evento.target.files?.[0];
    if (archivo) {
      importacion.mutate(archivo);
    }
    evento.target.value = "";
  }

  return (
    <section>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold">Contactos</h1>
          <p className="text-sm text-slate-500">
            {consulta.data ? `${consulta.data.total} empresas registradas` : "Cargando"}
          </p>
        </div>

        <div>
          <input
            ref={referenciaArchivo}
            type="file"
            accept=".csv,text/csv"
            className="hidden"
            onChange={seleccionarArchivo}
          />
          <button
            type="button"
            onClick={() => referenciaArchivo.current?.click()}
            disabled={importacion.isPending}
            className="rounded bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:opacity-60"
          >
            {importacion.isPending ? "Importando" : "Importar CSV"}
          </button>
        </div>
      </div>

      {resumen && (
        <div className="mt-4 rounded border border-slate-200 bg-white p-4 text-sm">
          <p>
            Creados: {resumen.creados}, actualizados: {resumen.actualizados}, omitidos:{" "}
            {resumen.omitidos}
          </p>
          {resumen.errores.length > 0 && (
            <ul className="mt-2 list-disc pl-5 text-xs text-amber-700">
              {resumen.errores.slice(0, 10).map((linea) => (
                <li key={linea}>{linea}</li>
              ))}
            </ul>
          )}
        </div>
      )}

      {error && (
        <p className="mt-4 rounded bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p>
      )}

      <div className="mt-6">
        <label htmlFor="busqueda" className="block text-sm font-medium">
          Buscar por nombre, empresa o correo
        </label>
        <input
          id="busqueda"
          type="search"
          value={busqueda}
          onChange={(evento) => setBusqueda(evento.target.value)}
          className="mt-1 w-full max-w-md rounded border border-slate-300 px-3 py-2 text-sm outline-none focus:border-slate-900"
        />
      </div>

      <div className="mt-4 overflow-x-auto rounded border border-slate-200 bg-white">
        <table className="w-full text-sm">
          <thead className="border-b border-slate-200 bg-slate-50 text-left">
            <tr>
              <th className="px-4 py-2 font-medium">Nombre</th>
              <th className="px-4 py-2 font-medium">Empresa</th>
              <th className="px-4 py-2 font-medium">Correo</th>
              <th className="px-4 py-2 font-medium">País</th>
              <th className="px-4 py-2" />
            </tr>
          </thead>
          <tbody>
            {consulta.data?.datos.map((contacto) => (
              <tr key={contacto.id} className="border-b border-slate-100 last:border-0">
                <td className="px-4 py-2">{contacto.nombre}</td>
                <td className="px-4 py-2">{contacto.empresa}</td>
                <td className="px-4 py-2 text-slate-600">{contacto.correo}</td>
                <td className="px-4 py-2 text-slate-600">{contacto.pais ?? "—"}</td>
                <td className="px-4 py-2 text-right">
                  <button
                    type="button"
                    onClick={() => eliminacion.mutate(contacto.id)}
                    className="text-xs text-red-600 hover:underline"
                  >
                    Eliminar
                  </button>
                </td>
              </tr>
            ))}
            {consulta.data?.datos.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-6 text-center text-slate-500">
                  No hay contactos que coincidan
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </section>
  );
}
