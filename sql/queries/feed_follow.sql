-- name: CreateFeedFollow :one
INSERT INTO feedfollow (id, created_at, updated_at, user_id, feed_id) 
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteFeedFollow :exec
DELETE FROM feedfollow WHERE id = $1 and user_id = $2;

-- name: ListFeedFollows :many
SELECT * FROM feedfollow WHERE user_id = $1;