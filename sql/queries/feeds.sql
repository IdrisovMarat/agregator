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

-- name: GetFeedIdByUrl :one
SELECT id FROM feeds
WHERE url = $1;

-- name: GetFeedsWithName :many
SELECT feeds.name, feeds.url, feeds.created_at, users.name
FROM feeds
RIGHT JOIN users
ON feeds.user_id = users.id
WHERE feeds.user_id = users.id
ORDER BY feeds.created_at DESC;

-- name: MarkFeedFetched :exec
UPDATE feeds 
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds 
ORDER BY last_fetched_at NULLS FIRST, created_at ASC
LIMIT 1;
