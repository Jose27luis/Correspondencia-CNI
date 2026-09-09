"use client";

import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { ArrowRight, FileText, ListChecks, Send, Users } from "lucide-react";
import Link from "next/link";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { api } from "@/lib/api";
import { colorEstadoCorrespondencia } from "@/lib/estados";

export default function PaginaInicio() {
  const contactos = useQuery({
    queryKey: ["contactos", ""],
    queryFn: () => api.listarContactos(),
  });

  const listas = useQuery({ queryKey: ["listas"], queryFn: () => api.listarListas() });

  const correspondencia = useQuery({
    queryKey: ["correspondencia", undefined],
    queryFn: () => api.listarCorrespondencia(),
  });

  const piezas = correspondencia.data?.datos ?? [];
  const borradores = piezas.filter((pieza) => pieza.estado === "borrador").length;
  const enviadas = piezas.filter((pieza) => pieza.estado === "enviada").length;

  const indicadores = [
    {
      titulo: "Empresas",
      valor: contactos.data?.total,
      detalle: "en la base de contactos",
      icono: Users,
      href: "/contactos",
    },
    {
      titulo: "Listas",
      valor: listas.data?.total,
      detalle: "grupos de destinatarios",
      icono: ListChecks,
      href: "/listas",
    },
    {
      titulo: "Borradores",
      valor: borradores,
      detalle: "cartas sin enviar",
      icono: FileText,
      href: "/correspondencia",
    },
    {
      titulo: "Enviadas",
      valor: enviadas,
      detalle: "campañas completadas",
      icono: Send,
      href: "/correspondencia",
    },
  ];

  const cargando = contactos.isLoading || listas.isLoading || correspondencia.isLoading;
  const sinContactos = contactos.data?.total === 0;

  return (
    <section className="space-y-8">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Inicio</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Resumen de la correspondencia comercial de CNI.
          </p>
        </div>

        <Button asChild>
          <Link href="/correspondencia/nueva">
            Nueva carta
            <ArrowRight className="ml-2 h-4 w-4" aria-hidden="true" />
          </Link>
        </Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {indicadores.map((indicador, posicion) => {
          const Icono = indicador.icono;

          return (
            <motion.div
              key={indicador.titulo}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.35, delay: posicion * 0.06, ease: [0.22, 1, 0.36, 1] }}
            >
              <Link href={indicador.href} className="block h-full">
                <Card className="h-full transition-shadow hover:shadow-md">
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">
                      {indicador.titulo}
                    </CardTitle>
                    <Icono className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
                  </CardHeader>
                  <CardContent>
                    {cargando ? (
                      <Skeleton className="h-8 w-16" />
                    ) : (
                      <p className="text-3xl font-semibold tabular-nums">{indicador.valor ?? 0}</p>
                    )}
                    <p className="mt-1 text-xs text-muted-foreground">{indicador.detalle}</p>
                  </CardContent>
                </Card>
              </Link>
            </motion.div>
          );
        })}
      </div>

      {sinContactos && (
        <Card className="border-dashed">
          <CardHeader>
            <CardTitle className="text-base">Empiece por cargar sus empresas</CardTitle>
            <CardDescription>
              Importe el Excel de contactos como CSV con las columnas nombre, empresa y correo.
              Después agrúpelas en una lista y redacte la primera carta.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild variant="outline">
              <Link href="/contactos">Ir a contactos</Link>
            </Button>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Correspondencia reciente</CardTitle>
          <CardDescription>Las últimas cartas registradas en el panel.</CardDescription>
        </CardHeader>
        <CardContent>
          {cargando ? (
            <div className="space-y-3">
              <Skeleton className="h-12 w-full" />
              <Skeleton className="h-12 w-full" />
            </div>
          ) : piezas.length === 0 ? (
            <p className="py-6 text-center text-sm text-muted-foreground">
              Todavía no hay cartas registradas.
            </p>
          ) : (
            <ul className="divide-y divide-zinc-100">
              {piezas.slice(0, 5).map((pieza) => (
                <li key={pieza.id}>
                  <Link
                    href={`/correspondencia/${pieza.id}`}
                    className="flex items-center justify-between gap-4 py-3 transition-colors hover:text-emerald-700"
                  >
                    <span className="min-w-0">
                      <span className="block truncate text-sm font-medium">{pieza.asunto}</span>
                      <span className="block text-xs text-muted-foreground">
                        {new Date(pieza.actualizado_en).toLocaleString("es-PE")}
                      </span>
                    </span>
                    <Badge variant="secondary" className={colorEstadoCorrespondencia[pieza.estado]}>
                      {pieza.estado}
                    </Badge>
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </section>
  );
}
