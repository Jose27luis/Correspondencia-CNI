"use client";

import { Loader2 } from "lucide-react";
import { useState } from "react";

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

export function DialogoRechazo({
  abierto,
  nombre,
  pendiente,
  alCerrar,
  alConfirmar,
}: {
  abierto: boolean;
  nombre: string;
  pendiente: boolean;
  alCerrar: () => void;
  alConfirmar: (motivo: string) => void;
}) {
  const [motivo, setMotivo] = useState("");

  return (
    <Dialog
      open={abierto}
      onOpenChange={(valor) => {
        if (!valor) {
          setMotivo("");
          alCerrar();
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Rechazar</DialogTitle>
          <DialogDescription>{nombre} no se publicará. El motivo queda registrado.</DialogDescription>
        </DialogHeader>
        <div className="space-y-1.5">
          <Label htmlFor="motivo-rechazo">Motivo</Label>
          <Textarea
            id="motivo-rechazo"
            value={motivo}
            onChange={(evento) => setMotivo(evento.target.value)}
            rows={3}
          />
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={alCerrar}>
            Cancelar
          </Button>
          <Button
            type="button"
            variant="destructive"
            disabled={pendiente || motivo.trim().length < 3}
            onClick={() => alConfirmar(motivo.trim())}
          >
            {pendiente && <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />}
            Rechazar
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
