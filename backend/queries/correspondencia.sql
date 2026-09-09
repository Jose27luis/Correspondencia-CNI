-- name: CrearCorrespondencia :one
INSERT INTO correspondencia (usuario_id, lista_id, asunto, cuerpo)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ObtenerCorrespondencia :one
SELECT * FROM correspondencia WHERE id = $1;

-- name: ListarCorrespondencia :many
SELECT * FROM correspondencia
WHERE (sqlc.narg('estado')::text IS NULL OR estado = sqlc.narg('estado')::text)
ORDER BY creado_en DESC
LIMIT $1 OFFSET $2;

-- name: ContarCorrespondencia :one
SELECT count(*) FROM correspondencia
WHERE (sqlc.narg('estado')::text IS NULL OR estado = sqlc.narg('estado')::text);

-- name: ActualizarCorrespondencia :one
UPDATE correspondencia
SET lista_id = $2,
    asunto = $3,
    cuerpo = $4,
    actualizado_en = now()
WHERE id = $1 AND estado = 'borrador'
RETURNING *;

-- name: CambiarEstadoCorrespondencia :one
UPDATE correspondencia
SET estado = $2,
    actualizado_en = now()
WHERE id = $1
RETURNING *;

-- name: EliminarCorrespondencia :execrows
DELETE FROM correspondencia WHERE id = $1 AND estado = 'borrador';

-- name: CrearAdjunto :one
INSERT INTO adjunto (correspondencia_id, nombre_archivo, url_archivo, tipo, tamano_bytes)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListarAdjuntos :many
SELECT * FROM adjunto WHERE correspondencia_id = $1 ORDER BY creado_en;

-- name: ObtenerAdjunto :one
SELECT * FROM adjunto WHERE id = $1 AND correspondencia_id = $2;

-- name: EliminarAdjunto :execrows
DELETE FROM adjunto WHERE id = $1 AND correspondencia_id = $2;

-- name: SumarTamanoAdjuntos :one
SELECT coalesce(sum(tamano_bytes), 0)::bigint FROM adjunto WHERE correspondencia_id = $1;

-- name: ObtenerPrimerContactoDeLista :one
SELECT c.*
FROM contacto c
JOIN contacto_lista cl ON cl.contacto_id = c.id
WHERE cl.lista_id = $1
ORDER BY c.empresa, c.nombre
LIMIT 1;
