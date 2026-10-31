-- Khayal schema migration 001
--
-- Apply to an existing database:
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/001_init.sql
--
-- This migration is idempotent: it creates the schema from scratch when needed
-- and adds the newer questions.user_id / questions.return_at columns to an
-- existing database.

BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    name          TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS questions (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER REFERENCES users(id) ON DELETE CASCADE,
    question   TEXT NOT NULL,
    answered   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    return_at  TIMESTAMPTZ
);

-- Bring older databases up to date.
ALTER TABLE questions ADD COLUMN IF NOT EXISTS user_id INTEGER;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS return_at TIMESTAMPTZ;

-- Add the users -> questions relationship for databases created before this
-- migration. The column stays nullable so existing questions remain valid
-- until ownership is backfilled.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'questions_user_id_fkey'
          AND conrelid = 'questions'::regclass
    ) THEN
        ALTER TABLE questions
            ADD CONSTRAINT questions_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS khayals (
    id          SERIAL PRIMARY KEY,
    question_id INTEGER NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    khayal      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_khayals_question_id ON khayals(question_id);
CREATE INDEX IF NOT EXISTS idx_questions_user_id ON questions(user_id);

-- Keep questions.updated_at fresh on every UPDATE.
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_questions_updated_at ON questions;
CREATE TRIGGER trg_questions_updated_at
    BEFORE UPDATE ON questions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
