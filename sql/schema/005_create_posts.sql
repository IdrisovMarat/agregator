-- +goose Up
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    title TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    description TEXT,
    published_at TIMESTAMP,
    feed_id UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE
);

-- Индексы для оптимизации запросов
CREATE INDEX posts_feed_id_idx ON posts(feed_id);
CREATE INDEX posts_published_at_idx ON posts(published_at DESC);
CREATE INDEX posts_url_idx ON posts(url);

-- +goose Down
DROP TABLE posts;