-- Creates the "users" table for people who sign in with Google.
-- Run once on any database:
--   psql -U postgres -d url_shortener -f migrations/003_create_users.sql
-- google_sub is Google's permanent ID for the person (the "sub" claim) and is the identity;
-- email and name are for display and can change.

BEGIN;

CREATE TABLE users (
    id            BIGSERIAL   PRIMARY KEY,
    google_sub    TEXT        NOT NULL UNIQUE,
    email         TEXT        NOT NULL,
    name          TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
