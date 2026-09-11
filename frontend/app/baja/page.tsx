"use client";

import { useMutation } from "@tanstack/react-query";
import { CircleCheck, Loader2, TriangleAlert } from "lucide-react";
import { useSearchParams } from "next/navigation";
import { Suspense } from "react";

import { Button } from "@/components/ui/button";
import { peticion, ErrorPeticion } from "@/lib/api-client";

interface ResultadoBaja {
  correo: string;
  empresa: string;
}

function ContenidoBaja() {
  const parametros = useSearchParams();
  const contacto = parametros.get("c") ?? "";
  const token = parametros.get("t") ?? "";
  const enlaceValido = contacto.length > 0 && token.length > 0;

  const baja = useMutation({
    mutationFn: () =>
      peticion<ResultadoBaja>("/baja", {
        metodo: "POST",
        parametros: { c: contacto, t: token },
      }),
  });

  if (!enlaceValido) {
    return (
      <Mensaje
        icono={<TriangleAlert className="h-6 w-6 text-amber-600" aria-hidden="true" />}
        titulo="Enlace incompleto"
        detalle="El enlace de baja no está completo. Use el enlace tal como aparece al final del correo."
      />
    );
  }

  if (baja.isSuccess) {
    return (
      <Mensaje
        icono={<CircleCheck className="h-6 w-6 text-emerald-600" aria-hidden="true" />}
        titulo="Solicitud registrada"
        detalle={`La dirección ${baja.data.correo} ya no recibirá más comunicaciones comerciales de CNI.`}
      />
    );
  }

  return (
    <div className="space-y-5">
      <div>
        <h1 className="text-xl font-semibold tracking-tight">Dejar de recibir correos de CNI</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Confirme que no desea recibir más comunicaciones comerciales de CNI – Corporación de
          Negocios Interoceánicos. Puede volver a solicitarlas cuando lo desee escribiendo a su
          ejecutiva comercial.
        </p>
      </div>

      {baja.isError && (
        <p
          role="alert"
          className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700"
        >
          {baja.error instanceof ErrorPeticion
            ? baja.error.message
            : "No se pudo registrar la solicitud. Intente nuevamente."}
        </p>
      )}

      <Button
        type="button"
        onClick={() => baja.mutate()}
        disabled={baja.isPending}
        className="w-full text-white"
      >
        {baja.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />}
        Confirmar la baja
      </Button>
    </div>
  );
}

function Mensaje({
  icono,
  titulo,
  detalle,
}: {
  icono: React.ReactNode;
  titulo: string;
  detalle: string;
}) {
  return (
    <div className="space-y-3 text-center">
      <div className="flex justify-center">{icono}</div>
      <h1 className="text-xl font-semibold tracking-tight">{titulo}</h1>
      <p className="text-sm text-muted-foreground">{detalle}</p>
    </div>
  );
}

export default function PaginaBaja() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-zinc-50 px-6 py-12">
      <div className="w-full max-w-md rounded-2xl border border-zinc-200/80 bg-white p-8 shadow-[0_1px_2px_rgba(16,24,40,0.04),0_12px_32px_-8px_rgba(16,24,40,0.12)]">
        <p className="mb-6 text-sm font-semibold">
          Correspondencia <span className="text-emerald-600">CNI</span>
        </p>
        <Suspense fallback={<p className="text-sm text-muted-foreground">Cargando</p>}>
          <ContenidoBaja />
        </Suspense>
      </div>
    </main>
  );
}
