-- name: CreatePost :one
INSERT INTO posts (title, url, description, published_at, feed_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPostsForUser :many
SELECT 
    posts.*,
    feeds.name as feed_name
FROM posts
INNER JOIN feeds ON posts.feed_id = feeds.id
INNER JOIN feed_follows ON feeds.id = feed_follows.feed_id
WHERE feed_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2;

-- name: GetPostByURL :one
SELECT * FROM posts WHERE url = $1;

-- name: GetPostsByFeed :many
SELECT * FROM posts 
WHERE feed_id = $1 
ORDER BY published_at DESC;