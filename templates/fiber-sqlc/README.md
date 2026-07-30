# Fiber + SQLC API

A Fiber API with SQLC-generated queries, pgx, PostgreSQL, and JWT cookie authentication.

## Run locally

The StackForge CLI creates `.env` from `.env.example`. The application loads
that file automatically during local development:

```bash
docker-compose up -d postgres
go run .
```

When `.env` is not present, such as inside a production container, configuration
can still be supplied through the process environment.

The API listens on `http://localhost:3000`.

## User authentication

- `GET /health`
- `POST /users/register`
- `POST /users/login`
- `GET /users/me`
- `POST /users/logout`

Register and save the authentication cookies:

```bash
curl -c cookies.txt -X POST http://localhost:3000/users/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"StackForge User","email":"user@example.com","password":"strong-password"}'
```

Read the authenticated user:

```bash
curl -b cookies.txt http://localhost:3000/users/me
```

After changing SQL files, regenerate the Go package with:

```bash
sqlc generate
```

The application loads `.env` with `godotenv` and reads configuration with
`os.Getenv`. The controller depends only on the SQLC `Querier` interface, while
PostgreSQL pooling remains in `main.go`.
