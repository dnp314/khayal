Khayal

a question specific time travelling journalling application

should i use AI ? reasonably YES

# issues
- more than one khayal can be written at once
- plain text contents

# features
- users
- placeholder showing various questions
- date picker

# dev
- logging/error handling: handled but have to take a look
- migrations
- dockerize



- applying migrations
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/001_init.sql
