CREATE TABLE IF NOT EXISTS url_created_events (
    id BIGSERIAL PRIMARY KEY,
    short_url TEXT,
    original_url TEXT,
    user_id TEXT,
    event_time TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_short_url ON url_created_events(short_url);
CREATE INDEX IF NOT EXISTS idx_original_url ON url_created_events(original_url);

CREATE TABLE IF NOT EXISTS url_deleted_events (
    id BIGSERIAL PRIMARY KEY,
    short_url TEXT,
    user_id TEXT,
    event_time TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_short_url ON url_deleted_events(short_url);

CREATE TABLE IF NOT EXISTS url_visited_events (
    id BIGSERIAL PRIMARY KEY,
    short_url TEXT,
    event_time TIMESTAMPTZ,
    user_id TEXT,
    referer TEXT,
    ip_address INET,
    user_agent TEXT,  
    country CHAR(2),
    region VARCHAR(100),
    city VARCHAR(100),
    browser VARCHAR(50),
    os VARCHAR(50),
    device_type VARCHAR(20)
);
CREATE INDEX IF NOT EXISTS idx_short_url ON url_visited_events(short_url);
CREATE INDEX IF NOT EXISTS idx_event_time ON url_visited_events(event_time);
CREATE INDEX IF NOT EXISTS idx_ip_address ON url_visited_events(ip_address);
CREATE INDEX IF NOT EXISTS idx_geography ON url_visited_events(country, region, city);
CREATE INDEX IF NOT EXISTS idx_short_url_event_time ON url_visited_events(short_url, event_time);