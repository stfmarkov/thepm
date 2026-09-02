-- name: ListProjectNotes :many
SELECT * FROM notes
WHERE project_id = $1 AND user_id = $2
ORDER BY created_at DESC;

-- name: CreateNote :one
INSERT INTO notes (user_id, project_id, body)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateNote :one
UPDATE notes
SET body = $4
WHERE id = $1 AND project_id = $2 AND user_id = $3
RETURNING *;

-- name: DeleteNote :execrows
DELETE FROM notes
WHERE id = $1 AND project_id = $2 AND user_id = $3;
