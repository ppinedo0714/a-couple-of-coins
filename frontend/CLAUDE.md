# frontend

> Up: [root CLAUDE.md](../CLAUDE.md) · Sibling services: [`backend/`](../backend/CLAUDE.md), [`services/predictions/`](../services/predictions/CLAUDE.md)

React + TypeScript UI. Communicates with the [backend API](../backend/CLAUDE.md) only — never directly with the [Python prediction service](../services/predictions/CLAUDE.md).

## Commands

| Command | Description |
|---------|-------------|
| `npm install` | Install dependencies |
| `npm run dev` | Start dev server (port 3000) |
| `npm test` | Run tests |
| `npm run lint` | Lint code |
| `npm run build` | Production build |

## Documentation

Full architecture docs live in `docs/`:

| File | Contents |
|------|----------|
| [`docs/architecture.md`](./docs/architecture.md) | Stack, directory layout, routing, data layer, auth strategy, theming |
| [`docs/pages.md`](./docs/pages.md) | Every page with purpose, components, data dependencies, behavior |
| [`docs/components.md`](./docs/components.md) | Shared components: layout, charts, transactions, accounts, primitives |
| [`docs/flows.md`](./docs/flows.md) | Sequence diagrams for register/login, OAuth, protected routes, CSV import |
| [`docs/design-system.md`](./docs/design-system.md) | Locked design tokens, typography, semantic-color rules (single source of truth) |

## Stack

- React + TypeScript (Vite)
- Tailwind CSS + shadcn/ui
- TanStack Query for server state
- React Router for routing
- Recharts for visualizations
- Auth: httpOnly cookies (set by backend) — fetches use `credentials: 'include'`

## Design system

**Warm Coin / Bronze-Gold** — amber `primary`, teal `accent`, rust `destructive`, cream/espresso backgrounds, Fraunces (display) + Inter (body) self-hosted via `@fontsource`. Full token reference in [`docs/design-system.md`](./docs/design-system.md).

Semantic-color rule (do not violate): brand color is amber (`primary`); positive money is teal (`income`); negative money is rust (`expense` / `destructive`). Never mix these uses — amber is **never** used to mean "income," and teal is **never** used for destructive actions. Don't reach for raw Tailwind palette colors (`bg-emerald-500`, `text-amber-400`, etc.); always go through theme tokens so dark mode and future palette tweaks keep working.

## Structure

Currently scaffolded:

```
frontend/
├── public/                 # static assets (favicon, icons)
├── src/
│   ├── api/
│   │   └── client.ts       # fetch wrapper with credentials: 'include'
│   ├── components/
│   │   └── layout/
│   │       └── Navbar.tsx  # top nav with route links
│   ├── lib/
│   │   ├── utils.ts        # cn() helper
│   │   ├── queryClient.ts  # TanStack Query config
│   │   └── theme.tsx       # ThemeProvider with light/dark/system
│   ├── pages/              # one file per route, all 8 stubbed
│   │   ├── Welcome.tsx
│   │   ├── Login.tsx
│   │   ├── Register.tsx
│   │   ├── Onboarding.tsx
│   │   ├── Dashboard.tsx
│   │   ├── Import.tsx
│   │   ├── Settings/index.tsx
│   │   └── NotFound.tsx
│   ├── types/
│   │   └── models.ts       # User, Account, Transaction, Category, ImportJob
│   ├── App.tsx             # ThemeProvider → QueryClientProvider → BrowserRouter
│   ├── main.tsx
│   ├── routes.tsx
│   └── index.css           # Tailwind + shadcn theme tokens
├── components.json         # shadcn/ui config
├── vite.config.ts          # port 3000, /api proxy → :8000, @ alias → /src
└── tsconfig.*              # TypeScript with @/* path alias
```

The `components/{ui,charts,transactions,accounts,shared}`, `hooks/`, and `fixtures/` directories from the architecture doc will be added as those features are built.

## Path alias

`@/*` resolves to `src/*` — use `import { cn } from '@/lib/utils'`, not relative paths.

## Conventions

- Components: `PascalCase` files in `src/components/`
- Hooks: `camelCase` prefixed with `use` in `src/hooks/`
- API calls: centralized in `src/api/`, one file per resource
- Tests: co-located with source (`ComponentName.test.tsx`)
- Env vars: must be prefixed with `VITE_` to be accessible in browser code
- TypeScript 6 has `erasableSyntaxOnly` enabled — no constructor parameter properties (`constructor(public x: T)`); declare fields explicitly
