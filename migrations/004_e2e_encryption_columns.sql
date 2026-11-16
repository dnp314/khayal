-- Add client-side encryption key material and encrypted content columns.
--
-- This migration is additive and keeps the legacy plaintext columns so existing
-- rows can be migrated from the browser. Apply after migrations/003.
--
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/004_e2e_encryption_columns.sql

BEGIN;

ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_key_hash TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS kdf_algorithm TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS kdf_iterations INTEGER;
ALTER TABLE users ADD COLUMN IF NOT EXISTS kdf_salt BYTEA;
ALTER TABLE users ADD COLUMN IF NOT EXISTS kdf_version SMALLINT NOT NULL DEFAULT 1;
ALTER TABLE users ADD COLUMN IF NOT EXISTS wrapped_data_key BYTEA;
ALTER TABLE users ADD COLUMN IF NOT EXISTS wrapped_data_key_iv BYTEA;
ALTER TABLE users ADD COLUMN IF NOT EXISTS recovery_wrapped_data_key BYTEA;
ALTER TABLE users ADD COLUMN IF NOT EXISTS recovery_wrapped_data_key_iv BYTEA;
ALTER TABLE users ADD COLUMN IF NOT EXISTS recovery_verifier BYTEA;
ALTER TABLE users ADD COLUMN IF NOT EXISTS crypto_setup_at TIMESTAMPTZ;

-- New accounts do not have a legacy password hash.
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE questions ADD COLUMN IF NOT EXISTS client_id UUID;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS question_ciphertext BYTEA;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS encryption_version SMALLINT;
ALTER TABLE questions ALTER COLUMN question DROP NOT NULL;

ALTER TABLE khayals ADD COLUMN IF NOT EXISTS client_id UUID;
ALTER TABLE khayals ADD COLUMN IF NOT EXISTS khayal_ciphertext BYTEA;
ALTER TABLE khayals ADD COLUMN IF NOT EXISTS encryption_version SMALLINT;
ALTER TABLE khayals ALTER COLUMN khayal DROP NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_questions_user_client_id
    ON questions(user_id, client_id)
    WHERE client_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_khayals_question_client_id
    ON khayals(question_id, client_id)
    WHERE client_id IS NOT NULL;

COMMIT;
