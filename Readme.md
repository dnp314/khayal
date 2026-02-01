TODO:

- Tweaking the activation process for better user experience
- Updating the question model

Reference:

- [Svelte Documentation](https://svelte.dev/)

Windows Development:

- Postgress Access: psql %KHAYAL_DB_DSN%
- Running Migrations: migrate -path=./migrations -database %KHAYAL_DB_DSN% up
- Creating Migrations: migrate create -seq -ext=.sql -dir=./migrations create_movies_table

Questions:

- How to setup users and their authentication and authorization
- User activation

Learn:

- Understanding the error library, error.Is, error.As,
  https://go.dev/blog/error-handling-and-go

- Embeddings in go

- Authentication and authorization
  https://www.youtube.com/watch?v=A95rliroC8Q&list=PLui3EUkuMTPgZcV0QhQrOcwMPcBCcd_Q1&index=8

Facts:

- QueryRow for queries, where only a single row of data is returned
