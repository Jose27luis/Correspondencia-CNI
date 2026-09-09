"use client";

import { Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface Props {
  abierto: boolean;
  titulo: string;
  descripcion: string;
  textoConfirmar?: string;
  procesando?: boolean;
  alCambiar: (abierto: boolean) => void;
  alConfirmar: () => void;
}

export function DialogoConfirmacion({
  abierto,
  titulo,
  descripcion,
  textoConfirmar = "Eliminar",
  procesando = false,
  alCambiar,
  alConfirmar,
}: Props) {
  return (
    <Dialog open={abierto} onOpenChange={alCambiar}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{titulo}</DialogTitle>
          <DialogDescription>{descripcion}</DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2 sm:gap-2">
          <Button type="button" variant="outline" onClick={() => alCambiar(false)}>
            Cancelar
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={alConfirmar}
            disabled={procesando}
            className="text-white"
          >
            {procesando && <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />}
            {textoConfirmar}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
