"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, Loader2, Plus, Search, X } from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useMemo, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { api } from "@/lib/api";
import { ErrorPeticion } from "@/lib/api-client";

export default function PaginaDetalleLista() {
  const parametros = useParams<{ id: string }>();
  const id = parametros.id;
  const clienteConsultas = useQueryClient();
  const [agregarAbierto, setAgregarAbierto] = useState(false);
  const [busqueda, setBusqueda] = useState("");
  const [seleccionados, setSeleccionados] = useState<Set<string>>(new Set());

  const listas = useQuery({ queryKey: ["listas"], queryFn: () => api.listarListas() });
  const lista = listas.data?.datos.find((registro) => registro.id === id);

  const miembros = useQuery({
    queryKey: ["miembros", id],
    queryFn: () => api.miembrosDeLista(id),
  });

  const candidatos = useQuery({
    queryKey: ["contactos", busqueda, 0],
    queryFn: () => api.listarContactos(busqueda || undefined, 100, 0),
    enabled: agregarAbierto,
  });

  const idsEnLista = useMemo(
    () => new Set((miembros.data?.datos ?? []).map((contacto) => contacto.id)),
    [miembros.data],
  );

  const disponibles = (candidatos.data?.datos ?? []).filter(
    (contacto) => !idsEnLista.has(contacto.id),
  );

  function refrescar() {
    void clienteConsultas.invalidateQueries({ queryKey: ["miembros", id] });
    void clienteConsultas.invalidateQueries({ queryKey: ["listas"] });
  }

  const agregado = useMutation({
    mutationFn: () => api.agregarALista(id, Array.from(seleccionados)),
    onSuccess: (resultado) => {
      refrescar();
      setAgregarAbierto(false);
      setSeleccionados(new Set());
      setBusqueda("");
      toast.success(
        resultado.agregados === 1
          ? "Se agregó 1 empresa a la lista"
          : `Se agregaron ${resultado.agregados} empresas a la lista`,
      );
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudieron agregar");
    },
  });

  const quitado = useMutation({
    mutationFn: (contactoId: string) => api.quitarDeLista(id, contactoId),
    onSuccess: () => {
      refrescar();
      toast.success("Empresa quitada de la lista");
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo quitar");
    },
  });

  function alternar(contactoId: string) {
    setSeleccionados((actual) => {
      const copia = new Set(actual);
      if (copia.has(contactoId)) {
        copia.delete(contactoId);
      } else {
        copia.add(contactoId);
      }
      return copia;
    });
  }

  function alternarTodos() {
    setSeleccionados((actual) =>
      actual.size === disponibles.length
        ? new Set()
        : new Set(disponibles.map((contacto) => contacto.id)),
    );
  }

  const contactos = miembros.data?.datos ?? [];

  return (
    <section className="space-y-6">
      <Button asChild variant="ghost" size="sm" className="-ml-2 text-muted-foreground">
        <Link href="/listas">
          <ArrowLeft className="mr-2 h-4 w-4" aria-hidden="true" />
          Volver a listas
        </Link>
      </Button>

      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            {lista?.nombre ?? "Lista de contactos"}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {lista?.descripcion ? `${lista.descripcion} · ` : ""}
            {miembros.isLoading ? "Cargando" : `${miembros.data?.total ?? 0} empresas`}
          </p>
        </div>

        <Button type="button" onClick={() => setAgregarAbierto(true)}>
          <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
          Agregar empresas
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Empresas en la lista</CardTitle>
          <CardDescription>
            Estas son las destinatarias que recibirán la correspondencia dirigida a esta lista.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {miembros.isLoading ? (
            <div className="space-y-2">
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
            </div>
          ) : contactos.length === 0 ? (
            <div className="py-12 text-center">
              <p className="text-sm font-medium">Esta lista todavía está vacía</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Agregue empresas para poder enviarles correspondencia.
              </p>
              <Button type="button" className="mt-4" onClick={() => setAgregarAbierto(true)}>
                <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
                Agregar empresas
              </Button>
            </div>
          ) : (
            <ul className="divide-y divide-zinc-100">
              {contactos.map((contacto) => (
                <li key={contacto.id} className="flex items-center justify-between gap-4 py-3">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{contacto.empresa}</p>
                    <p className="truncate text-xs text-muted-foreground">
                      {contacto.nombre} · {contacto.correo}
                    </p>
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => quitado.mutate(contacto.id)}
                    aria-label={`Quitar ${contacto.empresa} de la lista`}
                    className="h-8 w-8 shrink-0 text-muted-foreground hover:text-destructive"
                  >
                    <X className="h-4 w-4" aria-hidden="true" />
                  </Button>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      <Dialog open={agregarAbierto} onOpenChange={setAgregarAbierto}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Agregar empresas a la lista</DialogTitle>
            <DialogDescription>
              Marque las empresas que deben recibir esta correspondencia.
            </DialogDescription>
          </DialogHeader>

          <div className="relative">
            <Search
              className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
              aria-hidden="true"
            />
            <Label htmlFor="buscar-empresas" className="sr-only">
              Buscar empresas
            </Label>
            <Input
              id="buscar-empresas"
              type="search"
              value={busqueda}
              onChange={(evento) => setBusqueda(evento.target.value)}
              className="pl-9"
            />
          </div>

          {disponibles.length > 0 && (
            <div className="flex items-center justify-between border-b border-zinc-100 pb-2">
              <button
                type="button"
                onClick={alternarTodos}
                className="text-xs font-medium text-emerald-700 hover:underline"
              >
                {seleccionados.size === disponibles.length
                  ? "Quitar la selección"
                  : "Seleccionar las visibles"}
              </button>
              <span className="text-xs text-muted-foreground">
                {seleccionados.size} seleccionadas
              </span>
            </div>
          )}

          <div className="max-h-72 overflow-y-auto">
            {candidatos.isLoading ? (
              <div className="space-y-2 py-2">
                <Skeleton className="h-10 w-full" />
                <Skeleton className="h-10 w-full" />
              </div>
            ) : disponibles.length === 0 ? (
              <p className="py-8 text-center text-sm text-muted-foreground">
                {busqueda
                  ? "Ninguna empresa disponible coincide con la búsqueda."
                  : "Todas las empresas registradas ya están en esta lista."}
              </p>
            ) : (
              <ul className="divide-y divide-zinc-100">
                {disponibles.map((contacto) => (
                  <li key={contacto.id}>
                    <label className="flex cursor-pointer items-center gap-3 py-2.5">
                      <Checkbox
                        checked={seleccionados.has(contacto.id)}
                        onCheckedChange={() => alternar(contacto.id)}
                        aria-label={`Seleccionar ${contacto.empresa}`}
                      />
                      <span className="min-w-0">
                        <span className="block truncate text-sm font-medium">
                          {contacto.empresa}
                        </span>
                        <span className="block truncate text-xs text-muted-foreground">
                          {contacto.nombre} · {contacto.correo}
                        </span>
                      </span>
                    </label>
                  </li>
                ))}
              </ul>
            )}
          </div>

          <DialogFooter className="gap-2 sm:gap-2">
            <Button type="button" variant="outline" onClick={() => setAgregarAbierto(false)}>
              Cancelar
            </Button>
            <Button
              type="button"
              onClick={() => agregado.mutate()}
              disabled={agregado.isPending || seleccionados.size === 0}
              className="text-white"
            >
              {agregado.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
              )}
              Agregar {seleccionados.size > 0 ? seleccionados.size : ""}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  );
}
