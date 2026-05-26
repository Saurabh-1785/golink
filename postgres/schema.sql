-- GoLink Database Schema

-- Table to store shortened URLs
CREATE TABLE IF NOT EXISTS urls (
    id         SERIAL PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    long_url   TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Table to store click analytics
CREATE TABLE IF NOT EXISTS analytics (
    id             SERIAL PRIMARY KEY,
    short_code     VARCHAR(10) NOT NULL REFERENCES urls(short_code),
    click_count    BIGINT DEFAULT 0,
    last_clicked_at TIMESTAMP
);

-- Index for fast short_code lookups
CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_analytics_short_code ON analytics(short_code);
