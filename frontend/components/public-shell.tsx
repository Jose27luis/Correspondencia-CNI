import Link from "next/link";

import { campoTrampa } from "@/lib/intake";

export function MarcoPublico({
  titulo,
  descripcion,
  children,
}: {
  titulo: string;
  descripcion: string;
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen bg-zinc-50">
      <header className="border-b border-zinc-200 bg-white">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-3 px-6 py-4">
          <p className="text-sm font-semibold">
            Corporación de Negocios Interoceánicos <span className="text-emerald-600">CNI</span>
          </p>
          <nav className="flex gap-4 text-sm">
            <Link href="/directorio" className="text-zinc-600 hover:text-zinc-950">
              Directorio C-E
            </Link>
            <Link href="/rueda-de-negocios" className="text-zinc-600 hover:text-zinc-950">
              Rueda de Negocios
            </Link>
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-6xl space-y-8 px-6 py-10">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight">{titulo}</h1>
          <p className="mt-2 max-w-3xl text-muted-foreground">{descripcion}</p>
        </div>
        {children}
      </main>
    </div>
  );
}

export function CampoTrampa() {
  return (
    <div aria-hidden="true" className="absolute -left-[9999px] h-px w-px overflow-hidden">
      <label htmlFor={campoTrampa}>No completar</label>
      <input id={campoTrampa} name={campoTrampa} type="text" tabIndex={-1} autoComplete="off" />
    </div>
  );
}
