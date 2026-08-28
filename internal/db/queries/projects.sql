-- name: Ping :one
SELECT 1;

-- name: ListProjects :many
SELECT * FROM projects
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetProject :one
SELECT * FROM projects
WHERE id = $1 AND user_id = $2;

-- name: CreateProject :one
INSERT INTO projects (user_id, name, slug, status, stack, summary)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateProject :one
UPDATE projects
SET name = $3,
    slug = $4,
    status = $5,
    stack = $6,
    summary = $7,
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteProject :execrows
DELETE FROM projects
WHERE id = $1 AND user_id = $2;

-- name: GetProjectByIngest :one
SELECT * FROM projects
WHERE id = $1 AND feedback_ingest_key = $2;

-- name: UpdateFeedbackOrigin :one
UPDATE projects
SET feedback_origin = $3,
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: RotateFeedbackIngestKey :one
UPDATE projects
SET feedback_ingest_key = replace(gen_random_uuid()::text, '-', ''),
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;
