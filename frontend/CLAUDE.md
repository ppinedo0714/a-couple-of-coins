# frontend

React + TypeScript UI. Communicates with the backend API only — never directly with Python services.

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
| `docs/architecture.md` | Stack, directory layout, routing, data layer, auth strategy, theming |
| `docs/pages.md` | Every page with purpose, components, data dependencies, behavior |
| `docs/components.md` | Shared components: layout, charts, transactions, accounts, primitives |
| `docs/flows.md` | Sequence diagrams for register/login, OAuth, protected routes, CSV import |

## Stack

- React + TypeScript (Vite)
- Tailwind CSS + shadcn/ui
- TanStack Query for server state
- React Router for routing
- Recharts for visualizations
- Auth: httpOnly cookies (set by backend) — fetches use `credentials: 'include'`

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
