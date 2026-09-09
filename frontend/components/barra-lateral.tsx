"use client";

import { motion } from "framer-motion";
import { FileText, LayoutDashboard, ListChecks, LogOut, Send, Users } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { useSesion } from "@/lib/sesion";
import { cn } from "@/lib/utils";

interface Seccion {
  titulo: string;
  enlaces: Enlace[];
}

interface Enlace {
  href: string;
  texto: string;
  icono: typeof LayoutDashboard;
  exacto?: boolean;
}

const secciones: Seccion[] = [
  {
    titulo: "General",
    enlaces: [{ href: "/inicio", texto: "Inicio", icono: LayoutDashboard, exacto: true }],
  },
  {
    titulo: "Destinatarios",
    enlaces: [
      { href: "/contactos", texto: "Contactos", icono: Users },
      { href: "/listas", texto: "Listas", icono: ListChecks },
    ],
  },
  {
    titulo: "Envíos",
    enlaces: [
      { href: "/correspondencia", texto: "Correspondencia", icono: FileText },
      { href: "/correspondencia/nueva", texto: "Nueva carta", icono: Send, exacto: true },
    ],
  },
];

export function BarraLateral({ alNavegar }: { alNavegar?: () => void }) {
  const ruta = usePathname();
  const { usuario, salir } = useSesion();

  function estaActivo(enlace: Enlace): boolean {
    if (enlace.exacto) {
      return ruta === enlace.href;
    }
    return ruta === enlace.href || ruta.startsWith(`${enlace.href}/`);
  }

  return (
    <div className="flex h-full flex-col bg-zinc-950 text-zinc-300">
      <div className="px-5 py-6">
        <Link href="/inicio" onClick={alNavegar} className="block">
          <span className="text-base font-semibold text-white">Correspondencia</span>
          <span className="block text-sm font-semibold text-emerald-400">CNI</span>
        </Link>
      </div>

      <nav className="flex-1 space-y-6 overflow-y-auto px-3">
        {secciones.map((seccion) => (
          <div key={seccion.titulo}>
            <p className="px-3 pb-2 text-[0.7rem] font-medium uppercase tracking-wider text-zinc-500">
              {seccion.titulo}
            </p>
            <ul className="space-y-1">
              {seccion.enlaces.map((enlace) => {
                const activo = estaActivo(enlace);
                const Icono = enlace.icono;

                return (
                  <li key={enlace.href}>
                    <Link
                      href={enlace.href}
                      onClick={alNavegar}
                      aria-current={activo ? "page" : undefined}
                      className={cn(
                        "relative flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors",
                        activo ? "text-white" : "text-zinc-400 hover:bg-white/5 hover:text-white",
                      )}
                    >
                      {activo && (
                        <motion.span
                          layoutId="indicador-navegacion"
                          className="absolute inset-0 rounded-lg bg-white/10"
                          transition={{ type: "spring", stiffness: 380, damping: 32 }}
                        />
                      )}
                      <Icono className="relative h-4 w-4 shrink-0" aria-hidden="true" />
                      <span className="relative">{enlace.texto}</span>
                    </Link>
                  </li>
                );
              })}
            </ul>
          </div>
        ))}
      </nav>

      <div className="px-3 pb-5">
        <Separator className="mb-4 bg-white/10" />
        <div className="flex items-center justify-between gap-2 px-2">
          <div className="min-w-0">
            <p className="truncate text-sm text-white">{usuario?.nombre}</p>
            <p className="truncate text-xs text-zinc-500">{usuario?.correo}</p>
          </div>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={salir}
            aria-label="Cerrar sesión"
            title="Cerrar sesión"
            className="h-8 w-8 shrink-0 text-zinc-400 hover:bg-white/10 hover:text-white"
          >
            <LogOut className="h-4 w-4" aria-hidden="true" />
          </Button>
        </div>
      </div>
    </div>
  );
}
