-- One-time data migration from the old flat "urls" table into "short_urls".
-- Only for databases that still have the old "urls" table. Skip it on a fresh database.
-- Run after 001_create_short_urls.sql:
--   psql -U postgres -d url_shortener -f migrations/002_migrate_urls_data.sql

BEGIN;

-- Partitions for the months that had data in "urls".
CREATE TABLE IF NOT EXISTS short_urls_2026_04 PARTITION OF short_urls FOR VALUES FROM ('2026-04-01 00:00+00') TO ('2026-05-01 00:00+00');
CREATE TABLE IF NOT EXISTS short_urls_2026_07 PARTITION OF short_urls FOR VALUES FROM ('2026-07-01 00:00+00') TO ('2026-08-01 00:00+00');

-- Postgres routes each row into the right monthly partition.
INSERT INTO short_urls (short_id, long_url, user_id, expires_at, created_at)
SELECT short_id, long_url, user_id, expires_at, created_at FROM urls;

-- Keep the old table as a backup; drop it once everything works: DROP TABLE urls_old;
ALTER TABLE urls RENAME TO urls_old;

COMMIT;
