Khayal

a question specific time travelling journalling application

should i use AI ? reasonably YES

# issues
- more than one khayal can be written at once
- plain text contents stored in db

# features
- users
- placeholder showing various questions
- date picker
- slug instead of id

# dev
- logging/error handling: handled but have to take a look
- migrations
- dockerize
- testing



- applying migrations
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/001_init.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/002_sessions.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/003_questions_user_id.sql
