# Go Account Service

> Disclaimer:
> This repository is not the result of vibe coding. It is built from deliberate engineering decisions, practical production experience, and the writer's own implementation judgment. Its primary purpose is as a portfolio product, and it is also intended to be viable as a commercial foundation.

`go-account-service` is an account and identity service written in Go. It exposes an HTTP API for authentication and account management, persists core data in PostgreSQL, stores token sessions in Redis, and publishes domain events through RabbitMQ for asynchronous email workflows.

The service is split into two runtime entrypoints:

- `cmd/http/main.go`: REST API
- `cmd/worker/main.go`: event consumer for email and follow-up workflows

## What It Handles

- User registration
- Login with access token and refresh token
- Refresh token rotation
- Current user lookup
- Profile update
- Password change
- Email verification request and verification
- Reset password request, verification, and completion
- Change email request and verification

## Architecture

- `internal/application/usecase`
  Business use cases and orchestration
- `internal/handler/rest`
  HTTP handlers and route registration
- `internal/handler/event`
  Event consumers
- `internal/infrastructure/database/postgres`
  PostgreSQL repositories
- `internal/infrastructure/cache/redis`
  Redis session repository
- `internal/infrastructure/messaging/rabbitmq`
  Domain event producer and consumer
- `internal/infrastructure/notification/email/smtp`
  SMTP email delivery
- `internal/provider`
  Wiring for env, repositories, JWT, SMTP, RabbitMQ, telemetry, and use cases

Persistence is split by concern:

- PostgreSQL: users and user verifications
- Redis: access token and refresh token sessions
- RabbitMQ: async domain events

## HTTP API

The REST server mounts routes under `/api/v1`.

Public endpoints:

- `POST /api/v1/register`
- `POST /api/v1/login`
- `POST /api/v1/refresh-token`
- `GET /api/v1/email-verification/verify`
- `POST /api/v1/reset-password/request`
- `POST /api/v1/reset-password/verify`
- `POST /api/v1/reset-password`

Authenticated endpoints:

- `GET /api/v1/me`
- `PATCH /api/v1/me`
- `POST /api/v1/change-password`
- `POST /api/v1/email-verification/request`
- `POST /api/v1/change-email/request`
- `POST /api/v1/change-email/verify`

Utility endpoints:

- `GET /health`
- `GET /swagger/index.html`

## Event Flow

The API publishes domain events and the worker consumes them.

Published events:

- `account.user.created`
- `account.user.verification.created`
- `account.user.email_verified`
- `account.user.reset_password`
- `account.user.email_changed`
- `account.user.updated`

RabbitMQ bindings configured in the code:

- `email_queue`
  Handles user created, verification created, email verified, reset password, and email changed events
- `update_queue`
  Handles user updated events

## Local Development

The repository ships with a full local stack in `docker-compose.local.yaml`.

Included services:

- app on `8001`
- worker
- PostgreSQL on `5432`
- RabbitMQ on `5672`
- RabbitMQ management UI on `15672`
- Redis
- Redis Insight on `5540`
- Adminer on `8080`
- Grafana on `3000`
- Prometheus on `9090`
- Loki on `3100`
- Tempo
- OpenTelemetry Collector
- cAdvisor
- node-exporter

### Start

```sh
make start-local
make start-local-build
```

For other environments:

```sh
make start-dev
make start-dev-build
make start-prod
make start-prod-build
```

### Stop

```sh
make stop-local
make stop-local-v
```

For other environments:

```sh
make stop-dev
make stop-dev-v
make stop-prod
make stop-prod-v
```

### Restart

```sh
make restart-local
make restart-dev
make restart-prod
```

### Logs

```sh
make log-app-local
make log-worker-local
make log-app-dev
make log-worker-dev
make log-app-prod
make log-worker-prod
```

## Make Targets

Most environment-aware targets require `ENV=local`, `ENV=development`, or `ENV=production`. The Makefile also provides shortcut targets for common environment-specific commands.

Start and stop:

```sh
make start ENV=local
make start-build ENV=local
make stop ENV=local
make stop-v ENV=local
make restart ENV=local
```

Build images:

```sh
make build-local
make build-dev IMAGE_TAG=v1.0.0
make build-prod IMAGE_TAG=v1.0.0
```

Migrations:

```sh
make migration MIGRATION_NAME=create_users_table
make migrate-local
make migrate-rollback-local MIGRATION_STEP=1
make migrate-drop-local
make migrate-fresh-local
```

Development and production migration shortcuts:

```sh
make migrate-dev
make migrate-rollback-dev MIGRATION_STEP=1
make migrate-fresh-dev
make migrate-prod
make migrate-rollback-prod MIGRATION_STEP=1
make migrate-fresh-prod
```

Seeders:

```sh
make seed-local
make seed-dev
make seed-prod
```

Generated files and tests:

```sh
make swag
make mock
make test
```

## Environment

Environment files are selected by `APP_ENV`:

- `local` -> `.env.local`
- `development` -> `.env.development`
- `production` -> `.env.production`

At minimum, the code expects these variables.

Application:

- `APP_ENV`
- `APP_NAME`
- `APP_URL`
- `APP_ADDRESS`
- `APP_PRIVACY_URL`
- `APP_UNSUBSCRIBE_URL`
- `APP_RATE_LIMITER_ATTEMPT`
- `APP_RATE_LIMITER_DURATION`
- `APP_EMAIL_VERIFICATION_URL`
- `APP_EMAIL_VERIFICATION_EXPIRED`
- `APP_RESET_PASSWORD_EXPIRED`
- `APP_CHANGE_EMAIL_EXPIRED`
- `APP_VERIFICATION_MAX_ATTEMPT`
- `SWAGGER_BASE_URL`

PostgreSQL:

- `DB_USERNAME`
- `DB_PASSWORD`
- `DB_HOST`
- `DB_PORT`
- `DB_DATABASE`

Redis:

- `REDIS_USERNAME`
- `REDIS_PASSWORD`
- `REDIS_HOST`
- `REDIS_PORT`
- `REDIS_DATABASE`

RabbitMQ:

- `RABBITMQ_USERNAME`
- `RABBITMQ_PASSWORD`
- `RABBITMQ_HOST`
- `RABBITMQ_PORT`
- `RABBITMQ_EXCHANGE`
- `RABBITMQ_EXCHANGE_TYPE`

JWT:

- `JWT_ACCESS_TOKEN_SECRET_KEY`
- `JWT_ACCESS_TOKEN_EXPIRED`
- `JWT_REFRESH_TOKEN_SECRET_KEY`
- `JWT_REFRRESH_TOKEN_EXPIRED`

SMTP:

- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USERNAME`
- `SMTP_PASSWORD`
- `SMTP_FROM`
- `SMTP_TLS`

Telemetry:

- `TELEMETRY_COLLECTOR_HOST`
- `TELEMETRY_COLLECTOR_PORT`

## Build and Release

The repository uses GitHub Actions for versioning and image publishing.

- `release.yaml`
  Triggers on pushes to `main` and `dev`
- `versioning.yaml`
  Bumps a semantic version tag from the latest tag using the latest commit message
- `build.yaml`
  Builds and pushes Docker images to `ghcr.io/dani-susanto/go-account-service`

Commit message rules used by the versioning workflow:

- `feat!` or `BREAKING CHANGE`: major bump
- `feat`: minor bump
- `fix`, `perf`, or `refactor`: patch bump
- anything else: no release

Branch aliasing:

- `main` publishes `latest`
- `dev` publishes `dev`

## Testing

Tests are currently organized around:

- application use cases
- REST handlers
- event handlers
- Redis repository
- PostgreSQL repository

Run the repository test targets with:

```sh
make test
```

## Swagger

Swagger artifacts are committed under `docs/`.

Regenerate them with:

```sh
make swag
```

The HTTP service exposes Swagger UI at:

```text
/swagger/index.html
```

## Notes

- The REST server listens on `:8001`.
- The API base path is `/api/v1`.
- The worker is required for outbound email workflows to complete.
- The repository uses `github.com/dani-susanto/go-common` for shared env, JWT, REST, logging, SMTP, PostgreSQL, and Redis helpers.
