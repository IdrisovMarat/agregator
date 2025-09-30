-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds
ORDER BY created_at DESC;


-- name: GetFeedsByUser :many
SELECT * FROM feeds
WHERE user_id = $1
ORDER BY created_at DESC;