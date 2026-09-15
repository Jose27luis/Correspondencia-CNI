"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Building2, Check, ExternalLink, X } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { DialogoRechazo } from "@/components/dialogo-rechazo";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorPeticion } from "@/lib/api-client";
import {
  apiCaptacion,
  colorRevision,
  etiquetaRevision,
  type EmpresaDirectorio,
  type EstadoRevision,
} from "@/lib/intake";
import { cn } from "@/lib/utils";

const filtros: { valor: EstadoRevision | undefined; texto: string }[] = [
  { valor: "pendiente", texto: "Pendientes" },
  { valor: "aprobado", texto: "Aprobadas" },
  { valor: "rechazado", texto: "Rechazadas" },
  { valor: undefined, texto: "Todas" },
];

function avisarError(fallo: unknown, mensaje: string) {
  toast.error(fallo instanceof ErrorPeticion ? fallo.message : mensaje);
}

export default function PaginaRevisionDirectorio() {
  const [estado, setEstado] = useState<EstadoRevision | undefined>("pendiente");
  const [aRechazar, setARechazar] = useState<EmpresaDirectorio | null>(null);
  const clienteConsultas = useQueryClient();

  const consulta = useQuery({
    queryKey: ["revision-directorio", estado],
    queryFn: () => apiCaptacion.listarDirectorio(estado),
  });

  function refrescar() {
    void clienteConsultas.invalidateQueries({ queryKey: ["revision-directorio"] });
  }

  const aprobacion = useMutation({
    mutationFn: (id: string) => apiCaptacion.aprobarDirectorio(id),
    onSuccess: (empresa) => {
      refrescar();
      toast.success("Empresa publicada", {
        description: `${empresa.nombre} ya está en Contactos y en la lista Directorio C-E.`,
      });
    },
    onError: (fallo: unknown) => avisarError(fallo, "No se pudo aprobar"),
  });

  const rechazo = useMutation({
    mutationFn: ({ id, motivo }: { id: string; motivo: string }) =>
      apiCaptacion.rechazarDirectorio(id, motivo),
    onSuccess: () => {
      refrescar();
      setARechazar(null);
      toast.success("Solicitud rechazada");
    },
    onError: (fallo: unknown) => avisarError(fallo, "No se pudo rechazar"),
  });

  const datos = consulta.data?.datos ?? [];

  return (
    <section className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Directorio C-E</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Solicitudes que llegan desde la página pública. Al aprobar, la empresa se publica y pasa
            a Contactos.
          </p>
        </div>
        <Button asChild variant="outline">
          <a href="/directorio" target="_blank" rel="noreferrer">
            <ExternalLink className="mr-2 h-4 w-4" aria-hidden="true" />
            Ver página pública
          </a>
        </Button>
      </div>

      <div className="flex flex-wrap gap-2">
        {filtros.map((opcion) => (
          <Button
            key={opcion.texto}
            type="button"
            size="sm"
            variant={estado === opcion.valor ? "default" : "outline"}
            onClick={() => setEstado(opcion.valor)}
            className={cn(estado === opcion.valor && "text-white")}
          >
            {opcion.texto}
          </Button>
        ))}
      </div>

      {consulta.isLoading ? (
        <Skeleton className="h-40 w-full" />
      ) : datos.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Building2 className="mx-auto h-8 w-8 text-muted-foreground" aria-hidden="true" />
            <p className="mt-3 text-sm font-medium">No hay solicitudes en este estado</p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {datos.map((empresa) => (
            <Card key={empresa.id}>
              <CardContent className="flex flex-wrap gap-4 p-5">
                {empresa.logo_url ? (
                  <img
                    src={empresa.logo_url}
                    alt={`Logo de ${empresa.nombre}`}
                    className="h-14 w-14 rounded-md border object-contain"
                  />
                ) : (
                  <Building2 className="h-14 w-14 text-muted-foreground" aria-hidden="true" />
                )}
                <div className="min-w-0 flex-1 space-y-1 text-sm">
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="font-semibold">{empresa.nombre}</p>
                    <Badge variant="secondary" className={colorRevision[empresa.estado]}>
                      {etiquetaRevision[empresa.estado]}
                    </Badge>
                  </div>
                  <p className="text-muted-foreground">
                    RUC {empresa.ruc} · {empresa.correo} · {empresa.ciudad}
                  </p>
                  <p className="text-muted-foreground">{empresa.direccion}</p>
                  <p>{empresa.descripcion}</p>
                  <p className="text-muted-foreground">
                    {[empresa.telefono, empresa.celular, empresa.pagina_web, empresa.facebook]
                      .filter(Boolean)
                      .join(" · ")}
                  </p>
                  {empresa.motivo_rechazo && (
                    <p className="text-red-700">Motivo del rechazo: {empresa.motivo_rechazo}</p>
                  )}
                </div>
                {empresa.estado !== "aprobado" && (
                  <div className="flex items-start gap-2">
                    <Button
                      type="button"
                      size="sm"
                      disabled={aprobacion.isPending}
                      onClick={() => aprobacion.mutate(empresa.id)}
                      className="text-white"
                    >
                      <Check className="mr-1 h-4 w-4" aria-hidden="true" />
                      Aprobar
                    </Button>
                    {empresa.estado === "pendiente" && (
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        onClick={() => setARechazar(empresa)}
                      >
                        <X className="mr-1 h-4 w-4" aria-hidden="true" />
                        Rechazar
                      </Button>
                    )}
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <DialogoRechazo
        abierto={aRechazar !== null}
        nombre={aRechazar?.nombre ?? ""}
        pendiente={rechazo.isPending}
        alCerrar={() => setARechazar(null)}
        alConfirmar={(motivo) => {
          if (aRechazar) {
            rechazo.mutate({ id: aRechazar.id, motivo });
          }
        }}
      />
    </section>
  );
}
