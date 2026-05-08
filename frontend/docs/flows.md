# Frontend Flows

Sequence diagrams for the key user journeys, from the frontend's perspective. Backend internals are abstracted to "Backend API".

---

## 1. New User: Welcome → Register → Onboarding → Dashboard

```mermaid
sequenceDiagram
    participant U as User
    participant W as Welcome Page
    participant R as Register Page
    participant O as Onboarding
    participant D as Dashboard
    participant API as Backend API

    U->>W: Visits /
    W-->>U: Renders hero + interactive demo (sample data, no API calls)
    U->>W: Clicks "Get started"
    W->>R: Navigate to /register

    U->>R: Submits email + password
    R->>API: POST /auth/register
    API-->>R: 201 + Set-Cookie (httpOnly JWT) + user JSON
    R->>O: Navigate to /onboarding

    Note over O: Step 1 — Create first account
    U->>O: Fills account form, submits
    O->>API: POST /accounts
    API-->>O: 201 account

    Note over O: Step 2 — Pick categories
    U->>O: Toggles starter categories, adds custom, submits
    O->>API: POST /categories (one per selection)
    API-->>O: 201 each

    Note over O: Step 3 — Import prompt
    U->>O: Clicks "I'll do this later"
    O->>D: Navigate to /dashboard
    D->>API: GET /accounts, /transactions, /categories (parallel)
    API-->>D: Data (transactions empty)
    D-->>U: Renders dashboard with empty-state for transactions
```

---

## 2. Returning User: Login (Email/Password)

```mermaid
sequenceDiagram
    participant U as User
    participant L as Login Page
    participant D as Dashboard
    participant API as Backend API

    U->>L: Visits /login
    L->>API: GET /users/me (TanStack Query check)
    API-->>L: 401 (not logged in)
    L-->>U: Renders login form

    U->>L: Submits email + password
    L->>API: POST /auth/login (credentials: 'include')
    alt Invalid credentials
        API-->>L: 401
        L-->>U: Shows inline error
    else Success
        API-->>L: 200 + Set-Cookie + user JSON
        L->>L: queryClient.invalidateQueries(['me'])
        L->>D: Navigate to /dashboard (or ?next= param)
        D->>API: GET /users/me
        API-->>D: 200 user (cookie sent automatically)
        D-->>U: Renders dashboard
    end
```

---

## 3. OAuth Login (Google / GitHub)

OAuth flows happen mostly outside the React app — the browser performs full-page redirects. The backend sets the cookie before bouncing the user back.

```mermaid
sequenceDiagram
    participant U as User
    participant L as Login Page
    participant Browser as Browser
    participant API as Backend API
    participant Provider as OAuth Provider
    participant D as Dashboard

    U->>L: Clicks "Continue with Google"
    L-->>Browser: <a href="/api/v1/auth/oauth/google">

    Browser->>API: GET /auth/oauth/google
    API-->>Browser: 302 → Provider consent screen
    Browser->>Provider: GET consent URL
    Provider-->>U: Renders consent screen
    U->>Provider: Approves
    Provider-->>Browser: 302 → /auth/oauth/google/callback?code=...

    Browser->>API: GET /auth/oauth/google/callback?code=...
    API->>Provider: Exchange code for token, fetch profile
    Provider-->>API: User profile
    API->>API: Upsert user + oauth_connection
    API-->>Browser: 302 → /login?oauth=success + Set-Cookie

    Browser->>L: Loads /login?oauth=success
    L->>L: Detects ?oauth=success param
    L->>API: GET /users/me (cookie sent)
    API-->>L: 200 user
    L->>D: Navigate to /dashboard
```

---

## 4. Protected Route Guard

```mermaid
sequenceDiagram
    participant U as User
    participant R as Router
    participant PR as ProtectedRoute
    participant API as Backend API
    participant L as Login Page
    participant D as Dashboard

    U->>R: Navigates to /dashboard
    R->>PR: Renders <ProtectedRoute>
    PR->>API: GET /users/me (via useAuth())
    alt Authenticated
        API-->>PR: 200 user
        PR->>D: Renders <Dashboard />
    else Not authenticated
        API-->>PR: 401
        PR->>L: Redirect to /login?next=/dashboard
        Note over L: User logs in
        L->>D: After login, reads ?next=, navigates to /dashboard
    end
```

---

## 5. CSV Import with Job Polling

```mermaid
sequenceDiagram
    participant U as User
    participant I as Import Page
    participant API as Backend API
    participant Q as TanStack Query

    U->>I: Selects account, drops CSV file
    I->>API: POST /import/csv (multipart/form-data)
    API-->>I: 202 { job_id, status: 'pending' }
    I->>I: Show <JobProgressCard> with job_id

    loop Every 2 seconds while status in [pending, processing]
        Q->>API: GET /import/jobs/:id
        API-->>Q: { status, rows_imported, rows_total, ... }
        Q-->>I: Updated job state
        I-->>U: Updated progress bar
    end

    alt Status = done
        Q->>API: GET /import/jobs/:id (final)
        API-->>Q: { status: 'done', rows_imported: 45 }
        I-->>U: Toast: "45 transactions imported"
        I-->>U: Hide progress card, show in history table
    else Status = failed
        I-->>U: Toast (error variant): "Import failed"
    end
```

---

## 6. Theme Toggle

```mermaid
sequenceDiagram
    participant U as User
    participant PM as ProfileMenu
    participant TP as ThemeProvider
    participant LS as localStorage
    participant DOM as document.documentElement

    U->>PM: Clicks "Theme → Dark"
    PM->>TP: setTheme('dark')
    TP->>LS: localStorage.setItem('theme', 'dark')
    TP->>DOM: classList.add('dark')
    DOM-->>U: All Tailwind dark: variants take effect
```

On page load, `<ThemeProvider>` reads `localStorage` (or system preference) and applies the class before paint to avoid a flash of unstyled content.

---

## 7. Logout

```mermaid
sequenceDiagram
    participant U as User
    participant PM as ProfileMenu
    participant API as Backend API
    participant Q as TanStack Query
    participant W as Welcome Page

    U->>PM: Clicks "Log out"
    PM->>API: POST /auth/logout
    API-->>PM: 204 + Set-Cookie (cleared)
    PM->>Q: queryClient.clear() — wipe all cached queries
    PM->>W: Navigate to /
    W-->>U: Renders welcome (logged-out state)
```

---

## 8. Welcome Page Demo Preview (No API)

The interactive preview on `/` demonstrates the dashboard without requiring a backend.

```mermaid
sequenceDiagram
    participant U as User
    participant W as Welcome Page
    participant DP as <DemoPreview />
    participant Fixture as fixtures/demo-data.ts
    participant Charts as Chart Components

    U->>W: Visits /
    W->>DP: Renders <DemoPreview />
    DP->>Fixture: import demoAccounts, demoTransactions, demoCategories
    Fixture-->>DP: Static fake data
    DP->>Charts: Pass fake data as props
    Charts-->>U: Renders donut + bar charts populated with fake data

    Note over DP: Pointer-events: none on interactive bits — read-only preview
```

The same chart, table, and account-card components from `/dashboard` are reused — they take their data as props rather than fetching, so they work for both real and demo contexts.
