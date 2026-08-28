-- name: ListProjectFeedback :many
SELECT * FROM feedback
WHERE project_id = $1 AND user_id = $2
ORDER BY received_at DESC;

-- name: CreateFeedback :one
INSERT INTO feedback (user_id, project_id, author_name, author_email, message, rating, source)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: DeleteFeedback :execrows
DELETE FROM feedback
WHERE id = $1 AND project_id = $2 AND user_id = $3;
