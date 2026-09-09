"use client";

import { useQuery } from "@tanstack/react-query";
import { FileText, Plus } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { api } from "@/lib/api";
import { colorEstadoCorrespondencia, etiquetaEstadoCorrespondencia } from "@/lib/estados";
import type { EstadoCorrespondencia } from "@/lib/tipos";

const filtros: { valor: EstadoCorrespondencia | "todas"; texto: string }[] = [
  { valor: "todas", texto: "Todas" },
  { valor: "borrador", texto: "Borradores" },
  { valor: "encolada", texto: "En cola" },
  { valor: "enviada", texto: "Enviadas" },
  { valor: "fallida", texto: "Fallidas" },
];

export default function PaginaCorrespondencia() {
  const [filtro, setFiltro] = useState<EstadoCorrespondencia | "todas">("todas");

  const consulta = useQuery({
    queryKey: ["correspondencia", filtro === "todas" ? undefined : filtro],
    queryFn: () => api.listarCorrespondencia(filtro === "todas" ? undefined : filtro),
  });

  const piezas = consulta.data?.datos ?? [];

  return (
    <section className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Correspondencia</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {consulta.isLoading ? "Cargando" : `${consulta.data?.total ?? 0} cartas registradas`}
          </p>
        </div>

        <Button asChild>
          <Link href="/correspondencia/nueva">
            <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
            Nueva carta
          </Link>
        </Button>
      </div>

      <div className="flex flex-wrap gap-2">
        {filtros.map((opcion) => (
          <Button
            key={opcion.valor}
            type="button"
            variant={filtro === opcion.valor ? "default" : "outline"}
            size="sm"
            onClick={() => setFiltro(opcion.valor)}
            className={filtro === opcion.valor ? "text-white" : ""}
          >
            {opcion.texto}
          </Button>
        ))}
      </div>

      {consulta.isLoading ? (
        <div className="space-y-3">
          <Skeleton className="h-16 w-full" />
          <Skeleton className="h-16 w-full" />
        </div>
      ) : piezas.length === 0 ? (
        <Card className="border-dashed">
          <CardContent className="py-12 text-center">
            <FileText className="mx-auto h-8 w-8 text-muted-foreground" aria-hidden="true" />
            <p className="mt-3 text-sm font-medium">
              {filtro === "todas"
                ? "Todavía no hay correspondencia"
                : "No hay cartas en este estado"}
            </p>
            <p className="mt-1 text-sm text-muted-foreground">
              {filtro === "todas"
                ? "Redacte una carta, elija la lista de destinatarios y envíela."
                : "Pruebe con otro filtro."}
            </p>
            {filtro === "todas" && (
              <Button asChild className="mt-4">
                <Link href="/correspondencia/nueva">
                  <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
                  Nueva carta
                </Link>
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardContent className="p-0">
            <ul className="divide-y divide-zinc-100">
              {piezas.map((pieza) => (
                <li key={pieza.id}>
                  <Link
                    href={`/correspondencia/${pieza.id}`}
                    className="flex items-center justify-between gap-4 px-5 py-4 transition-colors hover:bg-zinc-50"
                  >
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium">{pieza.asunto}</p>
                      <p className="mt-0.5 text-xs text-muted-foreground">
                        Actualizada el {new Date(pieza.actualizado_en).toLocaleString("es-PE")}
                        {pieza.adjuntos.length > 0 &&
                          ` · ${pieza.adjuntos.length} adjunto${pieza.adjuntos.length > 1 ? "s" : ""}`}
                      </p>
                    </div>
                    <Badge
                      variant="secondary"
                      className={colorEstadoCorrespondencia[pieza.estado]}
                    >
                      {etiquetaEstadoCorrespondencia[pieza.estado]}
                    </Badge>
                  </Link>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}
    </section>
  );
}
