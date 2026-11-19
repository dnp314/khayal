Khayal

a question specific time travelling journalling application

when a question is created, the return time would be null, But when a question is answered an option would be provided either to select a random time or fixed
and in the case for display the question list at home page we only display the questions whose return time is exceeding the current time

should i use AI ? reasonably YES

# issues
- more than one khayal can be written at once
- plain text contents stored in db: solution created on a separate branch
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
