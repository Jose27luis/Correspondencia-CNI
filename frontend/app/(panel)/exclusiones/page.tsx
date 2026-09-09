"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, ShieldOff, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { api } from "@/lib/api";
import { ErrorPeticion } from "@/lib/api-client";
import type { Supresion } from "@/lib/tipos";

const etiquetaMotivo: Record<Supresion["motivo"], string> = {
  rebote: "Rebotó",
  baja: "Pidió la baja",
  manual: "Excluido a mano",
};

const colorMotivo: Record<Supresion["motivo"], string> = {
  rebote: "bg-red-100 text-red-800 hover:bg-red-100",
  baja: "bg-amber-100 text-amber-800 hover:bg-amber-100",
  manual: "bg-zinc-100 text-zinc-700 hover:bg-zinc-100",
};

export default function PaginaExclusiones() {
  const [correo, setCorreo] = useState("");
  const [detalle, setDetalle] = useState("");
  const clienteConsultas = useQueryClient();

  const consulta = useQuery({
    queryKey: ["exclusiones"],
    queryFn: () => api.listarExclusiones(),
  });

  function refrescar() {
    void clienteConsultas.invalidateQueries({ queryKey: ["exclusiones"] });
  }

  const agregado = useMutation({
    mutationFn: () => api.excluirCorreo(correo, detalle),
    onSuccess: () => {
      refrescar();
      setCorreo("");
      setDetalle("");
      toast.success("Correo excluido", {
        description: "No volverá a recibir correspondencia de CNI.",
      });
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo excluir el correo");
    },
  });

  const quitado = useMutation({
    mutationFn: (valor: string) => api.quitarExclusion(valor),
    onSuccess: () => {
      refrescar();
      toast.success("Exclusión retirada");
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo quitar");
    },
  });

  const datos = consulta.data?.datos ?? [];

  return (
    <section className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Correos excluidos</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Estos correos quedan fuera de todos los envíos, aunque sigan en sus listas.
        </p>
      </div>

      <Card className="border-dashed">
        <CardContent className="p-5">
          <p className="text-sm text-muted-foreground">
            Un correo entra aquí automáticamente cuando rebota o cuando el destinatario pide la
            baja. Seguir escribiendo a direcciones que rebotan daña la reputación del dominio y
            hace que los demás correos de CNI acaben en la carpeta de spam.
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Excluir un correo</CardTitle>
          <CardDescription>
            Use esta opción cuando una empresa pida por teléfono o en persona no recibir más
            correspondencia.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={(evento) => {
              evento.preventDefault();
              agregado.mutate();
            }}
            className="flex flex-wrap items-end gap-3"
          >
            <div className="min-w-56 flex-1 space-y-1.5">
              <Label htmlFor="correo">Correo</Label>
              <Input
                id="correo"
                type="email"
                value={correo}
                onChange={(evento) => setCorreo(evento.target.value)}
              />
            </div>
            <div className="min-w-56 flex-1 space-y-1.5">
              <Label htmlFor="detalle">Motivo</Label>
              <Input
                id="detalle"
                value={detalle}
                onChange={(evento) => setDetalle(evento.target.value)}
              />
            </div>
            <Button
              type="submit"
              disabled={agregado.isPending || !correo.includes("@")}
              className="text-white"
            >
              {agregado.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
              )}
              Excluir
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          {consulta.isLoading ? (
            <div className="space-y-2 p-5">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : datos.length === 0 ? (
            <div className="py-12 text-center">
              <ShieldOff className="mx-auto h-8 w-8 text-muted-foreground" aria-hidden="true" />
              <p className="mt-3 text-sm font-medium">No hay correos excluidos</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Todos los contactos de sus listas reciben la correspondencia.
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Correo</TableHead>
                  <TableHead>Motivo</TableHead>
                  <TableHead>Detalle</TableHead>
                  <TableHead>Fecha</TableHead>
                  <TableHead className="w-12" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {datos.map((registro) => (
                  <TableRow key={registro.correo}>
                    <TableCell className="font-medium">{registro.correo}</TableCell>
                    <TableCell>
                      <Badge variant="secondary" className={colorMotivo[registro.motivo]}>
                        {etiquetaMotivo[registro.motivo]}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {registro.detalle ?? "—"}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(registro.creado_en).toLocaleDateString("es-PE")}
                    </TableCell>
                    <TableCell>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => quitado.mutate(registro.correo)}
                        aria-label={`Quitar la exclusión de ${registro.correo}`}
                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                      >
                        <Trash2 className="h-4 w-4" aria-hidden="true" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </section>
  );
}
