# Backend Architecture

## System Overview

The backend is a Go REST API that acts as the central data layer for the application. It sits between the React frontend and the Python prediction service — the frontend never calls the Python service directly.

```
┌─────────────┐        ┌─────────────────┐        ┌──────────────────────┐
│   frontend  │ ──────▶│    backend API  │ ──────▶│ services/predictions │
│  React + TS │  HTTP  │      Go         │  HTTP  │   Python + FastAPI   │
│  port: 3000 │        │   port: 8000    │        │     port: 8001       │
└─────────────┘        └─────────────────┘        └──────────────────────┘
                                │
                                ▼
                        ┌───────────────┐
                        │  PostgreSQL   │
                        │   port: 5432  │
                        └───────────────┘
```

## Directory Layout

```
backend/
├── cmd/
│   └── api/
│       └── main.go              ← entry point: wire deps, run startup recovery sweep, register routes, start server
├── internal/
│   ├── config/
│   │   └── config.go            ← typed Config struct loaded from env vars
│   ├── db/
│   │   └── db.go                ← opens and exposes pgx connection pool
│   ├── auth/
│   │   ├── jwt.go               ← issue and validate JWT tokens
│   │   ├── oauth.go             ← Google + GitHub OAuth2 flows
│   │   ├── password.go          ← bcrypt hash + compare helpers
│   │   └── middleware.go        ← chi middleware: validates JWT, sets user on context
│   ├── models/
│   │   ├── user.go
│   │   ├── account.go
│   │   ├── group.go
│   │   ├── category.go
│   │   ├── transaction.go
│   │   └── import_job.go
│   ├── repository/
│   │   ├── users.go
│   │   ├── accounts.go
│   │   ├── groups.go
│   │   ├── categories.go
│   │   ├── transactions.go
│   │   └── import_jobs.go
│   ├── services/
│   │   ├── accounts.go
│   │   ├── transactions.go
│   │   ├── groups.go            ← business logic for both groups and categories
│   │   ├── importer/
│   │   │   └── csv.go           ← parse CSV, normalize rows, bulk-insert via repository
│   │   └── predictor/
│   │       └── client.go        ← HTTP client for the Python prediction service
│   └── handlers/
│       ├── auth.go
│       ├── accounts.go
│       ├── transactions.go
│       ├── groups.go            ← handles /groups and /groups/categories
│       ├── imports.go
│       └── health.go
├── migrations/
│   ├── 001_create_users.sql
│   ├── 002_create_accounts.sql
│   ├── 003_create_category_groups.sql
│   ├── 004_create_categories.sql
│   ├── 005_create_transactions.sql
│   ├── 006_create_import_jobs.sql
│   ├── 007_create_oauth_connections.sql
│   └── 008_create_balance_checkpoints.sql
├── go.mod
├── go.sum
└── CLAUDE.md
```

## Layer Rules

The codebase is split into four layers. Each layer may only call the layer directly below it.

| Layer | Responsibility | May call |
|-------|---------------|----------|
| `handlers/` | Parse HTTP request, call service, write HTTP response | `services/` |
| `services/` | Business logic, orchestration, external HTTP clients | `repository/`, `predictor/client.go` |
| `repository/` | Execute SQL queries, return model structs | `db/` |
| `models/` | Plain Go structs — data shapes only | Nothing |

**Never** call `repository` directly from a handler, or put SQL in a service.

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/go-chi/chi/v5` | v5 | HTTP router |
| `github.com/jackc/pgx/v5` | v5 | PostgreSQL driver with connection pool |
| `github.com/golang-jwt/jwt/v5` | v5 | JWT token issue and validation |
| `golang.org/x/crypto` | latest | bcrypt for password hashing |
| `golang.org/x/oauth2` | latest | OAuth2 flows (Google, GitHub) |
| `github.com/joho/godotenv` | latest | Load `.env` file in development |
| `github.com/pressly/goose/v3` | v3 | SQL-based database migrations |
| `github.com/go-chi/httprate` | latest | Rate-limiting middleware for auth routes |

## Migrations

`goose` migrations run automatically at startup (`cmd/api/main.go`, ahead of the recovery sweep in `flows.md` §5a) — there is no separate manual migration step in the deploy process.

Every `up` migration has a corresponding `down` in the same file (`-- +goose Up` / `-- +goose Down`) — a migration without a working `down` is not merged.

**Rollback policy:** the default response to a bad migration in production is to roll forward with a corrective migration, not to run `goose down`. Automatic rollback isn't wired into the deploy process, and running `down` against a live database is riskier than it looks — if the previous app version is still serving traffic (or the current version doesn't get redeployed at the same instant), it can end up querying a schema it no longer matches (e.g. a dropped column). If a migration genuinely needs to be undone rather than fixed forward: take the API offline, run `goose down` manually against `DATABASE_URL`, then redeploy the prior application version — never leave the app and schema versions mismatched while serving traffic.

## Rate Limiting

`httprate` middleware wraps the auth route group (`POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`) in `cmd/api/main.go`, ahead of the JWT auth middleware (that group is otherwise unauthenticated, so it can't be gated by user identity). Two limiters apply together:

| Scope | Limit | Key |
|-------|-------|-----|
| Per IP | 20 requests / 5 min | `httprate.KeyByIP` |
| Per email | 5 requests / 15 min | request body's `email` field (register/login only — `refresh` has no email, IP limit applies alone) |

Either limiter tripping returns `429` with a `Retry-After` header. Limits are in-process, in-memory counters (no Redis or other shared store) — the accepted tradeoff is that a horizontally-scaled deployment (multiple API instances behind a load balancer) would enforce these limits per-instance rather than globally, meaning actual allowed throughput scales with instance count. This is a looser tradeoff than the OAuth `state` cookie's (see `flows.md` §3) or the session/refresh-token design's (see `data-model.md` Key Design Decisions): rate limits are advisory throttling, not a security invariant, so degraded-but-nonzero protection under multiple instances is acceptable. Revisit with a shared store (e.g. Redis) if/when the backend runs as more than one process.

## Environment Variables

See `../../.env.example` for the full list. Key variables:

| Variable | Description |
|----------|-------------|
| `PORT` | Port the API listens on (default: 8000) |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Secret for signing JWT tokens (min 32 chars) |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | Google OAuth app credentials |
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | GitHub OAuth app credentials |
| `PREDICTIONS_SERVICE_URL` | Base URL of the Python prediction service |
| `PREDICTIONS_TIMEOUT_SECONDS` | Per-batch HTTP timeout for calls to the prediction service from `POST /transactions/classify` (default: 15) |

## Observability

There is no structured logging library or metrics/tracing tooling in the dependency table yet — the reliability behavior added for imports and classification (batching, timeouts, the advisory lock, the startup recovery sweep, dedup flagging; see `flows.md` §5, §5a, §6) currently only produces whatever ad hoc `log.Printf`-style output exists at each call site, with no consistent structure, log levels, or correlation between related log lines (e.g. everything produced by one `POST /transactions/classify` run for one user). This is a known gap, not yet addressed: a structured logger and, if needed, request/job-scoped correlation IDs are the natural next step.

## Testing

Tests are co-located per layer (`internal/services/*_test.go`, `internal/repository/*_test.go`, ...), named `*_test.go` per the root `CLAUDE.md` convention. Coverage is expected to concentrate on `services/` — business logic, especially the reliability-sensitive paths (advisory-lock acquisition in classification, checkpoint date validation, dedup-hash matching, the startup recovery sweep) — with `repository/` tests covering query correctness against a real (not mocked) Postgres instance. There is no integration or end-to-end test runner wired up yet that exercises a handler through to the database and back, or across the backend ↔ `services/predictions` boundary.
