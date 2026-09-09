"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState } from "react";

import { ErrorPeticion } from "@/lib/api-client";
import { ProveedorSesion } from "@/lib/sesion";

export function Providers({ children }: { children: React.ReactNode }) {
  const [cliente] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 30_000,
            retry: (intentos, error) => {
              if (error instanceof ErrorPeticion && error.codigo < 500) {
                return false;
              }
              return intentos < 2;
            },
          },
        },
      }),
  );

  return (
    <QueryClientProvider client={cliente}>
      <ProveedorSesion>{children}</ProveedorSesion>
    </QueryClientProvider>
  );
}
