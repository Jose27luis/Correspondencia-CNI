"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";

import { useSesion } from "@/lib/sesion";

const enlaces = [
  { href: "/correspondencia", texto: "Correspondencia" },
  { href: "/listas", texto: "Listas" },
  { href: "/contactos", texto: "Contactos" },
];

export default function LayoutPanel({ children }: { children: React.ReactNode }) {
  const { usuario, cargando, salir } = useSesion();
  const router = useRouter();
  const ruta = usePathname();

  useEffect(() => {
    if (!cargando && !usuario) {
      router.replace("/login");
    }
  }, [cargando, usuario, router]);

  if (cargando || !usuario) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-slate-500">Cargando</p>
      </main>
    );
  }

  return (
    <div className="min-h-screen">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
          <div className="flex items-center gap-6">
            <span className="font-semibold">Correspondencia CNI</span>
            <nav className="flex gap-4">
              {enlaces.map((enlace) => (
                <Link
                  key={enlace.href}
                  href={enlace.href}
                  className={
                    ruta.startsWith(enlace.href)
                      ? "text-sm font-medium text-slate-900"
                      : "text-sm text-slate-500 hover:text-slate-900"
                  }
                >
                  {enlace.texto}
                </Link>
              ))}
            </nav>
          </div>

          <div className="flex items-center gap-3">
            <span className="text-sm text-slate-500">{usuario.nombre}</span>
            <button
              type="button"
              onClick={salir}
              className="rounded border border-slate-300 px-3 py-1 text-sm hover:bg-slate-50"
            >
              Salir
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-4 py-8">{children}</main>
    </div>
  );
}
