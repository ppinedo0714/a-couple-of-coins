# backend

Go REST API. The central data layer — serves the frontend and calls out to Python services as needed.

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
| `docs/architecture.md` | System overview, directory layout, layer rules, dependencies |
| `docs/data-model.md` | ERD and field-by-field table descriptions |
| `docs/api.md` | Complete route reference with request/response shapes |
| `docs/flows.md` | Sequence diagrams for auth, CSV import, and classification |

## Structure

To be documented once the service is scaffolded.

## Conventions

- Entry point: `cmd/api/main.go`
- Internal packages: `internal/` (not importable by external code)
- Layer order: `handlers` → `services` → `repository` — never skip layers
- Route handlers: `internal/handlers/`
- Business logic: `internal/services/`
- DB access: `internal/repository/`
- Tests: co-located with source, named `*_test.go`
