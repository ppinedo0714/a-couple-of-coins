# Frontend Architecture

## System Overview

A React single-page application that consumes the Go backend API. The frontend never talks to the Python prediction service directly — all calls go through the backend.

```
┌─────────────┐        ┌─────────────────┐
│   frontend  │ ──────▶│    backend API  │
│  React + TS │  HTTP  │      Go         │
│  port: 3000 │        │   port: 8000    │
└─────────────┘        └─────────────────┘
        ↑
        │ Tailwind + shadcn/ui
        │ TanStack Query (server state)
        │ React Router (routing)
        │ Recharts (visualizations)
```

## Tech Stack

| Concern | Choice | Reason |
|---------|--------|--------|
| Build tool | Vite | Fast HMR, minimal config |
| Language | TypeScript (strict) | Type safety mirrors backend contract |
| Styling | Tailwind CSS | Utility-first, dark mode built in |
| Components | shadcn/ui | Copy-pasted, customizable, owns the source |
| Routing | React Router v6 | Standard, supports loaders + protected routes |
| Server state | TanStack Query | Caching, refetching, loading/error states |
| Client state | React Context + useState | No global store needed at this scale |
| Charts | Recharts | Declarative, React-first |
| Forms | React Hook Form + Zod | Schema validation matches backend types |

## Directory Layout

```
frontend/
├── public/                      ← static assets, favicon
├── src/
│   ├── api/                     ← typed API client (one file per resource)
│   │   ├── client.ts            ← fetch wrapper with credentials: 'include'
│   │   ├── auth.ts
│   │   ├── accounts.ts
│   │   ├── categories.ts
│   │   ├── transactions.ts
│   │   └── imports.ts
│   ├── components/
│   │   ├── ui/                  ← shadcn/ui primitives (Button, Input, Dialog, ...)
│   │   ├── layout/              ← Navbar, ProfileMenu, PageWrapper, ProtectedRoute
│   │   ├── charts/              ← Recharts wrappers
│   │   ├── transactions/        ← TransactionTable, TransactionFilters
│   │   ├── accounts/            ← AccountCard, AccountList
│   │   └── shared/              ← ThemeToggle, EmptyState, LoadingSpinner
│   ├── pages/
│   │   ├── Welcome.tsx
│   │   ├── Login.tsx
│   │   ├── Register.tsx
│   │   ├── Onboarding.tsx
│   │   ├── Dashboard.tsx
│   │   ├── Import.tsx
│   │   ├── Settings/
│   │   │   ├── index.tsx        ← tab container
│   │   │   ├── AccountsTab.tsx
│   │   │   ├── CategoriesTab.tsx
│   │   │   └── ProfileTab.tsx
│   │   └── NotFound.tsx
│   ├── hooks/                   ← TanStack Query hooks, one file per resource
│   │   ├── useAuth.ts
│   │   ├── useAccounts.ts
│   │   ├── useCategories.ts
│   │   ├── useTransactions.ts
│   │   ├── useImports.ts
│   │   └── useTheme.ts
│   ├── lib/
│   │   ├── queryClient.ts       ← TanStack Query default config
│   │   ├── theme.tsx            ← dark/light mode provider
│   │   └── utils.ts             ← cn() helper, formatters
│   ├── fixtures/
│   │   └── demo-data.ts         ← fake accounts/transactions for welcome preview
│   ├── types/
│   │   ├── api.ts               ← request/response shapes
│   │   └── models.ts            ← User, Account, Transaction, Category, etc.
│   ├── routes.tsx               ← React Router config
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css                ← Tailwind directives
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.js
├── postcss.config.js
├── components.json              ← shadcn/ui config
└── CLAUDE.md
```

## Routing Structure

Routes split into public and protected groups. Protected routes redirect to `/login` if the user is not authenticated.

```
Public:
  /                  Welcome
  /login             Login
  /register          Register
  /404               Not Found

Protected (require auth):
  /onboarding        First-time wizard
  /dashboard         General info / main view
  /import            CSV + bank import
  /settings          Tabs: accounts, categories, profile
```

Protection is enforced by a `<ProtectedRoute>` wrapper in `components/layout/`. It calls `useAuth()` (TanStack Query against `GET /users/me`) and redirects on 401.

## Auth Strategy — httpOnly Cookies

The JWT is stored in an `httpOnly` `Secure` `SameSite=Lax` cookie set by the backend on login/register. The frontend **never** reads or writes the token directly.

**How auth works in the frontend:**
1. All `fetch` calls in `src/api/client.ts` use `credentials: 'include'` — the browser sends the cookie automatically.
2. `useAuth()` calls `GET /users/me` to determine if the user is logged in. A `200` response means authenticated; `401` means not.
3. Login/register/OAuth-callback responses contain the `User` object only — no token in JSON.
4. Logout calls `POST /auth/logout` which clears the cookie server-side.

**Implications:**
- Frontend and backend must be on the same origin in production, or the backend must explicitly enable CORS with `Access-Control-Allow-Credentials: true` and an exact `Access-Control-Allow-Origin` (no wildcard).
- In development with Vite, configure a proxy in `vite.config.ts` so `/api/*` requests go to `localhost:8000` from the frontend's origin.

## Data Layer — TanStack Query

All server state is owned by TanStack Query. No useState/useEffect for API data.

**Conventions:**
- One hook file per resource: `useAccounts.ts`, `useTransactions.ts`, etc.
- Each hook exports `useXList`, `useX`, `useCreateX`, `useUpdateX`, `useDeleteX`.
- Query keys follow the shape `['resource', ...filters]` — e.g. `['transactions', { accountId, from, to }]`.
- Mutations call `queryClient.invalidateQueries(['resource'])` on success.

Default config (in `lib/queryClient.ts`):
- `staleTime: 30s` — avoid refetching too aggressively
- `retry: 1` — retry failed requests once
- 401 responses don't retry; they redirect to `/login`

## Client State

Two pieces of client state, both via React Context:
- **Theme** (`lib/theme.tsx`) — light/dark/system, persisted in localStorage
- **Auth user** (`hooks/useAuth.ts`) — derived from `GET /users/me`; not separately stored

Everything else is local component state.

## Theming

Tailwind's `darkMode: 'class'` strategy. The `<html>` element gets a `dark` class applied by the theme provider. shadcn/ui components respect this automatically via CSS variables.

The toggle lives in the profile dropdown (`components/layout/ProfileMenu.tsx`).

## Mobile Responsiveness

- Tailwind breakpoints: `sm` (640px), `md` (768px), `lg` (1024px)
- Default styles are mobile-first; responsive classes scale up
- Navbar collapses to a hamburger menu below `md`
- Dashboard charts stack vertically on mobile, grid on desktop
- Tables become card-style lists below `sm` (or scroll horizontally)

## Environment Variables

Vite requires the `VITE_` prefix for env vars exposed to the browser.

| Variable | Description |
|----------|-------------|
| `VITE_API_BASE_URL` | Backend base URL, e.g. `http://localhost:8000` (proxied in dev via Vite) |
