# backend

> Up: [root CLAUDE.md](../CLAUDE.md) · Sibling services: [`frontend/`](../frontend/CLAUDE.md), [`services/predictions/`](../services/predictions/CLAUDE.md)

Go REST API. The central data layer — serves the [React frontend](../frontend/CLAUDE.md) and calls out to the [Python prediction service](../services/predictions/CLAUDE.md) as needed. The frontend never talks to Python directly; all cross-service traffic flows through this API.

## Commands

| Command | Description |
|---------|-------------|
| `go mod download` | Download dependencies |
| `go run ./cmd/api` | Start dev server (port 8000) |
| `go test ./...` | Run all tests |
| `go build ./cmd/api` | Build binary |
| `go vet ./...` | Lint / static analysis |

## Documentation

Full architecture docs live in `docs/`:

| File | Contents |
|------|----------|
| [`docs/architecture.md`](./docs/architecture.md) | System overview, directory layout, layer rules, dependencies |
| [`docs/data-model.md`](./docs/data-model.md) | ERD and field-by-field table descriptions |
| [`docs/api.md`](./docs/api.md) | Complete route reference with request/response shapes |
| [`docs/flows.md`](./docs/flows.md) | Sequence diagrams for auth, CSV import, and classification |

## Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go         # entry point — wires config, DB, router, starts server
├── internal/
│   ├── config/             # env var loading (godotenv)
│   ├── db/                 # pgx connection pool setup
│   ├── auth/               # JWT creation/validation, bcrypt helpers, OAuth client
│   ├── models/             # Go structs matching DB tables (User, Account, Transaction, ...)
│   ├── repository/         # DB queries, one file per table
│   ├── services/           # business logic, one file per domain
│   └── handlers/           # HTTP handlers (chi router), one file per resource group
├── migrations/             # goose SQL migration files
├── go.mod
└── go.sum
```

Layer rules (one-way dependency): `handlers` → `services` → `repository`. Handlers never call repository directly; services never import handlers.

## Conventions

- Entry point: `cmd/api/main.go`
- Internal packages: `internal/` (not importable by external code)
- Layer order: `handlers` → `services` → `repository` — never skip layers
- Route handlers: `internal/handlers/`
- Business logic: `internal/services/`
- DB access: `internal/repository/`
- Tests: co-located with source, named `*_test.go`
