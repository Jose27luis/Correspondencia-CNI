"use client";

import { AnimatePresence, motion } from "framer-motion";
import { CircleAlert, CircleCheck, TriangleAlert } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { Revision } from "@/lib/tipos";

const etiquetaVeredicto: Record<Revision["veredicto"], string> = {
  lista: "Lista para enviar",
  revisar: "Conviene revisar",
  no_enviar: "No enviar todavía",
};

const colorVeredicto: Record<Revision["veredicto"], string> = {
  lista: "bg-emerald-100 text-emerald-800 hover:bg-emerald-100",
  revisar: "bg-amber-100 text-amber-800 hover:bg-amber-100",
  no_enviar: "bg-red-100 text-red-800 hover:bg-red-100",
};

const colorGravedad: Record<Revision["hallazgos"][number]["gravedad"], string> = {
  alta: "border-red-200 bg-red-50",
  media: "border-amber-200 bg-amber-50",
  baja: "border-zinc-200 bg-zinc-50",
};

function IconoVeredicto({ veredicto }: { veredicto: Revision["veredicto"] }) {
  if (veredicto === "lista") {
    return <CircleCheck className="h-4 w-4 text-emerald-700" aria-hidden="true" />;
  }
  if (veredicto === "no_enviar") {
    return <CircleAlert className="h-4 w-4 text-red-700" aria-hidden="true" />;
  }
  return <TriangleAlert className="h-4 w-4 text-amber-700" aria-hidden="true" />;
}

export function RevisionCarta({ revision }: { revision: Revision }) {
  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3, ease: [0.22, 1, 0.36, 1] }}
      >
        <Card>
          <CardHeader>
            <div className="flex flex-wrap items-center gap-3">
              <IconoVeredicto veredicto={revision.veredicto} />
              <CardTitle className="text-base">Revisión previa</CardTitle>
              <Badge variant="secondary" className={colorVeredicto[revision.veredicto]}>
                {etiquetaVeredicto[revision.veredicto]}
              </Badge>
            </div>
            <CardDescription>{revision.resumen}</CardDescription>
          </CardHeader>

          {revision.hallazgos.length > 0 && (
            <CardContent className="space-y-3">
              {revision.hallazgos.map((hallazgo, posicion) => (
                <div
                  key={`${hallazgo.titulo}-${posicion}`}
                  className={`rounded-lg border px-4 py-3 ${colorGravedad[hallazgo.gravedad]}`}
                >
                  <p className="text-sm font-medium">{hallazgo.titulo}</p>
                  <p className="mt-1 text-sm text-zinc-700">{hallazgo.detalle}</p>
                  {hallazgo.sugerencia && (
                    <p className="mt-2 text-sm text-zinc-600">
                      <span className="font-medium">Sugerencia:</span> {hallazgo.sugerencia}
                    </p>
                  )}
                </div>
              ))}

              <p className="pt-1 text-xs text-muted-foreground">
                Revisión generada automáticamente. La decisión final de enviar es suya.
              </p>
            </CardContent>
          )}
        </Card>
      </motion.div>
    </AnimatePresence>
  );
}
