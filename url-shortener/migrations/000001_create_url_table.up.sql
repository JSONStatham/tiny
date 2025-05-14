CREATE TABLE IF NOT EXISTS url (
    id BIGSERIAL PRIMARY KEY,
    short_url CHAR(16) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_short_url ON url(short_url);
