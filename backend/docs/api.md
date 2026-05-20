# API Reference

Base URL: `http://localhost:8000/api/v1`

## Authentication

The API uses **JWT stored in an httpOnly cookie**, set by the backend on successful login or registration.

- Login/register/OAuth-callback responses include `Set-Cookie: token=<jwt>; HttpOnly; Secure; SameSite=Lax; Path=/`.
- Authenticated requests must include the cookie. Browsers send it automatically when fetches use `credentials: 'include'`.
- The token is **not** returned in JSON response bodies.
- CORS is configured with `Access-Control-Allow-Credentials: true` and an exact `Access-Control-Allow-Origin` (no wildcard).
- Logout clears the cookie via `Set-Cookie: token=; Max-Age=0`.

All request and response bodies are JSON. All timestamps are ISO 8601 (`2024-01-15T10:30:00Z`). All IDs are UUIDs.

**Date fields** (e.g. `from`, `to`, transaction `date`) use `YYYY-MM-DD` calendar dates with no timezone component — treat as UTC midnight.

## Error Responses

All error responses use the same JSON shape:

```json
{
  "error": "human-readable message"
}
```

| Status | Meaning |
|--------|---------|
| `400` | Validation failed — malformed JSON, missing required field, invalid enum value |
| `401` | Not authenticated — missing or invalid auth cookie |
| `404` | Resource not found or not owned by the authenticated user |
| `409` | Conflict — e.g. duplicate email, account has transactions, duplicate category name |
| `500` | Internal server error — logged server-side; generic message returned to client |

---

## Auth

### `POST /auth/register`
Create a new account with email and password.

**Request**
```json
{
  "email": "user@example.com",
  "password": "minimum8chars"
}
```

**Response `201`** — sets `Set-Cookie: token=<jwt>; HttpOnly; Secure; SameSite=Lax`
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Errors:** `400` invalid input, `409` email already registered

---

### `POST /auth/login`
Authenticate with email and password.

**Request**
```json
{
  "email": "user@example.com",
  "password": "minimum8chars"
}
```

**Response `200`** — sets `Set-Cookie: token=<jwt>; HttpOnly; Secure; SameSite=Lax`
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Errors:** `400` invalid input, `401` wrong credentials

---

### `GET /auth/oauth/google`
Redirects the browser to Google's OAuth consent screen.

**Response:** `302` redirect

---

### `GET /auth/oauth/google/callback`
Google redirects here after the user grants permission. Sets the auth cookie and bounces back to the frontend.

**Query params:** `code`, `state`

**Response `302`** — redirects to `<frontend-url>/login?oauth=success` with `Set-Cookie: token=<jwt>; HttpOnly; Secure; SameSite=Lax`

---

### `GET /auth/oauth/github`
Redirects to GitHub's OAuth consent screen.

**Response:** `302` redirect

---

### `GET /auth/oauth/github/callback`
GitHub redirects here after authorization. Same behavior as the Google callback.

**Response `302`** — redirects to `<frontend-url>/login?oauth=success` with `Set-Cookie`

---

### `POST /auth/logout`
Auth required. Clears the auth cookie.

**Response `204`** — no body, sets `Set-Cookie: token=; Max-Age=0`

---

## Users

### `GET /users/me`
Auth required. Returns the current user's profile.

**Response `200`**
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

### `PUT /users/me`
Auth required. Update the current user's profile.

**Request** — all fields optional
```json
{
  "email": "newemail@example.com"
}
```

**Response `200`** — updated user object (same shape as `GET /users/me`)

**Errors:** `409` email taken

---

## Accounts

### `GET /accounts`
Auth required. List all accounts for the current user.

**Response `200`**
```json
[
  {
    "id": "uuid",
    "name": "Chase Checking",
    "type": "checking",
    "balance": 2450.00,
    "currency": "USD",
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

---

### `POST /accounts`
Auth required. Create a new account.

**Request**
```json
{
  "name": "Chase Checking",
  "type": "checking",
  "balance": 0.00,
  "currency": "USD"
}
```

`type` must be one of: `checking`, `savings`, `credit`, `investment`

**Response `201`** — created account object

**Errors:** `400` invalid input

---

### `GET /accounts/:id`
Auth required. Get a single account.

**Response `200`** — account object

**Errors:** `404` not found or not owned by user

---

### `PUT /accounts/:id`
Auth required. Update an account.

**Request** — all fields optional
```json
{
  "name": "Chase Checking (new name)",
  "type": "savings"
}
```

**Response `200`** — updated account object

**Errors:** `400`, `404`

---

### `DELETE /accounts/:id`
Auth required. Delete an account. Fails if the account has transactions.

**Response `204`** — no body

**Errors:** `404`, `409` account has transactions

---

### `GET /accounts/history`
Auth required. Return balance snapshots for one or more accounts over a date range.

**Query params**

| Param | Type | Description |
|-------|------|-------------|
| `account_ids` | string | Comma-separated UUIDs (optional; default: all user accounts) |
| `from` | date | Start date inclusive, `YYYY-MM-DD` (required) |
| `to` | date | End date inclusive, `YYYY-MM-DD` (required) |
| `interval` | string | `day` \| `week` \| `month` — granularity of returned snapshots (optional; default: `day`) |

For `week` and `month` intervals, the last available snapshot within each period is returned. Snapshots are ordered by date ascending. Only snapshots belonging to the authenticated user's accounts are returned.

**Response `200`**
```json
{
  "snapshots": [
    { "date": "2025-01-06", "account_id": "uuid", "balance": 2100.00 },
    { "date": "2025-01-13", "account_id": "uuid", "balance": 2350.00 }
  ]
}
```

**Errors:** `400` missing or invalid `from`/`to`

---

## Categories

Categories form a two-level hierarchy. Rows with `parent_id: null` are **Groups** (e.g. Entertainment); rows with a `parent_id` are **Categories** (e.g. Movies). A transaction's `category_id` may reference either level.

### `GET /categories`
Auth required. List all groups and categories for the current user.

**Response `200`**
```json
[
  {
    "id": "uuid",
    "name": "Entertainment",
    "color": "#EC407A",
    "parent_id": null,
    "created_at": "2024-01-15T10:30:00Z"
  },
  {
    "id": "uuid",
    "name": "Movies",
    "color": null,
    "parent_id": "<entertainment-group-id>",
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

---

### `POST /categories`
Auth required. Create a group or a category.

**Request**
```json
{
  "name": "Movies",
  "parent_id": "<group-uuid>"
}
```

- Omit `parent_id` (or pass `null`) to create a **Group**; include `color` for the group color.
- Include `parent_id` to create a **Category** under that group; `color` is ignored (Categories inherit their group's color).

**Response `201`** — created object

**Errors:** `400` invalid input or `parent_id` references a non-group, `409` name already exists within the same scope

---

### `PUT /categories/:id`
Auth required. Update a group or category.

**Request** — all fields optional
```json
{
  "name": "Streaming",
  "color": "#FF9800"
}
```

`color` is only applied when updating a Group (`parent_id IS NULL`); ignored for Categories.

**Response `200`** — updated object

---

### `DELETE /categories/:id`
Auth required. Delete a group or category.

- Deleting a **Group**: child categories have their `parent_id` set to null (they become top-level groups).
- Deleting a **Category**: transactions referencing it have `category_id` set to null.

**Response `204`** — no body

**Errors:** `404`

---

## Transactions

### `GET /transactions`
Auth required. List transactions with optional filters.

**Query params**

| Param | Type | Description |
|-------|------|-------------|
| `account_id` | uuid | Filter by account |
| `category_id` | uuid | Filter by category |
| `from` | date | Start date inclusive (`2024-01-01`) |
| `to` | date | End date inclusive (`2024-01-31`) |
| `search` | string | Full-text search on `description` and `merchant_name` |
| `unclassified` | bool | If `true`, return only unclassified transactions |
| `limit` | int | Page size (default 50, max 200) |
| `offset` | int | Pagination offset (default 0) |

**Response `200`**
```json
{
  "transactions": [
    {
      "id": "uuid",
      "account_id": "uuid",
      "category_id": "uuid or null",
      "amount": -42.50,
      "description": "WHOLE FOODS MARKET #123",
      "merchant_name": "Whole Foods",
      "date": "2024-01-15",
      "source": "csv",
      "classified": true,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 143,
  "limit": 50,
  "offset": 0
}
```

---

### `POST /transactions`
Auth required. Create a transaction manually.

**Request**
```json
{
  "account_id": "uuid",
  "category_id": "uuid",
  "amount": -42.50,
  "description": "Whole Foods",
  "date": "2024-01-15"
}
```

`category_id` is optional. `amount` is signed (negative = expense).

The backend sets `source = "manual"` and `classified = true` on creation — the user has already described the transaction, so it does not need prediction-service processing. `merchant_name` is left null for manually-created transactions. These fields are not accepted from the request body.

After insert, the backend updates `accounts.balance` by the transaction amount and upserts a row in `account_balance_snapshots` for the transaction date.

**Response `201`** — created transaction object

**Errors:** `400`, `404` account not found

---

### `GET /transactions/:id`
Auth required. Get a single transaction.

**Response `200`** — transaction object

**Errors:** `404`

---

### `PUT /transactions/:id`
Auth required. Update a transaction. Common use: assign or change a category.

**Request** — all fields optional
```json
{
  "category_id": "uuid",
  "description": "Updated description",
  "amount": -45.00,
  "date": "2024-01-16"
}
```

Pass `"category_id": null` explicitly to remove the category assignment.

If `amount` changes, the backend adjusts `accounts.balance` by the delta and upserts snapshots for both the old and new transaction dates. If `date` changes, the snapshot for the old date is recalculated from remaining transactions on that day.

**Response `200`** — updated transaction object

**Errors:** `400`, `404`

---

### `DELETE /transactions/:id`
Auth required. Delete a transaction. Updates the parent account balance.

**Response `204`** — no body

**Errors:** `404`

---

### `POST /transactions/classify`
Auth required. Sends all unclassified transactions (for the current user) to the prediction service. The prediction service returns a `category_id` from the user's existing categories and a normalized `merchant_name` for each transaction.

**Request** — no body required (classifies all unclassified transactions for the user)

The backend:
1. Fetches all transactions where `classified = false` for the user
2. Fetches all categories owned by the user
3. POSTs both to `services/predictions` (see [Prediction Service Internal Contract](#prediction-service-internal-contract))
4. Writes back the returned `category_id` and `merchant_name` and sets `classified = true` for each matched transaction

**Response `200`**
```json
{
  "classified": 12,
  "failed": 0
}
```

---

## Imports

### `POST /import/csv`
Auth required. Upload a CSV file to import transactions.

**Request** — `multipart/form-data`

| Field | Type | Description |
|-------|------|-------------|
| `file` | file | The CSV file |
| `account_id` | string | UUID of the account to import into |

Expected CSV columns (flexible — mapped during parsing):
`date`, `description`, `amount`

**Response `202`**
```json
{
  "job_id": "uuid",
  "status": "pending"
}
```

---

### `GET /import/jobs`
Auth required. List all import jobs for the current user, most recent first.

**Response `200`**
```json
[
  {
    "id": "uuid",
    "status": "done",
    "source_type": "csv",
    "file_name": "transactions_jan.csv",
    "rows_total": 45,
    "rows_imported": 45,
    "created_at": "2024-01-15T10:30:00Z",
    "completed_at": "2024-01-15T10:30:05Z"
  }
]
```

---

### `GET /import/jobs/:id`
Auth required. Get the status of a specific import job.

**Response `200`** — single import job object (same shape as list item above)

**Errors:** `404`

---

## Health

### `GET /health`
No auth required. Used for uptime checks.

**Response `200`**
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

---

## Prediction Service Internal Contract

This section documents the HTTP contract between the Go backend and the Python prediction service (`services/predictions`, port 8001). The frontend never calls this service directly.

### `POST /classify`

The backend sends unclassified transactions alongside the user's full category list. The service returns a best-matching `category_id` from that list for each transaction.

**Request**
```json
{
  "transactions": [
    {
      "id": "uuid",
      "description": "WHOLE FOODS MARKET #123",
      "amount": -42.50
    }
  ],
  "categories": [
    {
      "id": "uuid",
      "name": "Groceries",
      "parent_name": "Food"
    }
  ]
}
```

- `categories` contains all of the user's categories (both Groups and Categories). `parent_name` is the Group name for a Category, or `null` for a Group row.
- Transactions with no matching category should have `category_id` omitted or `null` in the response.

**Response `200`**
```json
{
  "predictions": [
    {
      "transaction_id": "uuid",
      "category_id": "uuid",
      "merchant_name": "Whole Foods"
    }
  ]
}
```

- `category_id` must reference one of the UUIDs sent in the request's `categories` list.
- `merchant_name` is a normalized, human-readable name derived from the raw `description`.
- If the service cannot classify a transaction, omit its entry from `predictions` — the backend will leave it `classified = false`.
