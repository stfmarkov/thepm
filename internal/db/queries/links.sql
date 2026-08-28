-- name: ListProjectLinks :many
SELECT * FROM links
WHERE project_id = $1 AND user_id = $2
ORDER BY created_at ASC;

-- name: CreateLink :one
INSERT INTO links (user_id, project_id, kind, url, label, notes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateLink :one
UPDATE links
SET kind = $4,
    url = $5,
    label = $6,
    notes = $7,
    updated_at = now()
WHERE id = $1 AND project_id = $2 AND user_id = $3
RETURNING *;

-- name: DeleteLink :execrows
DELETE FROM links
WHERE id = $1 AND project_id = $2 AND user_id = $3;
