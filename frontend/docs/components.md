# Components Reference

Shared, reusable components organized by category. Page-specific one-offs are not listed here — only components used in 2+ places or that own non-trivial behavior.

---

## Layout

### `<Navbar />`
Top-of-page navigation bar shown on all authenticated pages. Hidden on Welcome / Login / Register.

- Left: app logo + name → links to `/dashboard`
- Center (desktop only): nav links — Dashboard, Import
- Right: `<ProfileMenu />`
- Below `md` breakpoint: collapses to hamburger button that opens a drawer

**Path:** `components/layout/Navbar.tsx`

---

### `<ProfileMenu />`
The top-right profile icon and dropdown.

**States:**
- **Logged out:** Renders a "Log in" button → `/login`
- **Logged in:** Renders an avatar (initials from email) → dropdown with:
  - User email (header, non-clickable)
  - Settings → `/settings`
  - Theme toggle (sub-menu: Light / Dark / System)
  - Log out → calls `useAuth().logout()`, redirects to `/`

**Auth state source:** `useAuth()` hook.

**Path:** `components/layout/ProfileMenu.tsx`

---

### `<PageWrapper />`
Standard page container — sets max width, padding, and renders the navbar above its children.

**Props:**
- `title?: string` — optional page heading
- `actions?: ReactNode` — buttons rendered next to the heading (e.g., "Add account")
- `children: ReactNode`

**Path:** `components/layout/PageWrapper.tsx`

---

### `<ProtectedRoute />`
Route guard for authenticated pages. Wraps a route element.

**Behavior:**
- Calls `useAuth()`. While loading, renders a skeleton. If authenticated, renders children. If not, redirects to `/login` and preserves the intended URL in `?next=`.
- After login, the login page reads `?next=` and redirects there.

**Path:** `components/layout/ProtectedRoute.tsx`

---

## Charts

All chart components accept transaction data + a date range and produce a fully styled Recharts visualization. They are used both on `/dashboard` and inside `<DemoPreview />` on the welcome page.

### `<SpendingByCategory />`
Donut chart of expenses grouped by category for the period.

**Props:**
- `transactions: Transaction[]`
- `categories: Category[]`

Internally: filters to `amount < 0`, groups by `category_id`, computes totals, applies category colors.

**Path:** `components/charts/SpendingByCategory.tsx`

---

### `<SpendingOverTime />`
Bar chart of daily or weekly net spending.

**Props:**
- `transactions: Transaction[]`
- `granularity: 'day' | 'week' | 'month'` (auto-selected based on range)

**Path:** `components/charts/SpendingOverTime.tsx`

---

### `<CategoryBreakdown />`
Tabular breakdown showing each category, total spent, and percentage of total. Each row is clickable → filters the transaction table to that category.

**Props:**
- `transactions: Transaction[]`
- `categories: Category[]`
- `onCategoryClick?: (categoryId: string) => void`

**Path:** `components/charts/CategoryBreakdown.tsx`

---

## Transactions

### `<TransactionTable />`
The main transaction list. Used on `/dashboard` and as a sub-view in the category drilldown.

**Columns:** date, description, merchant, category (color-coded badge), account, amount (color: red for expense, green for income)

**Features:**
- Sortable column headers (date default, descending)
- Inline category editor — click the badge to open a popover with category picker
- Pagination (50 rows per page, server-paginated via `limit`/`offset` query params)
- Empty state when no rows: "No transactions yet — try importing a CSV"

**Props:**
- `filters: TransactionFilters` (account, category, date range, search)
- `onFiltersChange?: (filters) => void`

**Path:** `components/transactions/TransactionTable.tsx`

---

### `<TransactionFilters />`
Filter controls above `<TransactionTable />`.

- Account dropdown (multi-select)
- Category dropdown (multi-select)
- Date range picker
- Search input (debounced, hits `?search=`)

**Path:** `components/transactions/TransactionFilters.tsx`

---

## Accounts

### `<AccountCard />`
Single account display. Used in the dashboard accounts row and in the settings accounts table.

**Props:**
- `account: Account`
- `variant: 'card' | 'row'`
- `onEdit?: () => void`

Shows name, type icon, current balance (formatted with currency).

**Path:** `components/accounts/AccountCard.tsx`

---

### `<AccountList />`
Horizontally-scrolling row of `<AccountCard variant="card" />` for the dashboard. Includes an "+" card at the end that opens the account-create dialog.

**Path:** `components/accounts/AccountList.tsx`

---

### `<AccountFormDialog />`
Modal for creating or editing an account.

**Props:**
- `account?: Account` — if present, dialog is in edit mode
- `open: boolean`
- `onClose: () => void`

Uses react-hook-form + zod schema.

**Path:** `components/accounts/AccountFormDialog.tsx`

---

## Shared

### `<ThemeToggle />`
Three-option toggle: Light / Dark / System. Used in the profile dropdown and on the welcome page footer.

Reads/writes via `useTheme()` from `lib/theme.tsx`.

**Path:** `components/shared/ThemeToggle.tsx`

---

### `<EmptyState />`
Friendly empty-state placeholder for tables and lists.

**Props:**
- `icon?: ReactNode`
- `title: string`
- `description?: string`
- `action?: { label: string; href?: string; onClick?: () => void }`

**Path:** `components/shared/EmptyState.tsx`

---

### `<LoadingSpinner />` and `<Skeleton />`
Skeleton placeholders for loading states. `<Skeleton />` is the shadcn/ui primitive — composed into per-page skeletons (e.g. `<DashboardSkeleton />`).

**Path:** `components/shared/LoadingSpinner.tsx`, shadcn `<Skeleton />` in `components/ui/skeleton.tsx`

---

### `<OAuthButton />`
Branded button for "Continue with Google" / "Continue with GitHub". Renders the provider's icon + label.

**Props:**
- `provider: 'google' | 'github'`

Renders as an `<a>` pointing at `/api/v1/auth/oauth/{provider}` so the browser performs the redirect (cookies must be set by the backend).

**Path:** `components/shared/OAuthButton.tsx`

---

## shadcn/ui Primitives

These are copied into `components/ui/` from the shadcn registry as needed. They are owned by this project — modify freely. Primitives we'll need:

| Primitive | Used by |
|-----------|---------|
| `button` | Everywhere |
| `input`, `label`, `form` | Auth pages, settings, dialogs |
| `card` | Welcome, dashboard, settings |
| `dialog` | Account / category form modals |
| `dropdown-menu` | ProfileMenu |
| `select`, `popover` | Filters, category picker |
| `table` | TransactionTable, ImportJobsTable |
| `tabs` | Settings page |
| `toast` (sonner) | All success/error notifications |
| `skeleton` | Loading states |
| `badge` | Category labels |
| `avatar` | ProfileMenu |
| `tooltip` | Disabled buttons, info icons |
