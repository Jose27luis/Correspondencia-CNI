"use client";

import { useMutation } from "@tanstack/react-query";
import { ClipboardPaste, Loader2, Sparkles, TriangleAlert, X } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { api } from "@/lib/api";
import { ErrorPeticion } from "@/lib/api-client";
import type { ContactoExtraido } from "@/lib/tipos";

function aCSV(contactos: ContactoExtraido[]): File {
  const escapar = (valor: string) => `"${valor.replaceAll('"', '""')}"`;

  const lineas = [
    "nombre,empresa,correo,pais,cargo",
    ...contactos.map((contacto) =>
      [contacto.nombre, contacto.empresa, contacto.correo, contacto.pais, contacto.cargo]
        .map(escapar)
        .join(","),
    ),
  ];

  return new File([`﻿${lineas.join("\n")}\n`], "contactos-extraidos.csv", {
    type: "text/csv",
  });
}

interface Props {
  abierto: boolean;
  alCambiar: (abierto: boolean) => void;
  alImportar: () => void;
}

export function PegarContactos({ abierto, alCambiar, alImportar }: Props) {
  const [texto, setTexto] = useState("");
  const [extraidos, setExtraidos] = useState<ContactoExtraido[]>([]);
  const [aviso, setAviso] = useState("");

  function reiniciar() {
    setTexto("");
    setExtraidos([]);
    setAviso("");
  }

  const extraccion = useMutation({
    mutationFn: () => api.extraerContactos(texto),
    onSuccess: (resultado) => {
      setExtraidos(resultado.contactos);
      setAviso(resultado.aviso);

      if (resultado.contactos.length === 0) {
        toast.warning("No se encontraron contactos en el texto", {
          description: resultado.aviso || "Revise que el texto incluya correos electrónicos.",
        });
      }
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo leer el texto");
    },
  });

  const importacion = useMutation({
    mutationFn: () => api.importarContactos(aCSV(extraidos)),
    onSuccess: (resumen) => {
      toast.success(
        `${resumen.creados} contactos creados, ${resumen.actualizados} actualizados`,
        resumen.omitidos > 0
          ? { description: `${resumen.omitidos} entradas omitidas por datos incompletos.` }
          : undefined,
      );
      reiniciar();
      alCambiar(false);
      alImportar();
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudieron guardar");
    },
  });

  function quitar(correo: string) {
    setExtraidos((actuales) => actuales.filter((contacto) => contacto.correo !== correo));
  }

  return (
    <Dialog
      open={abierto}
      onOpenChange={(valor) => {
        if (!valor) {
          reiniciar();
        }
        alCambiar(valor);
      }}
    >
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Sparkles className="h-4 w-4 text-emerald-700" aria-hidden="true" />
            Pegar contactos
          </DialogTitle>
          <DialogDescription>
            Pegue un correo recibido, una firma, un listado o cualquier texto con datos de empresas.
            El asistente extraerá los contactos y usted decide cuáles guardar.
          </DialogDescription>
        </DialogHeader>

        {extraidos.length === 0 ? (
          <div className="space-y-1.5">
            <Label htmlFor="texto-pegado">Texto</Label>
            <Textarea
              id="texto-pegado"
              rows={10}
              value={texto}
              onChange={(evento) => setTexto(evento.target.value)}
              className="font-mono text-xs"
            />
            <p className="text-xs text-muted-foreground">
              Solo se extrae lo que esté escrito en el texto. Las entradas sin correo se descartan.
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {aviso && (
              <p className="flex items-start gap-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
                <TriangleAlert className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
                {aviso}
              </p>
            )}

            <p className="text-sm text-muted-foreground">
              {extraidos.length} contactos encontrados. Quite los que no correspondan antes de
              guardar.
            </p>

            <ul className="max-h-72 divide-y divide-zinc-100 overflow-y-auto rounded-lg border border-zinc-200">
              {extraidos.map((contacto) => (
                <li
                  key={contacto.correo}
                  className="flex items-center justify-between gap-3 px-3 py-2.5"
                >
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">
                      {contacto.empresa || "Empresa sin nombre"}
                    </p>
                    <p className="truncate text-xs text-muted-foreground">
                      {[contacto.nombre, contacto.correo, contacto.cargo, contacto.pais]
                        .filter(Boolean)
                        .join(" · ")}
                    </p>
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => quitar(contacto.correo)}
                    aria-label={`Quitar ${contacto.correo}`}
                    className="h-8 w-8 shrink-0 text-muted-foreground hover:text-destructive"
                  >
                    <X className="h-4 w-4" aria-hidden="true" />
                  </Button>
                </li>
              ))}
            </ul>
          </div>
        )}

        <DialogFooter className="gap-2 sm:gap-2">
          <Button type="button" variant="outline" onClick={() => alCambiar(false)}>
            Cancelar
          </Button>

          {extraidos.length === 0 ? (
            <Button
              type="button"
              onClick={() => extraccion.mutate()}
              disabled={extraccion.isPending || texto.trim().length < 10}
              className="text-white"
            >
              {extraccion.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
              ) : (
                <ClipboardPaste className="mr-2 h-4 w-4" aria-hidden="true" />
              )}
              Extraer contactos
            </Button>
          ) : (
            <Button
              type="button"
              onClick={() => importacion.mutate()}
              disabled={importacion.isPending}
              className="text-white"
            >
              {importacion.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
              )}
              Guardar {extraidos.length} contactos
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
