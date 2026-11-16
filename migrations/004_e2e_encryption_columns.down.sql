-- Revert the additive encryption columns. This only makes sense before any
-- encrypted rows or accounts have been created.

BEGIN;

DROP INDEX IF EXISTS idx_khayals_question_client_id;
DROP INDEX IF EXISTS idx_questions_user_client_id;

ALTER TABLE khayals DROP COLUMN IF EXISTS encryption_version;
ALTER TABLE khayals DROP COLUMN IF EXISTS khayal_ciphertext;
ALTER TABLE khayals DROP COLUMN IF EXISTS client_id;
ALTER TABLE khayals ALTER COLUMN khayal SET NOT NULL;

ALTER TABLE questions DROP COLUMN IF EXISTS encryption_version;
ALTER TABLE questions DROP COLUMN IF EXISTS question_ciphertext;
ALTER TABLE questions DROP COLUMN IF EXISTS client_id;
ALTER TABLE questions ALTER COLUMN question SET NOT NULL;

ALTER TABLE users DROP COLUMN IF EXISTS crypto_setup_at;
ALTER TABLE users DROP COLUMN IF EXISTS recovery_verifier;
ALTER TABLE users DROP COLUMN IF EXISTS recovery_wrapped_data_key_iv;
ALTER TABLE users DROP COLUMN IF EXISTS recovery_wrapped_data_key;
ALTER TABLE users DROP COLUMN IF EXISTS wrapped_data_key_iv;
ALTER TABLE users DROP COLUMN IF EXISTS wrapped_data_key;
ALTER TABLE users DROP COLUMN IF EXISTS kdf_version;
ALTER TABLE users DROP COLUMN IF EXISTS kdf_salt;
ALTER TABLE users DROP COLUMN IF EXISTS kdf_iterations;
ALTER TABLE users DROP COLUMN IF EXISTS kdf_algorithm;
ALTER TABLE users DROP COLUMN IF EXISTS auth_key_hash;

COMMIT;
