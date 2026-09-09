-- name: CrearEnviosDeLista :execrows
INSERT INTO envio (correspondencia_id, contacto_id)
SELECT $1, cl.contacto_id
FROM contacto_lista cl
WHERE cl.lista_id = $2
ON CONFLICT (correspondencia_id, contacto_id) DO NOTHING;

-- name: ListarEnviosPendientes :many
SELECT e.id, e.correspondencia_id, e.contacto_id
FROM envio e
WHERE e.correspondencia_id = $1 AND e.estado = 'pendiente'
ORDER BY e.creado_en;

-- name: ObtenerEnvioConDatos :one
SELECT
    e.id,
    e.correspondencia_id,
    e.estado,
    c.id AS contacto_id,
    c.nombre,
    c.empresa,
    c.correo,
    c.pais,
    c.campos_extra,
    c.creado_en AS contacto_creado_en,
    c.actualizado_en AS contacto_actualizado_en
FROM envio e
JOIN contacto c ON c.id = e.contacto_id
WHERE e.id = $1;

-- name: MarcarEnvioEnviado :execrows
UPDATE envio
SET estado = 'enviado',
    proveedor_mensaje_id = $2,
    fecha_envio = now(),
    actualizado_en = now()
WHERE id = $1;

-- name: MarcarEnvioFallido :execrows
UPDATE envio
SET estado = 'fallido',
    actualizado_en = now()
WHERE id = $1;

-- name: ActualizarEstadoPorMensaje :one
UPDATE envio
SET estado = $2,
    actualizado_en = now()
WHERE proveedor_mensaje_id = $1
RETURNING id;

-- name: RegistrarEventoEnvio :one
INSERT INTO evento_envio (envio_id, tipo, fecha, payload)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListarEnviosDeCorrespondencia :many
SELECT
    e.id,
    e.estado,
    e.proveedor_mensaje_id,
    e.fecha_envio,
    c.nombre,
    c.empresa,
    c.correo
FROM envio e
JOIN contacto c ON c.id = e.contacto_id
WHERE e.correspondencia_id = $1
  AND (sqlc.narg('estado')::text IS NULL OR e.estado = sqlc.narg('estado')::text)
ORDER BY c.empresa, c.nombre
LIMIT $2 OFFSET $3;

-- name: ContarEnviosDeCorrespondencia :one
SELECT count(*) FROM envio
WHERE correspondencia_id = $1
  AND (sqlc.narg('estado')::text IS NULL OR estado = sqlc.narg('estado')::text);

-- name: ResumenEnvios :many
SELECT estado, count(*) AS total
FROM envio
WHERE correspondencia_id = $1
GROUP BY estado;

-- name: ListarEventosDeEnvio :many
SELECT * FROM evento_envio WHERE envio_id = $1 ORDER BY fecha DESC;

-- name: ObtenerEnvioPorMensaje :one
SELECT id FROM envio WHERE proveedor_mensaje_id = $1;

-- name: ContarPendientesDeCorrespondencia :one
SELECT count(*) FROM envio
WHERE correspondencia_id = $1 AND estado = 'pendiente';

-- name: CerrarCorrespondenciaSiTermino :one
UPDATE correspondencia c
SET estado = CASE
        WHEN NOT EXISTS (
            SELECT 1 FROM envio e
            WHERE e.correspondencia_id = c.id AND e.estado <> 'fallido'
        ) THEN 'fallida'
        ELSE 'enviada'
    END,
    actualizado_en = now()
WHERE c.id = $1
  AND c.estado IN ('encolada', 'enviando')
  AND NOT EXISTS (
      SELECT 1 FROM envio e
      WHERE e.correspondencia_id = c.id AND e.estado = 'pendiente'
  )
RETURNING c.estado;
