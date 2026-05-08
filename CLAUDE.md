# a-couple-of-coins

Money tracking and budget app with predictive analytics.

## Architecture

```
┌─────────────┐        ┌─────────────────┐        ┌──────────────────────┐
│   frontend  │ ──────▶│    backend API  │ ──────▶│ services/predictions │
│  React + TS │  HTTP  │      Go         │  HTTP  │   Python + FastAPI   │
│  port: 3000 │        │   port: 8000    │        │     port: 8001       │
└─────────────┘        └─────────────────┘        └──────────────────────┘
```

- The frontend talks only to the backend API — never directly to Python services.
- The backend API aggregates data and calls Python services as needed.
- Each service is independently runnable.

## Running locally

| Command | Description |
|---------|-------------|
| (TBD) | Start all services — to be added once each service is scaffolded |

See each service's own CLAUDE.md for per-service commands.

## Services

| Directory | Language | Role |
|-----------|----------|------|
| `frontend/` | TypeScript (React + Vite) | User interface |
| `backend/` | Go | REST API, data layer |
| `services/predictions/` | Python (FastAPI) | Predictive analytics |

Read each service's `CLAUDE.md` for its commands, structure, and conventions.

## Shared conventions

- TypeScript strict mode in the frontend
- REST JSON APIs between all services
- Env vars documented in `.env.example` — never commit `.env`
- Tests: `*_test.go` (Go), co-located `*.test.ts` (frontend), `test_*.py` (Python)
