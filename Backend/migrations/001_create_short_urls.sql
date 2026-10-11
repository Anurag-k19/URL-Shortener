-- Creates the "short_urls" table, partitioned by month of created_at (UTC).
-- Run once on any database, including a fresh one:
--   psql -U postgres -d url_shortener -f migrations/001_create_short_urls.sql
-- Monthly partitions are created by the app on startup (storage.EnsurePartitions), not here.

BEGIN;

CREATE TABLE short_urls (
    id         BIGSERIAL,
    short_id   TEXT        NOT NULL,
    long_url   TEXT        NOT NULL,
    user_id    TEXT        NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Created on the parent, so Postgres adds them to every monthly partition automatically.
CREATE INDEX ON short_urls (short_id);           -- redirect lookups
CREATE INDEX ON short_urls (user_id, long_url);  -- "already shortened?" check

COMMIT;
