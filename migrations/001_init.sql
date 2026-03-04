-- Up миграция
CREATE EXTENSION IF NOT EXISTS ltree;

CREATE TABLE
    IF NOT EXISTS comments (
        id SERIAL PRIMARY KEY,
        parent_id INTEGER REFERENCES comments (id) ON DELETE CASCADE,
        author VARCHAR(255) NOT NULL,
        content TEXT NOT NULL,
        path LTREE NOT NULL,
        created_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

CREATE INDEX IF NOT EXISTS idx_comments_path ON comments USING gist (path);

CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments (parent_id);

CREATE INDEX IF NOT EXISTS idx_comments_content ON comments USING gin (to_tsvector ('russian', content));