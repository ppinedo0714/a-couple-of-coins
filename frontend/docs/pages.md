# Pages Reference

Every page in the app, in route order. For each: route, auth requirement, purpose, data, and key components.

---

## `/` — Welcome

**Auth:** Public. If user is already authenticated, redirect to `/dashboard`.

**Purpose:** First impression. Sells the app to a new visitor and provides a path to sign up or log in.

**Sections (top to bottom):**
1. **Hero** — Product name, one-line tagline, primary CTA ("Get started" → `/register`) and secondary CTA ("Log in" → `/login`)
2. **Interactive preview** — A read-only, fake-data version of the dashboard. Uses fixtures from `src/fixtures/demo-data.ts`. The same `<Dashboard>` components render — they accept their data as props so they can be reused here without API calls.
3. **Feature highlights** — 3–4 cards with icon + title + short description (transaction tracking, smart categorization, multi-account, secure)
4. **Footer** — Links, GitHub, etc.

**Data:** None from API. Sample data is bundled at build time.

**Components:** `<HeroSection>`, `<DemoPreview>` (wraps real dashboard components), `<FeatureGrid>`, `<Footer>`

---

## `/login` — Login

**Auth:** Public. Authenticated users redirect to `/dashboard`.

**Purpose:** Email/password login + OAuth.

**Layout:**
- Centered card on a plain background
- Email input, password input, "Log in" submit button
- Divider with "or"
- Two OAuth buttons: "Continue with Google", "Continue with GitHub"
- Link below: "Don't have an account? Sign up"

**Behavior:**
- On submit, calls `POST /auth/login`. On `200`, navigate to `/dashboard`. On `401`, show inline error.
- OAuth buttons are anchor tags pointing at `/api/v1/auth/oauth/google` (or `/github`). The backend handles the redirect dance and sets the cookie before bouncing back.
- After successful OAuth, the backend redirects to `/login?oauth=success`. The page detects this and navigates to `/dashboard` (or `/onboarding` if first login — see below).

**Components:** `<AuthCard>`, `<OAuthButton>`, form via `react-hook-form` + `zod`

---

## `/register` — Registration

**Auth:** Public.

**Purpose:** Create a new account with email + password (or OAuth).

**Layout:** Same shape as login, with these differences:
- Email + password + "confirm password" inputs
- Submit button: "Create account"
- Same OAuth buttons (registration via OAuth happens automatically on first sign-in)
- Link: "Already have an account? Log in"

**Validation (Zod schema):**
- Email: valid email format
- Password: min 8 chars
- Confirm password: must match password

**Behavior:** On `201`, navigate to `/onboarding`. On `409` (email taken), show inline error.

**Components:** `<AuthCard>`, `<OAuthButton>`

---

## `/onboarding` — First-Time Wizard

**Auth:** Protected.

**Purpose:** Get a brand-new user from "empty account" to "useful state" in 3 steps. Reachable only via post-registration redirect; users can skip the entire flow.

**Steps:**
1. **Create your first account** — Form: name, type (dropdown: checking/savings/credit/investment), starting balance, currency. Submit creates the account. Skip → step 2.
2. **Add some categories** — Pre-suggest a small starter set (Groceries, Rent, Transport, Salary, Entertainment) as toggle chips. User picks which to create + can add custom ones inline. Submit creates them in bulk. Skip → step 3.
3. **Import your transactions** — Two buttons: "Upload CSV" (links to `/import`), "I'll do this later" (links to `/dashboard`).

A progress indicator at the top shows step 1/3, 2/3, 3/3. Each step has a "Skip" link that advances without action.

After step 3 (or any skip), navigate to `/dashboard`.

**Data:** Calls `POST /accounts` and `POST /categories` (in batch). No state stored across steps beyond what's saved.

**Components:** `<OnboardingShell>` (progress + skip), `<Step1AccountForm>`, `<Step2CategoryPicker>`, `<Step3ImportPrompt>`

---

## `/dashboard` — Main View

**Auth:** Protected.

**Purpose:** The user's home base. Shows current financial snapshot and recent activity.

**Layout (top to bottom on mobile, grid on desktop):**

1. **Accounts overview** — A row of account cards, each showing name, type, and current balance. A "+" card at the end opens the "create account" dialog.
2. **Spending charts** (side-by-side on desktop, stacked on mobile)
   - **By category** — donut chart of expenses for the selected period, grouped by category
   - **Over time** — bar chart of daily/weekly spending for the selected period
3. **Category breakdown** — Table or list: category name | total spent | % of total | budget bar (future). Click a row → filtered transactions view.
4. **Recent transactions** — Table with columns: date, description, merchant, category (badge), account, amount. Filters: account, category, date range, search. Pagination.

**Period selector** in the page header: This Month / Last Month / Last 3 Months / Custom range. Stored in URL query params so links are shareable.

**Data:**
- `GET /accounts` (cards)
- `GET /transactions?from=...&to=...` (table + chart inputs)
- `GET /categories` (badge labels and breakdown)

The charts and breakdown derive their data client-side from the transactions response — one fetch, multiple visualizations.

**Components:** `<AccountList>`, `<SpendingByCategory>`, `<SpendingOverTime>`, `<CategoryBreakdown>`, `<TransactionTable>`, `<PeriodSelector>`

---

## `/import` — Import

**Auth:** Protected.

**Purpose:** Bring transactions in from external sources.

**Layout:**

Two clearly separated sections, vertically stacked:

1. **Link a bank account** *(disabled placeholder for now — Plaid integration TBD)*
   - Card with explanation, "Connect a bank" button (disabled with tooltip "Coming soon")

2. **Upload a CSV**
   - Account selector (which account these transactions belong to)
   - Drag-and-drop file upload zone (also clickable to open file picker)
   - Expected format hint: "CSV with columns: date, description, amount"
   - On upload: shows in-progress state, then summary on success

3. **Import history** (below the two sections)
   - Table: file name, account, status, rows imported, date. Auto-refreshes via TanStack Query while any job has status `pending` or `processing`.

**Behavior:**
- Drag-drop or pick a file → `POST /import/csv` with `multipart/form-data`
- Returns a `job_id`. UI shows a progress card that polls `GET /import/jobs/:id` every 2s while in `pending` or `processing`.
- On `done`: success toast with row count + link to view classified transactions
- On `failed`: error toast with message

**Data:** `GET /accounts` (selector), `GET /import/jobs` (history), `POST /import/csv`, `GET /import/jobs/:id` (polling)

**Components:** `<BankConnectCard>` (placeholder), `<CsvUploadZone>`, `<ImportJobsTable>`, `<JobProgressCard>`

---

## `/settings` — Settings (Tabbed)

**Auth:** Protected.

**Purpose:** Manage accounts, categories, and profile.

**Layout:** Three tabs along the top.

### Tab 1: Accounts
- Table of accounts: name, type, balance, created date, edit/delete buttons
- "Add account" button → modal with the same form as the onboarding step
- Edit → modal pre-filled with current values
- Delete → confirmation dialog. Backend rejects with `409` if account has transactions.

### Tab 2: Categories
- Grid of category chips (color-coded). Click a chip to edit.
- "Add category" button → small modal: name + color picker
- Delete → confirmation. Affected transactions get `category_id = null` automatically (backend behavior).

### Tab 3: Profile
- Email field (editable)
- "Update email" button → calls `PUT /users/me`
- Theme toggle (light / dark / system)
- "Log out" button at the bottom
- "Delete account" (with confirmation) — *future*

**Data:** `GET /accounts`, `GET /categories`, `GET /users/me`, plus their respective mutations.

**Components:** `<TabsContainer>`, `<AccountsTab>`, `<CategoriesTab>`, `<ProfileTab>`, `<AccountFormDialog>`, `<CategoryFormDialog>`

---

## `/404` — Not Found

**Auth:** Public.

**Purpose:** Friendly fallback for unknown routes.

**Layout:** Centered illustration or icon, "404 — Page not found" headline, short message, button: "Go home" → `/` (or `/dashboard` if logged in).

**Data:** None.

**Components:** `<NotFoundCard>`
