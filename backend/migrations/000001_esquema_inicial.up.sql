CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE usuario (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre text NOT NULL,
    correo citext NOT NULL UNIQUE,
    password_hash text NOT NULL,
    creado_en timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE contacto (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre text NOT NULL,
    empresa text NOT NULL,
    correo citext NOT NULL UNIQUE,
    pais text,
    campos_extra jsonb NOT NULL DEFAULT '{}'::jsonb,
    creado_en timestamptz NOT NULL DEFAULT now(),
    actualizado_en timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE lista_contactos (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre text NOT NULL,
    descripcion text,
    creado_en timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE contacto_lista (
    contacto_id uuid NOT NULL REFERENCES contacto (id) ON DELETE CASCADE,
    lista_id uuid NOT NULL REFERENCES lista_contactos (id) ON DELETE CASCADE,
    agregado_en timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (contacto_id, lista_id)
);

CREATE TABLE correspondencia (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id uuid NOT NULL REFERENCES usuario (id) ON DELETE RESTRICT,
    lista_id uuid REFERENCES lista_contactos (id) ON DELETE SET NULL,
    asunto text NOT NULL,
    cuerpo text NOT NULL,
    estado text NOT NULL DEFAULT 'borrador'
        CHECK (estado IN ('borrador', 'encolada', 'enviando', 'enviada', 'fallida')),
    creado_en timestamptz NOT NULL DEFAULT now(),
    actualizado_en timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE adjunto (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    correspondencia_id uuid NOT NULL REFERENCES correspondencia (id) ON DELETE CASCADE,
    nombre_archivo text NOT NULL,
    url_archivo text NOT NULL,
    tipo text NOT NULL,
    tamano_bytes bigint NOT NULL,
    creado_en timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE envio (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    correspondencia_id uuid NOT NULL REFERENCES correspondencia (id) ON DELETE CASCADE,
    contacto_id uuid NOT NULL REFERENCES contacto (id) ON DELETE RESTRICT,
    proveedor_mensaje_id text UNIQUE,
    estado text NOT NULL DEFAULT 'pendiente'
        CHECK (estado IN ('pendiente', 'enviado', 'entregado', 'rebotado', 'fallido')),
    fecha_envio timestamptz,
    creado_en timestamptz NOT NULL DEFAULT now(),
    actualizado_en timestamptz NOT NULL DEFAULT now(),
    UNIQUE (correspondencia_id, contacto_id)
);

CREATE TABLE evento_envio (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    envio_id uuid NOT NULL REFERENCES envio (id) ON DELETE CASCADE,
    tipo text NOT NULL,
    fecha timestamptz NOT NULL,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    creado_en timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_contacto_empresa ON contacto (empresa);
CREATE INDEX idx_contacto_lista_lista ON contacto_lista (lista_id);
CREATE INDEX idx_correspondencia_usuario ON correspondencia (usuario_id);
CREATE INDEX idx_correspondencia_estado ON correspondencia (estado);
CREATE INDEX idx_envio_correspondencia ON envio (correspondencia_id);
CREATE INDEX idx_envio_estado ON envio (estado);
CREATE INDEX idx_evento_envio_envio ON evento_envio (envio_id);
