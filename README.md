# Quran App — Core API

Go backend for the Quran app: auth (register, login, refresh, verify-email, logout, profile), Postgres via GORM, JWT access/refresh tokens, and S3-compatible object storage (MinIO in dev, AWS S3 in prod).

## Stack

- Go 1.26, Chi router, GORM + Postgres 16, golang-jwt
- AWS SDK Go v2 for S3-compatible storage
- MinIO (dev) via `docker-compose.dev.yml`
- Swagger UI (dark themed, non-production only), zerolog logging

## Prerequisites

- Go 1.26+
- Docker + Docker Compose plugin (for Postgres and MinIO)
- [`golang-migrate`](https://github.com/golang-migrate/migrate) is built locally via `make tools` — no global install needed

## Quickstart

```bash
make setup   # start Postgres + MinIO, apply migrations
make dev     # run API with live reload (air) on :8080
```

Or without reload: `make run`. Build a binary: `make build` (`./bin/app`).

Override the DB URL when needed:

```bash
make migrate-up DB_URL=postgres://...
```

## Services & ports

| Service          | Address               | Credentials (dev) |
|------------------|-----------------------|-------------------|
| API              | http://localhost:8080 | —                 |
| Swagger UI       | http://localhost:8080/swagger/index.html (not in `production`) | — |
| Postgres         | localhost:5432        | gorm / gorm       |
| MinIO API        | http://localhost:9000 | minioadmin / minioadmin |
| MinIO console    | http://localhost:9001 | minioadmin / minioadmin |

## Configuration

`config.toml` holds dev defaults. Every section can be overridden by environment variables (production should use env, never edit the file):

| Section    | Env vars                                                      |
|------------|---------------------------------------------------------------|
| `app`      | `APP_PORT`, `APP_ENV` (`production` disables Swagger UI)      |
| `database` | `DATABASE_URL`, `DATABASE_MAX_OPEN_CONNS`, `DATABASE_MAX_IDLE_CONNS` |
| `jwt`      | `JWT_SECRET` (min 32 random chars), `JWT_ISSUER`, `JWT_ACCESS_TTL`, `JWT_REFRESH_TTL` |
| `s3`       | `S3_ENDPOINT`, `S3_REGION`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET`, `S3_USE_PATH_STYLE`, `S3_PRESIGN_TTL` |

For real AWS S3 set `S3_ENDPOINT=""`, the bucket region, and IAM credentials.

## Project layout

```
cmd/app                 entrypoint (wiring only)
config.toml             dev config (env overrides in prod)
docker-compose.dev.yml  Postgres + MinIO
migrations/             golang-migrate SQL pairs
internal/core           config, di, database, storage, security (JWT),
                        middleware (auth, errors, logging), routes, swagger
internal/modules/user   feature module: domain → application → infrastructure → presentation
internal/shared         response envelope, AppError + registry, validation
```

Conventions: handlers return errors, `middleware.Handle` maps them to responses; each package registers its own sentinel errors via `RegisterError` in an `errors_mapping.go` so the handler stays domain-free. New modules follow the same `domain → application → infrastructure → presentation` split.

## Development

```bash
make test    # all tests (storage live-tests run when MinIO is up, skip otherwise)
make vet     # go vet + gofmt check
make fmt     # format code
make swag    # regenerate Swagger docs after annotation changes
make migrate-new NAME=create_x  # new migration pair
```

Full target list: `make help`.
