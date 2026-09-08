# Panel de Correspondencia CNI

Sistema interno para reemplazar el envío manual de cartas comerciales (Word + correspondencia) por un panel web que permite componer, personalizar y enviar correspondencia masiva a empresas a través de Resend, con seguimiento de entregas.

---

## 1. Descripción general

CNI (Coorporación de Negocios Interoceánicos) envía actualmente cartas comerciales (ofertas de maíz, transporte, logística) usando el módulo de correspondencia de Word combinado con listas en Excel. Este proceso es manual, no deja registro de qué empresas recibieron qué oferta, y no permite saber si el correo fue entregado o abierto.

Este proyecto reemplaza ese flujo por un **panel interno** donde el equipo de CNI puede:

- Importar y mantener una base de contactos de empresas
- Componer correspondencia (asunto, cuerpo, variables de personalización y adjuntos)
- Enviar de forma masiva a través de **Resend**
- Ver el estado de cada envío (entregado, rebotado, abierto)

---

## 2. Objetivos del proyecto

| Objetivo | Descripción |
|---|---|
| Centralizar contactos | Una sola fuente de verdad para empresas y correos, reemplazando el Excel disperso |
| Agilizar la composición | Redactar una vez y enviar a muchas empresas, sin repetir el proceso de Word |
| Personalizar sin esfuerzo | Variables tipo `{nombre}`, `{empresa}` que se reemplazan automáticamente |
| Visibilidad de resultados | Saber qué se entregó, qué rebotó y qué se abrió |
| Profesionalizar el envío | Dominio verificado (SPF/DKIM/DMARC) para evitar spam |

---

## 3. Alcance

**Incluye (fase 1):**
- Gestión de contactos y listas/segmentos
- Composición de correspondencia con variables y adjunto (PDF/Word ya generado)
- Envío masivo vía Resend con control de límites
- Registro de eventos de envío vía webhooks

**Fuera de alcance (por ahora):**
- Generación automática de PDF personalizado por destinatario (fase 2)
- Editor visual de plantillas HTML tipo boletín (fase 2)
- Respuestas entrantes / bandeja de entrada compartida

---

## 4. Stack tecnológico

| Capa | Tecnología | Motivo |
|---|---|---|
| Frontend + Backend | **Next.js 14 (App Router) + TypeScript** | Un solo proyecto, SSR, API routes, fácil despliegue |
| UI | **Tailwind CSS + shadcn/ui** | Componentes accesibles y consistentes, desarrollo rápido |
| Base de datos | **PostgreSQL (Supabase)** | Relacional, robusto, con panel de administración incluido |
| ORM | **Prisma** | Tipado seguro end-to-end con TypeScript |
| Envío de correo | **Resend + React Email** | API simple, soporte de adjuntos, webhooks de estado |
| Cola de envío | **Upstash QStash** (o cola propia con rate-limit) | Evita saturar el límite de envíos de Resend |
| Autenticación | **NextAuth (Auth.js)** | Acceso restringido solo al equipo de CNI |
| Almacenamiento de adjuntos | **Supabase Storage** | Guardar PDFs/Word subidos antes de enviarlos |
| Validación | **Zod** | Validación de formularios y payloads de API |
| Hosting | **Vercel** (app) + **Supabase** (DB/Storage) | Despliegue continuo, escalabilidad, bajo mantenimiento |

---

## 5. Casos de uso

**Actor principal:** Usuario del equipo de CNI (rol único por ahora, ampliable a roles después)

```mermaid
flowchart LR
    Usuario(["Usuario CNI"])

    UC1(("Importar
contactos"))
    UC2(("Crear / editar
lista de contactos"))
    UC3(("Componer
correspondencia"))
    UC4(("Adjuntar
documento"))
    UC5(("Previsualizar
correo personalizado"))
    UC6(("Enviar
correspondencia masiva"))
    UC7(("Consultar estado
de envíos"))
    UC8(("Reenviar a
rebotados"))

    Usuario --- UC1
    Usuario --- UC2
    Usuario --- UC3
    Usuario --- UC4
    Usuario --- UC5
    Usuario --- UC6
    Usuario --- UC7
    Usuario --- UC8

    UC3 -.incluye.-> UC4
    UC6 -.requiere.-> UC5
    UC8 -.depende de.-> UC7

    classDef actor fill:#eef2ff,stroke:#4f46e5,stroke-width:2px,color:#1e1b4b
    classDef uc fill:#ecfdf5,stroke:#059669,stroke-width:2px,color:#064e3b
    class Usuario actor
    class UC1,UC2,UC3,UC4,UC5,UC6,UC7,UC8 uc
```

### Descripción de casos de uso clave

| Caso de uso | Precondición | Resultado |
|---|---|---|
| Importar contactos | Archivo CSV/Excel con columnas nombre, empresa, correo | Contactos creados o actualizados en la base de datos |
| Componer correspondencia | Al menos una lista de contactos existente | Borrador de correspondencia guardado |
| Enviar correspondencia masiva | Correspondencia con asunto, cuerpo y lista seleccionada | Correos encolados y enviados vía Resend |
| Consultar estado de envíos | Al menos un envío realizado | Vista con entregados, rebotados y abiertos por envío |

---

## 6. Diagrama de clases (modelo de dominio)

```mermaid
classDiagram
    class Usuario {
        +string id
        +string nombre
        +string correo
        +string passwordHash
        +login()
    }

    class Contacto {
        +string id
        +string nombre
        +string empresa
        +string correo
        +string pais
        +json camposExtra
        +actualizar()
    }

    class ListaContactos {
        +string id
        +string nombre
        +string descripcion
        +agregarContacto(contacto)
        +quitarContacto(contacto)
    }

    class Correspondencia {
        +string id
        +string asunto
        +string cuerpo
        +string estado
        +crearBorrador()
        +previsualizar(contacto)
    }

    class Adjunto {
        +string id
        +string nombreArchivo
        +string urlArchivo
        +string tipo
    }

    class Envio {
        +string id
        +datetime fechaEnvio
        +string estado
        +enviar()
    }

    class EventoEnvio {
        +string id
        +string tipo
        +datetime fecha
        +json payload
    }

    Usuario "1" --> "many" Correspondencia : crea
    Correspondencia "1" --> "many" Adjunto : contiene
    Correspondencia "1" --> "1" ListaContactos : se envía a
    ListaContactos "many" --> "many" Contacto : agrupa
    Correspondencia "1" --> "many" Envio : genera
    Envio "many" --> "1" Contacto : destinado a
    Envio "1" --> "many" EventoEnvio : registra
```

---

## 7. Diagrama de secuencia (flujo de envío masivo)

```mermaid
sequenceDiagram
    actor U as Usuario CNI
    participant P as Panel Web (Next.js)
    participant DB as Base de Datos
    participant Q as Cola de Envío
    participant R as Resend API
    participant E as Empresa Destinataria

    U->>P: Selecciona lista y compone correspondencia
    P->>DB: Guarda borrador de correspondencia
    U->>P: Confirma envío masivo
    P->>DB: Obtiene contactos de la lista
    P->>Q: Encola un envío por contacto
    loop Por cada contacto en la cola
        Q->>R: Envía correo personalizado con adjunto
        R->>E: Entrega el correo
        R-->>Q: Confirma aceptación del envío
    end
    R-->>P: Webhook: entregado / rebotado / abierto
    P->>DB: Registra evento de envío
    U->>P: Consulta dashboard de resultados
    P->>DB: Obtiene estado de todos los envíos
    P-->>U: Muestra entregados, rebotados y abiertos
```

---

## 8. Diagrama entidad-relación

```mermaid
erDiagram
    USUARIO ||--o{ CORRESPONDENCIA : crea
    LISTA_CONTACTOS ||--o{ CONTACTO_LISTA : agrupa
    CONTACTO ||--o{ CONTACTO_LISTA : pertenece
    CORRESPONDENCIA ||--o{ ADJUNTO : incluye
    CORRESPONDENCIA ||--|| LISTA_CONTACTOS : dirigido_a
    CORRESPONDENCIA ||--o{ ENVIO : genera
    CONTACTO ||--o{ ENVIO : recibe
    ENVIO ||--o{ EVENTO_ENVIO : registra

    USUARIO {
        uuid id PK
        string nombre
        string correo
        string password_hash
        datetime creado_en
    }

    CONTACTO {
        uuid id PK
        string nombre
        string empresa
        string correo
        string pais
        jsonb campos_extra
        datetime creado_en
    }

    LISTA_CONTACTOS {
        uuid id PK
        string nombre
        string descripcion
    }

    CONTACTO_LISTA {
        uuid contacto_id FK
        uuid lista_id FK
    }

    CORRESPONDENCIA {
        uuid id PK
        uuid usuario_id FK
        string asunto
        text cuerpo
        string estado
        uuid lista_id FK
        datetime creado_en
    }

    ADJUNTO {
        uuid id PK
        uuid correspondencia_id FK
        string nombre_archivo
        string url_archivo
        string tipo
    }

    ENVIO {
        uuid id PK
        uuid correspondencia_id FK
        uuid contacto_id FK
        string estado
        datetime fecha_envio
    }

    EVENTO_ENVIO {
        uuid id PK
        uuid envio_id FK
        string tipo
        datetime fecha
        jsonb payload
    }
```

---

## 9. Diagrama de despliegue

```mermaid
flowchart TB
    subgraph CLIENTE["Cliente"]
        Browser["Navegador
Usuario CNI"]
    end

    subgraph VERCEL["Vercel"]
        NextApp["Next.js App
Frontend + API Routes"]
    end

    subgraph SUPABASE["Supabase"]
        PG[("PostgreSQL
Base de datos")]
        Storage["Storage
Adjuntos PDF/Word"]
    end

    subgraph UPSTASH["Upstash"]
        Queue["QStash
Cola de envío"]
    end

    subgraph RESEND["Resend"]
        API["Resend API"]
        Webhook["Webhooks
de estado"]
    end

    subgraph DOMINIO["Dominio verificado"]
        DNS["SPF / DKIM / DMARC"]
    end

    Browser -->|HTTPS| NextApp
    NextApp -->|Prisma| PG
    NextApp -->|Sube archivo| Storage
    NextApp -->|Encola envíos| Queue
    Queue -->|Dispara envío| API
    API -->|Usa registros de| DNS
    API -->|Entrega correo| Empresas["Empresas
destinatarias"]
    API -->|Eventos| Webhook
    Webhook -->|Actualiza estado| NextApp

    classDef cliente fill:#eef2ff,stroke:#4f46e5,stroke-width:2px,color:#1e1b4b
    classDef app fill:#ecfdf5,stroke:#059669,stroke-width:2px,color:#064e3b
    classDef datos fill:#fff7ed,stroke:#ea580c,stroke-width:2px,color:#7c2d12
    classDef envio fill:#fdf2f8,stroke:#db2777,stroke-width:2px,color:#831843

    class Browser cliente
    class NextApp app
    class PG,Storage,Queue datos
    class API,Webhook,DNS envio
```

---

## 10. Estructura de carpetas propuesta

```
panel-correspondencia-cni/
├── app/
│   ├── (auth)/
│   │   └── login/
│   ├── contactos/
│   ├── correspondencia/
│   │   ├── nueva/
│   │   └── [id]/
│   ├── envios/
│   └── api/
│       ├── contactos/
│       ├── correspondencia/
│       ├── envios/
│       └── webhooks/resend/
├── components/
├── lib/
│   ├── resend.ts
│   ├── prisma.ts
│   └── queue.ts
├── prisma/
│   └── schema.prisma
└── emails/
    └── plantilla-correspondencia.tsx
```

---

## 11. Variables de entorno

```
DATABASE_URL=
RESEND_API_KEY=
RESEND_WEBHOOK_SECRET=
NEXTAUTH_SECRET=
NEXTAUTH_URL=
SUPABASE_URL=
SUPABASE_SERVICE_ROLE_KEY=
QSTASH_TOKEN=
```

---

## 12. Roadmap por fases

| Fase | Contenido |
|---|---|
| **Fase 1 (MVP)** | Contactos, listas, composición con adjunto manual, envío vía Resend, webhook básico |
| **Fase 2** | Generación automática de PDF/Word personalizado por destinatario |
| **Fase 3** | Plantillas HTML tipo boletín, editor visual |
| **Fase 4** | Roles de usuario, historial avanzado, reenvío automático a rebotados |

---

## 13. Consideraciones de seguridad

- Acceso al panel restringido por autenticación (solo equipo CNI)
- Verificación de dominio (SPF/DKIM/DMARC) antes de enviar en producción
- Validación de archivos adjuntos (tipo y tamaño) antes de subir a Storage
- Rate-limiting en el envío para respetar límites de Resend y evitar marcarse como spam
- Registro de auditoría de quién envió qué correspondencia y cuándo
