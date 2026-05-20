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
│       └── main.go              ← entry point: wire deps, register routes, start server
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
│   │   ├── category.go
│   │   ├── transaction.go
│   │   └── import_job.go
│   ├── repository/
│   │   ├── users.go
│   │   ├── accounts.go
│   │   ├── categories.go
│   │   ├── transactions.go
│   │   └── import_jobs.go
│   ├── services/
│   │   ├── accounts.go
│   │   ├── transactions.go
│   │   ├── categories.go
│   │   ├── importer/
│   │   │   └── csv.go           ← parse CSV, normalize rows, bulk-insert via repository
│   │   └── predictor/
│   │       └── client.go        ← HTTP client for the Python prediction service
│   └── handlers/
│       ├── auth.go
│       ├── accounts.go
│       ├── transactions.go
│       ├── categories.go
│       ├── imports.go
│       └── health.go
├── migrations/
│   ├── 001_create_users.sql
│   ├── 002_create_accounts.sql
│   ├── 003_create_categories.sql
│   ├── 004_create_transactions.sql
│   ├── 005_create_import_jobs.sql
│   ├── 006_create_oauth_connections.sql
│   └── 007_create_account_balance_snapshots.sql
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
