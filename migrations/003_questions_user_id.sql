-- Link questions to the user who created them.
--
-- Apply after migrations/001_init.sql:
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/003_questions_user_id.sql

BEGIN;

ALTER TABLE questions ADD COLUMN IF NOT EXISTS user_id INTEGER;

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

CREATE INDEX IF NOT EXISTS idx_questions_user_id ON questions(user_id);

COMMIT;
