Khayal

a question specific time travelling journalling application

when a question is created, the return time would be null, But when a question is answered an option would be provided either to select a random time or fixed
and in the case for display the question list at home page we only display the questions whose return time is exceeding the current time

should i use AI ? reasonably YES

# issues
- more than one khayal can be written at once
- plain text contents stored in db
- timezone: store in UTC and while displaying localize

# features
- users
- placeholder showing various questions
- date picker
- slug instead of id

# dev
- logging/error handling: handled but have to take a look
- migrations
- dockerize
- testing/profiling



- applying migrations
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/001_init.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/002_sessions.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/003_questions_user_id.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/004_e2e_encryption_columns.sql

- final migration (DESTRUCTIVE, only after every user and row is migrated)
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/005_drop_plaintext_columns.sql

# client-side encryption

Question and khayal text is encrypted in the browser with Web Crypto and only
ciphertext is stored in Postgres. See:

- static/js/crypto.js   - PBKDF2/HKDF/AES-GCM primitives
- static/js/keystore.js  - IndexedDB storage for the non-extractable data key
- static/js/auth.js      - signup, login, recovery and password change
- static/js/content.js   - htmx encrypt/decrypt hooks

Key setup:
- PBKDF2-SHA256, 600000 iterations, per-user random salt
- HKDF-SHA256 splits the master key into an auth key and a wrapping key
- the server stores bcrypt(auth key), KDF parameters, and the wrapped data key
- the data key is stored in the browser as a non-extractable AES-GCM CryptoKey

Passwords never reach the server for new accounts. The legacy login path sends
the old password once to authenticate existing accounts before setting up
encryption and migrating their plaintext rows.

Environment:
- DATABASE_URL - Postgres connection string
- KHAYAL_INSECURE_COOKIES=1 - allow non-Secure cookies for local HTTP development

Tests:
go test ./...
KHAYAL_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/repository
npm test
