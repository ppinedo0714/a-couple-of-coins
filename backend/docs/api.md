# API Reference

Base URL: `http://localhost:8000/api/v1`

## Authentication

The API uses two **httpOnly cookies**, both set by the backend on successful login, registration, or refresh:

| Cookie | Contents | TTL | `Path` | `SameSite` |
|--------|----------|-----|--------|------------|
| `token` | Signed JWT (access token); claims: `userID`, `sid` (session ID) | 15 minutes | `/` | `Lax` |
| `refresh_token` | Opaque random value | 30 days | `/api/v1/auth` | `Strict` |

A third, narrower-purpose cookie, `oauth_state`, is set only for the duration of an OAuth redirect round-trip — see `GET /auth/oauth/google` below.

- Every authenticated endpoint (`Auth required` below) checks only the `token` cookie — a valid, unexpired JWT signature is sufficient; there's no database lookup per request. This makes auth checks cheap but means revoking access is not instantaneous: see the bounded revocation window below.
- Access tokens are signed with HMAC-SHA256 (`HS256`) using `JWT_SECRET`. Verification pins the algorithm explicitly (`golang-jwt/jwt/v5`'s `jwt.WithValidMethods([]string{"HS256"})`) rather than trusting the token's own `alg` header — this closes the classic alg-confusion attack class (e.g. an attacker crafting an `alg: none` or RS256-with-public-key-as-HMAC-secret token). See `flows.md` §4 and `architecture.md` JWT Signing for the `JWT_SECRET` rotation story.
- `refresh_token` is scoped to `Path=/api/v1/auth` — the browser only ever sends it to auth endpoints (`/auth/refresh`, `/auth/logout`, `/auth/logout/all`, `/auth/sessions`), not to every request. It is validated against the `sessions` table (see `data-model.md`), which is what makes it actually revocable.
- Neither token is returned in a JSON response body.
- CORS is configured with `Access-Control-Allow-Credentials: true` and an exact `Access-Control-Allow-Origin` (no wildcard).
- When the `token` cookie is expired or missing, the frontend calls `POST /auth/refresh` to get a new one using the still-valid `refresh_token` — see below. If that also fails (expired or revoked), the user must log in again.

**Revocation is not instant for the access token.** Logging out, revoking a specific session, or "log out everywhere" all take effect on the *refresh token* immediately — no future `POST /auth/refresh` call will succeed for a revoked session. But an already-issued `token` JWT is not individually checked against the database, so it remains valid for up to its remaining 15-minute TTL even after the session it came from is revoked. This bound (15 minutes, worst case) was a deliberate tradeoff against adding a DB check to every request — see `data-model.md` Key Design Decisions.

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
| `409` | Conflict — e.g. duplicate email, account has transactions, duplicate group or category name, OAuth identity already linked to another user |
| `429` | Rate limited — `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/password/reset/request`, `POST /auth/password/reset/confirm`, `POST /auth/password/change`, `POST /auth/password/set` only. Response includes a `Retry-After` header (seconds). See `architecture.md` Rate Limiting |
| `413` | Payload too large — `POST /import/csv` only, file exceeds `CSV_IMPORT_MAX_FILE_SIZE_BYTES`. See `architecture.md` CSV Import Limits |
| `500` | Internal server error — logged server-side; generic message returned to client |

**Exception:** the OAuth callback endpoints (`GET /auth/oauth/:provider/callback`) never return a JSON error body — by the time an error is known, the browser is mid-redirect from the OAuth provider, not making a fetch the frontend can inspect. Errors there are instead surfaced as a `302` redirect to the frontend with an `error` query param (e.g. `/login?error=oauth_email_conflict`); see the endpoint docs below.

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

**Response `201`** — sets the `token` and `refresh_token` cookies (see Authentication above)
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

Email is normalized to lowercase before the uniqueness check and storage — `Jane@Example.com` and `jane@example.com` are the same account (see `data-model.md` `users.email`). The response's `email` field reflects the normalized (lowercase) value, not necessarily what was submitted.

**Errors:** `400` invalid input, `409` email already registered, `429` rate limited (per-IP and per-email, see Error Responses above)

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

**Response `200`** — sets the `token` and `refresh_token` cookies (see Authentication above)
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

Email is normalized to lowercase before lookup, so login is case-insensitive regardless of how the address was cased at registration.

**Errors:** `400` invalid input, `401` wrong credentials, `429` rate limited (per-IP and per-email, see Error Responses above)

---

### `POST /auth/refresh`
Exchanges a valid `refresh_token` cookie for a new `token` (access token). Called by the frontend when the access token has expired (or is about to), so the user isn't forced to log in again every 15 minutes.

**Request** — no body; reads the `refresh_token` cookie

**Response `200`** — sets new `token` and `refresh_token` cookies (rotation — the old refresh token is invalidated, see `flows.md` §2a)

**Errors:** `401` missing, invalid, expired, or already-used `refresh_token`. A reused (already-rotated) refresh token additionally revokes every session for that user server-side, as a compromise response — the client only sees `401`. `429` rate limited (per-IP only — no email in the request, see Error Responses above)

---

### `GET /auth/oauth/google`
Redirects the browser to Google's OAuth consent screen, to log in or register via Google.

Generates a signed `state` token (nonce + `intent=login`, ~10 min expiry), passes it as the `state` query param to Google, and sets it as the `oauth_state` cookie (httpOnly, ~10 min, `Path=/api/v1/auth/oauth`). See `flows.md` §3.

**Response:** `302` redirect, sets `oauth_state` cookie

---

### `GET /auth/oauth/google/callback`
Google redirects here after the user grants permission. Behavior depends on the `state` token's `intent`, set when the flow was initiated — see `flows.md` §3 and §3a.

**Query params:** `code`, `state`

Before anything else, this endpoint validates `state`'s signature and expiry, then compares it byte-for-byte against the `oauth_state` cookie. This double-check is the actual CSRF protection: the signature alone proves `state` is genuine, but only the cookie proves the callback is happening in the same browser that started the flow — `state` itself is visible in the redirect URL and could otherwise be replayed by an attacker from a different browser. On success or failure, `oauth_state` is cleared.

- Missing/invalid/expired `state`, or a mismatch against `oauth_state`: `302` redirect to `<frontend-url>/login?error=oauth_state_mismatch` (or `/settings?...` if `intent=link`). No cookie is set, no user is created or modified.

**Response `302` (login/signup, `intent=login`)**
- Success — an existing `oauth_connections` match, or no user with this email exists yet (a new user is created): redirects to `<frontend-url>/login?oauth=success`, setting the `token` and `refresh_token` cookies (see Authentication above).
- Conflict — no `oauth_connections` match, but the email already belongs to an existing user (password-only, or linked to a different provider): redirects to `<frontend-url>/login?error=oauth_email_conflict`. No cookie is set, no user is created or modified. The user must log in with their existing method and link this provider from settings (`GET /auth/oauth/google/link` below).
- Unverified email — no `oauth_connections` match, and the provider reports the profile email as unverified (`email_verified: false`, or GitHub's per-address `verified` flag `false`): redirects to `<frontend-url>/login?error=oauth_email_unverified`. No cookie is set, no user is created or modified — the email is never even checked against existing accounts, since an unverified provider email isn't trustworthy enough to match or create with. The user should verify the address with the provider, or use a different sign-in method. See `flows.md` §3.

**Response `302` (linking, `intent=link`, requires the initiating request to have been authenticated)**
- Success: creates an `oauth_connections` row for the already-logged-in user, redirects to `<frontend-url>/settings?oauth=linked`.
- Conflict — this provider identity is already linked to a user (this one or another): redirects to `<frontend-url>/settings?error=oauth_already_linked`.

---

### `GET /auth/oauth/google/link`
Auth required. Redirects to Google's OAuth consent screen to link a Google identity to the current (already-authenticated) account. Uses the same callback as `GET /auth/oauth/google/callback` above, distinguished by `state`'s `intent=link`. See `flows.md` §3a.

Generates and sets `oauth_state` the same way as `GET /auth/oauth/google` above, except the signed `state` token additionally encodes the caller's `userID` (`intent=link`) so the callback knows which account to link to.

**Response:** `302` redirect, sets `oauth_state` cookie

---

### `GET /auth/oauth/github`
Redirects to GitHub's OAuth consent screen, to log in or register via GitHub.

Generates and sets `oauth_state` the same way as `GET /auth/oauth/google` above.

**Response:** `302` redirect, sets `oauth_state` cookie

---

### `GET /auth/oauth/github/callback`
GitHub redirects here after authorization. Same behavior as the Google callback above (login/signup vs. link, conflict handling, `state`/`oauth_state` validation, and redirect targets), for the `github` provider.

---

### `GET /auth/oauth/github/link`
Auth required. Same as `GET /auth/oauth/google/link`, for the `github` provider.

**Response:** `302` redirect

---

### `POST /auth/logout`
Auth required. Revokes the session tied to the `refresh_token` cookie (so it can no longer be used at `POST /auth/refresh`) and clears both cookies. This is a real, server-side revocation of the refresh token — not just clearing the browser's cookie — though the caller's own already-issued access token remains valid for up to its remaining 15-minute TTL, per the Authentication section above.

**Response `204`** — no body, clears `token` and `refresh_token` cookies

---

### `POST /auth/logout/all`
Auth required. Revokes every non-revoked session belonging to the current user — "log out everywhere," e.g. after a stolen or lost device. Clears the caller's own cookies too, since their session is included.

**Response `204`** — no body, clears `token` and `refresh_token` cookies

---

### `GET /auth/sessions`
Auth required. Lists the current user's active sessions (`revoked_at IS NULL AND expires_at > now()`), most recently used first — powers a "your devices" settings page.

**Response `200`**
```json
[
  {
    "id": "uuid",
    "user_agent": "Mozilla/5.0 (Macintosh...)",
    "created_at": "2024-01-15T10:30:00Z",
    "last_used_at": "2024-01-20T08:00:00Z",
    "is_current": true
  }
]
```

`is_current` marks the session tied to the request's own access token (`sid` claim — see `flows.md` §4), so the UI can label it "this device."

---

### `DELETE /auth/sessions/:id`
Auth required. Revokes a single session — "log out this device" for a session other than (or including) the one making the request. Only affects sessions owned by the current user.

**Response `204`** — no body. If the revoked session was the caller's own current session, its `refresh_token` cookie is now invalid but the response does not clear cookies (the caller is expected to be acting on a different session's ID, e.g. from the `GET /auth/sessions` list).

**Errors:** `404` session not found or not owned by user

---

### `POST /auth/password/set`
Auth required. Sets a password for an account that doesn't have one yet — the OAuth-only-signup path to enabling password login alongside (or instead of) OAuth. Use `POST /auth/password/change` instead if the account already has a password (`GET /users/me`'s `has_password` field tells the frontend which to call).

**Request**
```json
{
  "new_password": "minimum8chars"
}
```

**Response `204`** — no body. Does not affect any existing session — the caller stays logged in.

**Errors:** `400` invalid input, `409` account already has a password, `429` rate limited (per-IP only, see Error Responses above)

---

### `POST /auth/password/change`
Auth required. Changes the password on an account that already has one.

**Request**
```json
{
  "current_password": "oldpassword123",
  "new_password": "newpassword456"
}
```

**Response `204`** — no body. On success, revokes every session for this user *except* the one making the request (same effect as `POST /auth/logout/all`, minus the caller's own session) — every other device is signed out, but the caller isn't.

**Errors:** `400` invalid input, `401` `current_password` incorrect (or not authenticated, same as any other endpoint), `409` account has no password yet — use `POST /auth/password/set`, `429` rate limited (per-IP only, see Error Responses above)

---

### `POST /auth/password/reset/request`
No auth required. Starts the "forgot password" flow: if the given email belongs to an account with a password, emails a time-limited reset link (`<frontend-url>/reset-password?token=...`) to it via `internal/services/email/client.go`.

**Request**
```json
{
  "email": "user@example.com"
}
```

**Response `200`** — always the same response, regardless of whether the email exists, belongs to an OAuth-only account with no password, or a reset was actually sent, so the endpoint can't be used to enumerate registered emails:
```json
{ "message": "If an account with a password exists for that email, a reset link has been sent." }
```

The reset token is a signed, self-contained value (not stored server-side) — see `data-model.md` Key Design Decisions ("Password-reset tokens are stateless and self-invalidating") and `flows.md` §2d.

**Errors:** `400` invalid input, `429` rate limited (per-IP and per-email, see Error Responses above)

---

### `POST /auth/password/reset/confirm`
No auth required. Completes the "forgot password" flow using the token emailed by the request step above.

**Request**
```json
{
  "token": "<opaque signed token from the email link>",
  "new_password": "newpassword456"
}
```

**Response `204`** — no body. Revokes **every** session for the user (not "every session but this one" — there is no authenticated caller session here, since this endpoint isn't authenticated, and the flow is a recovery path that should assume the account may have been compromised). The user must log in again afterward; this endpoint does not itself set cookies or log the user in.

**Errors:** `400` invalid input, invalid/expired token, or token's password fingerprint stale (see `data-model.md`) — all three are reported identically to avoid distinguishing "expired" from "already used" from "malformed" for an attacker; `429` rate limited (per-IP only, see Error Responses above)

---

## Users

### `GET /users/me`
Auth required. Returns the current user's profile.

**Response `200`**
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "currency": "USD",
  "created_at": "2024-01-15T10:30:00Z",
  "has_password": true,
  "oauth_providers": ["google"]
}
```

`oauth_providers` lists which providers (`google`, `github`) the account has linked, from `oauth_connections` — drives whether the settings page offers "Link Google" or shows it as already connected. `has_password` is `false` for accounts created via OAuth that have never set a password — it drives both whether to show a "password login" option at all, and which endpoint a settings page's password form should call: `POST /auth/password/set` when `false`, `POST /auth/password/change` when `true`. `currency` (ISO 4217, e.g. `USD`) is set once at registration (defaults to `USD`) and applies to all of the user's accounts — see `data-model.md` `users.currency`.

---

### `PUT /users/me`
Auth required. Update the current user's profile.

**Request** — all fields optional
```json
{
  "email": "newemail@example.com"
}
```

A new `email` is normalized to lowercase before the uniqueness check and storage, same as registration. `currency` cannot be changed via this endpoint (or any endpoint) — see `data-model.md` Key Design Decisions.

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
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

`balance` is computed from the account's latest checkpoint plus transactions since — not a stored column (see `data-model.md`). Currency is not included per-account — it's a single, user-level setting (`GET /users/me`); every account returned here is denominated in it.

---

### `POST /accounts`
Auth required. Create a new account. `balance` is required — it becomes the account's first `balance_checkpoints` row, dated today. Both inserts run in one DB transaction — see `flows.md` §7a.

**Request**
```json
{
  "name": "Chase Checking",
  "type": "checking",
  "balance": 2450.00
}
```

`type` must be one of: `checking`, `savings`, `credit`, `investment`

**Response `201`** — created account object (see `GET /accounts` for shape; `balance` reflects the checkpoint just created)

**Errors:** `400` invalid input (including missing `balance`)

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
Auth required. Return computed balance-over-time for one or more accounts over a date range. There is no stored snapshot table — for each returned point, the backend finds the applicable checkpoint (the latest checkpoint at or before that date) and adds transactions since. See "Balance History" in `flows.md`.

**Query params**

| Param | Type | Description |
|-------|------|-------------|
| `account_ids` | string | Comma-separated UUIDs (optional; default: all user accounts) |
| `from` | date | Start date inclusive, `YYYY-MM-DD` (required) |
| `to` | date | End date inclusive, `YYYY-MM-DD` (required) |
| `interval` | string | `day` \| `week` \| `month` — granularity of returned points (optional; default: `day`) |

For `week` and `month` intervals, the balance as of the last day in each period is returned. Points are ordered by date ascending. Only accounts belonging to the authenticated user are returned. A date before an account's first checkpoint is not returned for that account — there is no balance to report before the account existed.

**Response `200`**
```json
{
  "history": [
    { "date": "2025-01-06", "account_id": "uuid", "balance": 2100.00 },
    { "date": "2025-01-13", "account_id": "uuid", "balance": 2350.00 }
  ]
}
```

No `currency` per entry — every account belonging to a user shares the same currency (`users.currency`, see `data-model.md`), so entries here never need disambiguating and are always safe to compare or sum across accounts.

**Errors:** `400` missing or invalid `from`/`to`

---

### `POST /accounts/:id/checkpoints`
Auth required. Record a new balance check-in for an account — "as of today (or a past date), my balance is actually X."

**Request**
```json
{
  "date": "2025-01-15",
  "balance": 75.00
}
```

`date` must be on or after the account's current latest checkpoint date, and not in the future. Same-day corrections are allowed — submitting another checkpoint for a date that already has one is valid; it becomes the new latest checkpoint for that date.

**Response `201`**
```json
{
  "id": "uuid",
  "account_id": "uuid",
  "date": "2025-01-15",
  "balance": 75.00,
  "expected_balance": 80.00,
  "discrepancy": -5.00,
  "created_at": "2025-01-15T18:04:00Z"
}
```

`expected_balance` is what the backend calculated before this checkpoint was applied (previous checkpoint + transactions since); `discrepancy` is `balance - expected_balance`. Both are `null` on an account's very first checkpoint — there is nothing prior to compare against.

**Errors:** `400` `date` before the account's latest checkpoint, or `date` in the future; `404` account not found or not owned by user

---

### `GET /accounts/:id/checkpoints`
Auth required. List an account's checkpoint history, most recent first.

**Response `200`**
```json
[
  {
    "id": "uuid",
    "account_id": "uuid",
    "date": "2025-01-15",
    "balance": 75.00,
    "expected_balance": 80.00,
    "discrepancy": -5.00,
    "created_at": "2025-01-15T18:04:00Z"
  }
]
```

**Errors:** `404` account not found or not owned by user

---

## Groups

A **group** is a broad label (e.g. "Food", "Entertainment"). A **category** is a specific item under exactly one group (e.g. "Groceries", "Movies"). Groups and categories are separate resources — see `data-model.md` for why. A transaction's `category_id` always references a category, never a group.

### `GET /groups`
Auth required. List all groups for the current user, each with its categories embedded.

**Response `200`**
```json
[
  {
    "id": "uuid",
    "name": "Entertainment",
    "color": "#EC407A",
    "created_at": "2024-01-15T10:30:00Z",
    "categories": [
      { "id": "uuid", "group_id": "uuid", "name": "Movies", "created_at": "2024-01-15T10:30:00Z" }
    ]
  }
]
```

---

### `POST /groups`
Auth required. Create a group.

**Request**
```json
{
  "name": "Entertainment",
  "color": "#EC407A"
}
```

`color` is required and must be `#RRGGBB`.

**Response `201`** — created group object (`categories: []`)

**Errors:** `400` invalid input (including missing or malformed `color`), `409` name already exists

---

### `GET /groups/:id`
Auth required. Get a single group, with its categories embedded (same shape as a `GET /groups` list item).

**Response `200`** — group object

**Errors:** `404` not found or not owned by user

---

### `PUT /groups/:id`
Auth required. Update a group.

**Request** — all fields optional
```json
{
  "name": "Streaming & Fun",
  "color": "#FF9800"
}
```

**Response `200`** — updated group object

**Errors:** `400` invalid input, `404`, `409` name already exists

---

### `DELETE /groups/:id`
Auth required. Delete a group. Blocked if it still has categories — the only way to empty a group today is deleting each category individually via `DELETE /groups/categories/:id` (which nullifies `category_id` on that category's transactions); there is no bulk-move.

**Response `204`** — no body

**Errors:** `404`, `409` group still has categories

---

### `GET /groups/categories`
Auth required. List all categories for the current user as a flat list. Each entry includes `color`, inherited from its group, so the frontend doesn't need a separate lookup against `/groups`.

**Response `200`**
```json
[
  {
    "id": "uuid",
    "group_id": "uuid",
    "name": "Movies",
    "color": "#EC407A",
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

---

### `POST /groups/categories`
Auth required. Create a category under a group.

**Request**
```json
{
  "name": "Movies",
  "group_id": "<group-uuid>"
}
```

`group_id` is required and must reference a group owned by the current user.

**Response `201`** — created category object (`color` inherited from the group, as in `GET /groups/categories`)

**Errors:** `400` invalid input or missing `group_id`, `404` group not found, `409` name already exists within that group

---

### `GET /groups/categories/:id`
Auth required. Get a single category.

**Response `200`** — category object

**Errors:** `404` not found or not owned by user

---

### `PUT /groups/categories/:id`
Auth required. Update a category, or move it to a different group.

**Request** — all fields optional
```json
{
  "name": "Streaming",
  "group_id": "<other-group-uuid>"
}
```

`group_id` moves the category to a different group owned by the same user. Moving a category does not affect `category_id` on any transaction referencing it, since the category's `id` is unchanged; it does mean historical reports that group transactions by category will reflect the new grouping retroactively.

**Response `200`** — updated category object

**Errors:** `400` invalid input or `group_id` references a nonexistent group, `404`, `409` name already exists within the target group

---

### `DELETE /groups/categories/:id`
Auth required. Delete a category. Transactions referencing it have `category_id` set to null.

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
| `duplicate` | bool | If `true`, return only transactions flagged `is_duplicate` — powers a "review possible duplicates" view |
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
      "is_duplicate": false,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 143,
  "limit": 50,
  "offset": 0
}
```

`is_duplicate` is `true` when this row was flagged as a likely re-import at CSV import time (see `data-model.md` `transactions.is_duplicate`); always `false` for `manual` and `bank` sources. It's a flag for the user to review, not an automatic rejection — the row exists and counts toward balance like any other transaction regardless of its value.

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

`category_id` is optional. If provided, it must belong to a category owned by the current user — enforced at the database level via a composite foreign key (see `data-model.md` `transactions.category_id`), not a separate pre-check; a `category_id` belonging to another user is rejected the same way a nonexistent one is. `amount` is signed (negative = expense).

The backend sets `source = "manual"` and `classified = true` on creation — the user has already described the transaction, so it does not need prediction-service processing. `merchant_name` is left null for manually-created transactions. These fields are not accepted from the request body.

The account's current balance reflects this transaction immediately — balance is computed at read time, not stored, so there is no separate balance-update step.

**Response `201`** — created transaction object

**Errors:** `400`, `404` account not found, or `category_id` not found or not owned by user

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

Pass `"category_id": null` explicitly to remove the category assignment. A non-null `category_id` must belong to the current user, enforced the same way as on creation (see `POST /transactions`) — a `category_id` belonging to another user, or a nonexistent one, is rejected as `404`. `account_id` is not an accepted field on this endpoint — a transaction's account cannot be changed after creation.

`category_id` and `description` may always be edited. `amount` and `date` may only be changed if the transaction's **current** `date` is after the account's latest checkpoint — changing either field on a transaction in a closed period is rejected, since it would silently fail to affect the account's already-reconciled balance.

`is_duplicate` may also be included, and only accepts `false` — this is how a user dismisses a flagged duplicate they've reviewed and determined is actually legitimate (e.g. a real repeated charge), without deleting the row. There's no way to set it back to `true` via this endpoint; it's only ever set `true` by the CSV import path.

**Response `200`** — updated transaction object

**Errors:** `400`, `404` (including a `category_id` not found or not owned by user), `409` transaction is in a closed period (`date` at or before the account's latest checkpoint) and `amount` or `date` was included in the request

---

### `DELETE /transactions/:id`
Auth required. Delete a transaction. Only allowed if the transaction's `date` is after the account's latest checkpoint — deleting a transaction from a closed period would silently fail to affect the account's already-reconciled balance.

**Response `204`** — no body

**Errors:** `404`, `409` transaction is in a closed period

---

### `POST /transactions/classify`
Auth required. Sends all unclassified transactions (for the current user) to the prediction service, in batches, and writes back what it can. The prediction service returns a `category_id` from the user's existing categories and a normalized `merchant_name` for each transaction.

**Request** — no body required (classifies all unclassified transactions for the user)

The backend:
1. Checks out a dedicated connection from the pool and acquires a Postgres advisory lock scoped to the user's ID (`pg_try_advisory_lock`) on it — advisory locks are connection-scoped, so this run holds that one connection for its full duration rather than borrowing a fresh one per query (see `flows.md` §6 for why). If another classify run for this user already holds the lock (a double-click, two open tabs), the request fails immediately with `409` instead of duplicating the work — see Errors below.
2. Fetches all transactions where `classified = false` for the user, and all categories owned by the user.
3. Splits the transactions into batches of 100 and, for each batch, POSTs it (with the full category list) to `services/predictions` (see [Prediction Service Internal Contract](#prediction-service-internal-contract)) under a `PREDICTIONS_TIMEOUT_SECONDS` timeout (default 15s — see `architecture.md` Environment Variables).
4. For each batch that succeeds, writes back the returned `category_id` and `merchant_name` and sets `classified = true` for each matched transaction, committing before moving to the next batch. A batch that times out or errors is skipped — its transactions stay `classified = false` and are counted in `failed` — and processing continues with the remaining batches, so one bad batch doesn't discard progress already committed from earlier ones.
5. Releases the advisory lock and returns the pinned connection to the pool (always — including on error, timeout, or panic).

**Response `200`**
```json
{
  "classified": 312,
  "failed": 8
}
```

`failed` counts both transactions the prediction service considered but couldn't match to a category, and transactions in a batch that errored or timed out before the service ever responded — the response doesn't distinguish between the two. Either way, calling the endpoint again will retry them, since they remain `classified = false`.

**Errors:** `409` a classification run is already in progress for this user (see step 1 above)

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

The upload is subject to three limits, all enforced before or during processing rather than left unbounded — see `architecture.md` CSV Import Limits and `flows.md` §5 for the full rationale:
- The file itself may not exceed `CSV_IMPORT_MAX_FILE_SIZE_BYTES` (default 5 MiB), checked before parsing starts.
- The file may not contain more than `CSV_IMPORT_MAX_ROWS` (default 50,000) rows; a file that does has its import stopped mid-stream, ending as a `failed` job (see `GET /import/jobs/:id`) rather than silently truncating.
- Only one import job (`pending` or `processing`) may be active per user at a time.

**Response `202`**
```json
{
  "job_id": "uuid",
  "status": "pending"
}
```

**Errors:** `400` invalid input (missing `file` or `account_id`), `404` account not found or not owned by user, `409` an import job is already in progress for this user, `413` file exceeds `CSV_IMPORT_MAX_FILE_SIZE_BYTES`

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
    "rows_duplicate": 3,
    "rows_failed": 0,
    "created_at": "2024-01-15T10:30:00Z",
    "completed_at": "2024-01-15T10:30:05Z"
  }
]
```

`rows_duplicate` counts rows within `rows_imported` that were flagged as likely re-imports of an existing transaction (see `data-model.md` `transactions.is_duplicate`) — it's a subset, not additional to, `rows_imported`. `rows_failed` counts rows that couldn't be parsed and were not inserted at all; `rows_total = rows_imported + rows_failed` always holds.

---

### `GET /import/jobs/:id`
Auth required. Get the status of a specific import job.

**Response `200`** — same shape as list item above, plus `row_errors` (omitted from the list endpoint to keep it lean):
```json
{
  "id": "uuid",
  "status": "done",
  "source_type": "csv",
  "file_name": "transactions_jan.csv",
  "rows_total": 47,
  "rows_imported": 45,
  "rows_duplicate": 3,
  "rows_failed": 2,
  "row_errors": [
    { "row": 12, "reason": "invalid_date" },
    { "row": 30, "reason": "invalid_amount" }
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:30:05Z"
}
```

`row_errors` is empty/omitted when `rows_failed` is `0`. `row` is the 1-indexed row number in the uploaded file (header row excluded), so it can be surfaced to the user as "row 12 in your file failed: invalid date."

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
      "group_name": "Food"
    }
  ]
}
```

- `categories` contains all of the user's categories — never groups, since a transaction's `category_id` can only ever reference a category. `group_name` gives the model context about which broad label each category falls under.
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
- The backend does not trust this contractually — writing a prediction back relies on the same `categories(user_id, id)` composite foreign key as any other write (see `data-model.md`). A `category_id` that isn't actually one of the user's categories (a misbehaving service, or a category deleted mid-run) fails that constraint, and the transaction is counted in `failed` rather than written — the same outcome as a batch timeout. See `flows.md` §6.
