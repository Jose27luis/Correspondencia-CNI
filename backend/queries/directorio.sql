-- name: CrearDirectorioEmpresa :one
INSERT INTO directorio_empresa (
    nombre, ruc, correo, direccion, ciudad, telefono, celular,
    facebook, pagina_web, descripcion, logo_url
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: ObtenerDirectorioEmpresa :one
SELECT * FROM directorio_empresa WHERE id = $1;

-- name: ListarDirectorio :many
SELECT * FROM directorio_empresa
WHERE (sqlc.narg('estado')::text IS NULL OR estado = sqlc.narg('estado')::text)
ORDER BY creado_en DESC
LIMIT $1 OFFSET $2;

-- name: ContarDirectorio :one
SELECT count(*) FROM directorio_empresa
WHERE (sqlc.narg('estado')::text IS NULL OR estado = sqlc.narg('estado')::text);

-- name: ListarDirectorioPublico :many
SELECT * FROM directorio_empresa
WHERE estado = 'aprobado'
ORDER BY nombre;

-- name: AprobarDirectorioEmpresa :one
UPDATE directorio_empresa
SET estado = 'aprobado',
    contacto_id = $2,
    motivo_rechazo = NULL,
    revisado_en = now()
WHERE id = $1
RETURNING *;

-- name: RechazarDirectorioEmpresa :one
UPDATE directorio_empresa
SET estado = 'rechazado',
    motivo_rechazo = $2,
    revisado_en = now()
WHERE id = $1
RETURNING *;

-- name: ObtenerListaPorNombre :one
SELECT * FROM lista_contactos WHERE nombre = $1 ORDER BY creado_en LIMIT 1;
