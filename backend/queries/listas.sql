-- name: CrearLista :one
INSERT INTO lista_contactos (nombre, descripcion)
VALUES ($1, $2)
RETURNING *;

-- name: ObtenerLista :one
SELECT l.*, count(cl.contacto_id) AS total_contactos
FROM lista_contactos l
LEFT JOIN contacto_lista cl ON cl.lista_id = l.id
WHERE l.id = $1
GROUP BY l.id;

-- name: ListarListas :many
SELECT l.*, count(cl.contacto_id) AS total_contactos
FROM lista_contactos l
LEFT JOIN contacto_lista cl ON cl.lista_id = l.id
GROUP BY l.id
ORDER BY l.creado_en DESC
LIMIT $1 OFFSET $2;

-- name: ContarListas :one
SELECT count(*) FROM lista_contactos;

-- name: ActualizarLista :one
UPDATE lista_contactos
SET nombre = $2,
    descripcion = $3
WHERE id = $1
RETURNING *;

-- name: EliminarLista :execrows
DELETE FROM lista_contactos WHERE id = $1;

-- name: AgregarContactosALista :execrows
INSERT INTO contacto_lista (lista_id, contacto_id)
SELECT $1, unnest(@contacto_ids::uuid[])
ON CONFLICT DO NOTHING;

-- name: QuitarContactoDeLista :execrows
DELETE FROM contacto_lista WHERE lista_id = $1 AND contacto_id = $2;

-- name: ListarContactosDeLista :many
SELECT c.*
FROM contacto c
JOIN contacto_lista cl ON cl.contacto_id = c.id
WHERE cl.lista_id = $1
ORDER BY c.empresa, c.nombre
LIMIT $2 OFFSET $3;

-- name: ContarContactosDeLista :one
SELECT count(*) FROM contacto_lista WHERE lista_id = $1;

-- name: ExisteLista :one
SELECT EXISTS (SELECT 1 FROM lista_contactos WHERE id = $1);
