"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { api } from "./api";
import { borrarToken, guardarToken, leerToken } from "./api-client";
import type { Usuario } from "./tipos";

interface EstadoSesion {
  usuario: Usuario | null;
  cargando: boolean;
  acceder: (correo: string, contrasena: string) => Promise<void>;
  salir: () => void;
}

const ContextoSesion = createContext<EstadoSesion | null>(null);

export function ProveedorSesion({ children }: { children: React.ReactNode }) {
  const [usuario, setUsuario] = useState<Usuario | null>(null);
  const [cargando, setCargando] = useState(true);
  const router = useRouter();

  useEffect(() => {
    if (!leerToken()) {
      setCargando(false);
      return;
    }

    api
      .perfil()
      .then(setUsuario)
      .catch(() => {
        borrarToken();
        setUsuario(null);
      })
      .finally(() => setCargando(false));
  }, []);

  const acceder = useCallback(
    async (correo: string, contrasena: string) => {
      const sesion = await api.acceder(correo, contrasena);
      guardarToken(sesion.token);
      setUsuario(sesion.usuario);
      router.push("/inicio");
    },
    [router],
  );

  const salir = useCallback(() => {
    borrarToken();
    setUsuario(null);
    router.push("/login");
  }, [router]);

  const valor = useMemo<EstadoSesion>(
    () => ({ usuario, cargando, acceder, salir }),
    [usuario, cargando, acceder, salir],
  );

  return <ContextoSesion.Provider value={valor}>{children}</ContextoSesion.Provider>;
}

export function useSesion(): EstadoSesion {
  const contexto = useContext(ContextoSesion);
  if (!contexto) {
    throw new Error("useSesion debe usarse dentro de ProveedorSesion");
  }
  return contexto;
}
