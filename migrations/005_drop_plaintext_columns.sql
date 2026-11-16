-- Final E2E encryption migration. DESTRUCTIVE: drops legacy plaintext columns.
--
-- Apply only after every user has crypto_setup_at set and every question/khayal
-- has ciphertext. The guards below abort the migration otherwise.
--
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/005_drop_plaintext_columns.sql

BEGIN;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE crypto_setup_at IS NULL) THEN
        RAISE EXCEPTION 'not all users have completed encryption setup';
    END IF;
    IF EXISTS (SELECT 1 FROM questions WHERE question_ciphertext IS NULL) THEN
        RAISE EXCEPTION 'not all questions have ciphertext';
    END IF;
    IF EXISTS (SELECT 1 FROM khayals WHERE khayal_ciphertext IS NULL) THEN
        RAISE EXCEPTION 'not all khayals have ciphertext';
    END IF;
END
$$;

ALTER TABLE questions ALTER COLUMN question_ciphertext SET NOT NULL;
ALTER TABLE khayals ALTER COLUMN khayal_ciphertext SET NOT NULL;

ALTER TABLE users DROP COLUMN password_hash;
ALTER TABLE questions DROP COLUMN question;
ALTER TABLE khayals DROP COLUMN khayal;

COMMIT;
