CREATE TABLE IF NOT EXISTS "url" (
    id serial PRIMARY KEY,
    alias CHAR(16) NOT NULL UNIQUE,
    url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_alias ON "url"(alias);
