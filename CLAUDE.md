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

| Directory | Language | Role | Service guide |
|-----------|----------|------|---------------|
| `frontend/` | TypeScript (React + Vite) | User interface | [`frontend/CLAUDE.md`](./frontend/CLAUDE.md) |
| `backend/` | Go | REST API, data layer | [`backend/CLAUDE.md`](./backend/CLAUDE.md) |
| `services/predictions/` | Python (FastAPI) | Predictive analytics | [`services/predictions/CLAUDE.md`](./services/predictions/CLAUDE.md) |

Each service's `CLAUDE.md` is the entry point to its commands, structure, conventions, and (for `frontend/` and `backend/`) a `docs/` folder with deeper architecture, API, and data-model references.

## Shared conventions

- TypeScript strict mode in the frontend
- REST JSON APIs between all services
- Env vars documented in `.env.example` — never commit `.env`
- Tests: `*_test.go` (Go), co-located `*.test.ts` (frontend), `test_*.py` (Python)

## Design

The frontend uses a locked **Warm Coin / Bronze-Gold** design system. Any visual or brand change must update [`frontend/docs/design-system.md`](./frontend/docs/design-system.md) first — that file is the single source of truth for tokens, typography, and semantic-color rules.
