Khayal

a question specific time travelling journalling application

when a question is created, the return time would be null, But when a question is answered an option would be provided either to select a random time or fixed
and in the case for display the question list at home page we only display the questions whose return time is exceeding the current time

should i use AI ? reasonably YES
caring is important !

# issues
- more than one khayal can be written at once
- plain text contents stored in db: solution created on a separate branch: first self hostable release
- timezone: store in UTC and while displaying localize

# features
- users
- placeholder showing various questions
- archive section for retiring questions
- date picker
- slug instead of id

# dev
- github actions on dev branch
- focus on infra: reliable: should perform as expected, scalable, maintainable: should be easy to add new features
- observability
- logging/error handling: handled but have to take a look
- migrations
- dockerize: having a docker container for db with migrations would have helped
- AGENTS.md
- testing/profiling: table driven tests ?

# docker
- database + migrations:
  `docker compose up -d db`

  migrations run automatically on first start of the db volume. To reset:
  `docker compose down -v`
- app (single arch, local):
  `docker compose up --build`

  serves http://localhost:8000, table `DATABASE_URL` env (defaults to localhost)
- multiarch image (pushed to a registry):
  `docker buildx build --platform linux/amd64,linux/arm64 -t <registry>/khayal:latest --push .`
