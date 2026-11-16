-- Best-effort rollback of the final migration.
--
-- The server never has the data key, so plaintext cannot be restored. This
-- recreates the old columns as nullable and removes the ciphertext NOT NULL
-- constraints. The running application must be an older version that can read
-- legacy rows before this rollback is useful.

BEGIN;

ALTER TABLE questions ADD COLUMN IF NOT EXISTS question TEXT;
ALTER TABLE khayals ADD COLUMN IF NOT EXISTS khayal TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;

ALTER TABLE questions ALTER COLUMN question_ciphertext DROP NOT NULL;
ALTER TABLE khayals ALTER COLUMN khayal_ciphertext DROP NOT NULL;

COMMIT;
