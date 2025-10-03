# fullstack-go-api

This repository provides a simple Go backend and React + TypeScript frontend for user management flows. The project now includes container orchestration and database provisioning scripts so you can spin up a complete development stack with PostgreSQL, MongoDB, and Redis.

## Getting started with Docker Compose

1. Ensure Docker and Docker Compose are installed on your machine.
2. From the repository root, build and start the full stack:

   ```bash
   docker compose up --build
   ```

   The command launches the following services:

   | Service   | Description                                | Exposed Port |
   |-----------|--------------------------------------------|--------------|
   | backend   | Go HTTP API (JWT-authenticated)            | 8080         |
   | frontend  | React/Vite single-page application         | 3000         |
   | postgres  | Primary relational data store              | 5432         |
   | mongo     | Optional document store for user metadata  | 27017        |
   | redis     | Caching/session infrastructure             | 6379         |

3. Access the frontend at `http://localhost:3000` and the API at `http://localhost:8080`.

### Environment variables

The Docker Compose configuration wires together the services with sensible defaults. If you need to override them, create a `.env` file at the repository root and reference it from `docker-compose.yml` or override values on the command line, for example:

```bash
docker compose run --rm backend /bin/sh
```

Within the backend container you can export new connection strings or JWT secrets before restarting the service.

## Database design for user management

### PostgreSQL schema

The file [`deployments/postgres/init-users.sql`](deployments/postgres/init-users.sql) provisions an `app_users` table tailored for authentication and profile management:

- `id` uses UUIDs for globally unique identifiers.
- `email` is unique and indexed for fast lookups during login.
- `password_hash` stores a pre-hashed (e.g., SHA-256) password.
- `display_name`, `roles`, `is_active`, and `metadata` capture profile and authorization context.
- `last_login_at`, `created_at`, and `updated_at` support auditing, with a trigger ensuring `updated_at` stays current.

This schema provides a solid foundation for expanding into role-based access control or multi-tenant scenarios.

### MongoDB collections

[`deployments/mongo/init-users.js`](deployments/mongo/init-users.js) seeds a complementary `users` collection. It creates indexes that mirror the relational schema (unique email, active flag) and demonstrates how to upsert a bootstrap administrator account. Use MongoDB to persist flexible profile metadata without altering the relational schema.

### Redis usage

Redis is provisioned for ephemeral caching or session storage (`redis:6379`). You can integrate it into the Go API by reading the `REDIS_ADDR` environment variable injected by Docker Compose.

## Local development without Docker

- **Backend**: `cd backend && go run ./`
- **Frontend**: `cd frontend && npm install && npm run dev`

Both processes expect the services to run on their default local ports.

## Continuous integration (Jenkins)

A [`Jenkinsfile`](Jenkinsfile) is included with a declarative pipeline that performs linting/build steps for both the Go backend and React frontend. Adapt the stages or add credentials bindings as needed for your environment.
