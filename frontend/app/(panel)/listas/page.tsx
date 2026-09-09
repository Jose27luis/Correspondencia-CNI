"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronRight, Loader2, Plus, Trash2, Users } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { toast } from "sonner";

import { DialogoConfirmacion } from "@/components/dialogo-confirmacion";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
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
import type { Lista } from "@/lib/tipos";

export default function PaginaListas() {
  const [formularioAbierto, setFormularioAbierto] = useState(false);
  const [porEliminar, setPorEliminar] = useState<Lista | null>(null);
  const [nombre, setNombre] = useState("");
  const [descripcion, setDescripcion] = useState("");
  const clienteConsultas = useQueryClient();

  const listas = useQuery({ queryKey: ["listas"], queryFn: () => api.listarListas() });

  function refrescar() {
    void clienteConsultas.invalidateQueries({ queryKey: ["listas"] });
  }

  const creacion = useMutation({
    mutationFn: () => api.crearLista({ nombre, descripcion: descripcion.trim() || null }),
    onSuccess: (lista) => {
      refrescar();
      setFormularioAbierto(false);
      setNombre("");
      setDescripcion("");
      toast.success(`Lista "${lista.nombre}" creada`, {
        description: "Ahora agregue las empresas que la integran.",
      });
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo crear la lista");
    },
  });

  const eliminacion = useMutation({
    mutationFn: (id: string) => api.eliminarLista(id),
    onSuccess: () => {
      refrescar();
      setPorEliminar(null);
      toast.success("Lista eliminada");
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo eliminar la lista");
    },
  });

  const datos = listas.data?.datos ?? [];

  return (
    <section className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Listas</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Agrupe empresas para enviarles la misma correspondencia.
          </p>
        </div>

        <Button type="button" onClick={() => setFormularioAbierto(true)}>
          <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
          Nueva lista
        </Button>
      </div>

      {listas.isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2">
          <Skeleton className="h-28 w-full" />
          <Skeleton className="h-28 w-full" />
        </div>
      ) : datos.length === 0 ? (
        <Card className="border-dashed">
          <CardContent className="py-12 text-center">
            <p className="text-sm font-medium">Todavía no hay listas</p>
            <p className="mt-1 text-sm text-muted-foreground">
              Cree una lista, por ejemplo Compradores de maíz, y agréguele las empresas.
            </p>
            <Button type="button" className="mt-4" onClick={() => setFormularioAbierto(true)}>
              <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
              Nueva lista
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {datos.map((lista) => (
            <Card key={lista.id} className="group transition-shadow hover:shadow-md">
              <CardContent className="flex items-start justify-between gap-3 p-5">
                <Link href={`/listas/${lista.id}`} className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{lista.nombre}</p>
                  {lista.descripcion && (
                    <p className="mt-0.5 truncate text-xs text-muted-foreground">
                      {lista.descripcion}
                    </p>
                  )}
                  <p className="mt-3 flex items-center gap-1.5 text-xs text-muted-foreground">
                    <Users className="h-3.5 w-3.5" aria-hidden="true" />
                    {lista.total_contactos === 0
                      ? "Sin empresas asignadas"
                      : `${lista.total_contactos} empresas`}
                  </p>
                </Link>

                <div className="flex shrink-0 items-center gap-1">
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => setPorEliminar(lista)}
                    aria-label={`Eliminar la lista ${lista.nombre}`}
                    className="h-8 w-8 text-muted-foreground hover:text-destructive"
                  >
                    <Trash2 className="h-4 w-4" aria-hidden="true" />
                  </Button>
                  <Link
                    href={`/listas/${lista.id}`}
                    aria-label={`Abrir la lista ${lista.nombre}`}
                    className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-zinc-100 hover:text-foreground"
                  >
                    <ChevronRight className="h-4 w-4" aria-hidden="true" />
                  </Link>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={formularioAbierto} onOpenChange={setFormularioAbierto}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Nueva lista</DialogTitle>
            <DialogDescription>
              Después de crearla podrá agregarle las empresas destinatarias.
            </DialogDescription>
          </DialogHeader>

          <form
            id="formulario-lista"
            onSubmit={(evento) => {
              evento.preventDefault();
              creacion.mutate();
            }}
            className="space-y-4"
          >
            <div className="space-y-1.5">
              <Label htmlFor="nombre">Nombre de la lista</Label>
              <Input
                id="nombre"
                value={nombre}
                onChange={(evento) => setNombre(evento.target.value)}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="descripcion">Descripción</Label>
              <Input
                id="descripcion"
                value={descripcion}
                onChange={(evento) => setDescripcion(evento.target.value)}
              />
            </div>
          </form>

          <DialogFooter className="gap-2 sm:gap-2">
            <Button type="button" variant="outline" onClick={() => setFormularioAbierto(false)}>
              Cancelar
            </Button>
            <Button
              type="submit"
              form="formulario-lista"
              disabled={creacion.isPending || nombre.trim().length < 2}
              className="text-white"
            >
              {creacion.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
              )}
              Crear lista
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <DialogoConfirmacion
        abierto={porEliminar !== null}
        titulo="Eliminar lista"
        descripcion={
          porEliminar
            ? `Se eliminará la lista "${porEliminar.nombre}". Las empresas seguirán en la base de contactos, solo se pierde la agrupación.`
            : ""
        }
        procesando={eliminacion.isPending}
        alCambiar={(abierto) => {
          if (!abierto) {
            setPorEliminar(null);
          }
        }}
        alConfirmar={() => {
          if (porEliminar) {
            eliminacion.mutate(porEliminar.id);
          }
        }}
      />
    </section>
  );
}
