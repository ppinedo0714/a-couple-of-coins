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

## Categories

### `GET /categories`
Auth required. List all categories for the current user.

**Response `200`**
```json
[
  {
    "id": "uuid",
    "name": "Groceries",
    "color": "#4CAF50",
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

---

### `POST /categories`
Auth required. Create a category.

**Request**
```json
{
  "name": "Groceries",
  "color": "#4CAF50"
}
```

`color` is optional.

**Response `201`** — created category object

**Errors:** `400`, `409` name already exists for this user

---

### `PUT /categories/:id`
Auth required. Update a category.

**Request** — all fields optional
```json
{
  "name": "Food & Drink",
  "color": "#FF9800"
}
```

**Response `200`** — updated category object

---

### `DELETE /categories/:id`
Auth required. Delete a category. Transactions in this category will have `category_id` set to null.

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

**Response `200`** — updated transaction object

**Errors:** `400`, `404`

---

### `DELETE /transactions/:id`
Auth required. Delete a transaction. Updates the parent account balance.

**Response `204`** — no body

**Errors:** `404`

---

### `POST /transactions/classify`
Auth required. Sends all unclassified transactions (for the current user) to the prediction service. The prediction service returns a suggested category and normalized merchant name for each.

**Request** — no body required (classifies all unclassified transactions for the user)

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
