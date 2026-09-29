# Backend Docs Review — Consolidated Findings

Source: three independent adversarial reviews of `architecture.md`, `api.md`, `data-model.md`, `flows.md` (2026-08-15) — one from the perspective of an AI agent implementing the spec cold, one security-focused, one architecture-focused. Findings below are deduplicated and merged where more than one review hit the same issue. Tag shows which review(s) raised it: **[Impl]** implementability/ambiguity, **[Sec]** security, **[Arch]** architecture/operability.

This is a punch list for iterating on the docs (and, once fixed, the implementation) — not a verdict on code that exists yet, since none does.

---

## 1. Authentication & session design

- **JWT algorithm/key handling unspecified.** Docs never state the signing algorithm (presumably HS256) or that verification pins it — an alg-confusion risk if parsing ever trusts the token's own `alg` header. No `JWT_SECRET` rotation story either (rotating it currently mass-invalidates every access token, undocumented side effect). **[Sec]**
- **`Secure` cookie attribute inconsistent between documents.** `api.md`'s cookie table omits it; `flows.md` sequence diagrams show `Secure` on every `Set-Cookie`. Unclear whether it's unconditional (which would break local HTTP dev) or environment-conditional. State the rule explicitly. **[Impl]**
- **User enumeration via `409 email already registered`.** Combined with no CAPTCHA/bot mitigation on registration, this enables scripted probing of which emails exist ahead of credential stuffing. Consider a generic response or accept the tradeoff explicitly. **[Sec]**
- **Hard-delete of transactions, no audit trail.** `balance_checkpoints` are append-only by design, but transactions in an open period are truly deleted with no recovery path — including during the 15-minute access-token revocation-lag window after a session is hijacked/stolen. Consider soft-delete or a change log, at least for un-checkpointed rows. **[Sec]**

## 2. OAuth flow

- **Provider email trust: `email_verified` not checked.** `flows.md` §3 uses `profile.email` directly for account lookup/creation with no mention of checking the provider's verification flag. An attacker could claim an unverified email belonging to someone else and squat a new account under it. Require `email_verified = true` before using provider email for matching/creation, or fall back to provider-ID-only identity. **[Sec]**
- **`state` token not bound to the initiating provider.** `state` encodes `{nonce, intent, [userID]}` but not `provider`, and the `oauth_state` cookie path is shared across both callback routes — nothing stops a state minted for Google from being presented at the GitHub callback (code exchange would fail downstream, limiting real impact, but it's an avoidable gap). Encode and check `provider` inside the signed `state`. **[Sec]**
- **GitHub's private-email case isn't addressed.** GitHub can omit `email` from the profile response when the user has it set to private, requiring a separate `GET /user/emails` call and `user:email` scope. `flows.md` treats "fetch profile (email, provider ID)" as one uniform step for both providers — as written, private-email GitHub users will hit a null email. State the per-provider scopes and GitHub's fallback call. **[Impl]**

## 3. Input validation & injection

- **CSV upload has no size, row-count, or concurrency limits.** `flows.md` §5 implies in-memory parsing (`Parse CSV rows`) with no stated max file size, no streaming/bounded-memory guarantee, and no per-user cap on simultaneous import jobs — an authenticated user can exhaust memory/disk or queue unbounded concurrent imports. **[Sec]**
- **CSV column mapping is "flexible" with no actual mapping rules.** No header-matching rules (case sensitivity, accepted synonyms, header-row requirement, column-order independence) and no accepted CSV date format(s) are given, despite this being core, unguessable logic for `importer/csv.go`. **[Impl]**
- **`search` query param mechanism unspecified.** "Full-text search" could mean `ILIKE` substring matching or Postgres `tsvector`/`to_tsquery` — these differ materially in behavior. Name the exact mechanism, and separately confirm (as a hard rule, not an assumption) that the repository layer always parameterizes rather than string-building this or any query. **[Impl] [Sec]**
- **Numeric wire precision for `amount`/`balance` isn't pinned down.** Columns are `numeric(15,2)`; JSON has no decimal type. Unclear whether over-precise input (`42.567`) is rejected, rounded, or truncated. **[Impl]**
- **Password/email validation given only by example, not by rule.** `"password": "minimum8chars"` is a placeholder, not a stated rule (min/max length? complexity?); "email format" validation names no concrete check (regex? RFC parser?). **[Impl]**

## 4. Rate limiting & abuse protection

- **Only `register`/`login`/`refresh` are rate-limited.** `/transactions/classify`, `/import/csv`, and checkpoint creation have no request-volume limits — the advisory lock stops *concurrent* classify runs per user but not serial hammering, so an authenticated (or stolen-session) client can drive sustained load against Postgres and the Python prediction service. Extend `httprate` or a per-user token bucket to these routes. **[Sec]**
- **No bot mitigation (CAPTCHA or equivalent) on registration** — per-IP/email limits alone are straightforward to distribute around. **[Sec]**

## 5. Logging & data exposure

- **Ad hoc, unstructured `log.Printf` logging is flagged in `architecture.md` as a known gap, but its security implication isn't called out**: nothing currently prevents a call site from logging full request bodies (passwords, refresh tokens) or PII/financial data (transaction descriptions, emails) in the clear. When a structured logger is adopted, explicitly ban logging request bodies and cookie/token values as a stated rule, not just an implementation detail left to whoever writes each call site. **[Sec] [Arch — same gap, operability angle below]**

## 6. Cross-service trust (prediction service)

- **The Go → Python call has no documented authentication.** `PREDICTIONS_SERVICE_URL` is just a base URL — no API key, mTLS, or explicit network-isolation statement. If the two services aren't provably network-isolated, anything reaching port 8001 can inject arbitrary `category_id`/`merchant_name` values — though the `category_id` half is now caught by the `categories(user_id, id)` composite foreign key regardless of what the service sends. State the trust boundary explicitly or add a shared-secret header. **[Sec]**
- **No resilience or versioning story beyond the one per-batch timeout** — no retry/backoff, no circuit breaker, no schema versioning for the `/classify` contract. A response-shape change on the Python side fails silently into `failed` counts with nothing to catch it before production. **[Arch]**

## 7. Layering & architecture consistency

- **The startup recovery sweep breaks the stated 4-layer rule.** Both `CLAUDE.md` and `architecture.md` state a strict `handlers → services → repository → db` rule with no skipping, but `flows.md` §5a shows `main.go` calling the `ImportJob Repository` directly with no service in between. Either extend the rule to explicitly permit bootstrap code in `main.go` to call repositories, or route the sweep through a thin service method so the rule has no silent exception. **[Arch]**

## 8. Scalability & data model

- **Only one index is documented** (`(account_id, dedup_hash)`, for import dedup), but the hottest read paths — current-balance lookups, the per-anchor re-scan in `GET /accounts/history`, the closed-period check on every transaction edit/delete, and `GET /transactions` filtering — all need `(account_id, date)` and likely `(user_id, date)` composite indexes to stay fast as `transactions` grows. No partitioning/archival/retention strategy is mentioned for a table that's append-heavy and never pruned by design. **[Arch]**
- **No default sort order (or pagination tie-break) is specified for any list endpoint** (`GET /transactions`, `GET /accounts`, `GET /groups`, `GET /groups/categories`). Without a stable `ORDER BY`, offset-based pagination on `GET /transactions` is undefined — rows can shift, duplicate, or be skipped across pages. Pin an explicit order (e.g. `date DESC, created_at DESC, id`) per endpoint. **[Impl]**

## 9. Observability & operability

- **The "no structured logging" gap is understated in scope.** Beyond just being unstructured, there's no correlation ID linking the batches of one classify run or one import job, `GET /health` returns a hardcoded `{status:"ok"}` with no DB-connectivity check (a container can pass health checks while the pool is down), and there's no stated behavior for migration failure at boot (crash-loop vs. serve-anyway degraded). For the reliability-sensitive paths these docs already call out (batched imports, classify retries, advisory locks), operators currently have no way to alert on a stuck `processing` job before the 10-minute sweep fires, or to see which batches failed and why. **[Arch]**

## 10. Testing & CI strategy

- **No test-DB provisioning, fixtures, or CI pipeline description**, despite `architecture.md` stating repository tests need "a real (not mocked) Postgres instance" and that no integration/e2e runner exists yet. For money-handling logic (checkpoints, dedup, classify batching) this is a real gap for a project at this stage, not a nice-to-have to defer. **[Arch]**

## 11. API specification gaps (implementability)

Smaller, individually low-risk but collectively significant ambiguities that would cause two independent implementations of this spec to diverge:

- `POST /import/csv` has no Errors section at all, despite `flows.md` §5 showing an explicit "validate account belongs to user" step (needs `400`/`404`).
- `currency` is described in `data-model.md` as "set once at registration," but no registration or OAuth-signup path in `api.md`/`flows.md` actually accepts a `currency` field — every user is permanently USD-only in practice. Either state that explicitly as a v1 constraint or add the field.
- `pg_try_advisory_lock(hash(userID))` names no concrete hash function for turning a UUID into the required bigint lock key — collision behavior is currently unspecified and non-reproducible between implementations.
- Invalid/unowned IDs inside `GET /accounts/history`'s comma-separated `account_ids` — silently dropped vs. `400`/`404` — isn't stated, inconsistent with the explicit ownership-`404` pattern used elsewhere.
- No general statement on validation-error precedence when multiple request fields are simultaneously invalid.

---

## Suggested next step

All three candidates from the original punch list are now resolved (below). What's left is the rest of §4–§11 above — rate limiting on `classify`/`import`/checkpoints (§4), the ad hoc-logging gap (§5), prediction-service trust/resilience (§6), the startup-sweep layering exception (§7), missing indexes and list-endpoint sort order (§8), observability (§9), test/CI strategy (§10), and the smaller API spec gaps (§11) — no single standout among them; pick based on what's next in priority.

~~CSV upload size/row/concurrency limits~~ — done: three limits added, none of them app-level pre-checks alone. `CSV_IMPORT_MAX_FILE_SIZE_BYTES` (default 5 MiB) is enforced synchronously in the handler via `http.MaxBytesReader` before any parsing (`413`). `CSV_IMPORT_MAX_ROWS` (default 50,000) is enforced against a running count as rows are streamed from `csv.Reader` — parsing itself is now documented as streamed (bounded to one batch/100 rows in memory) rather than buffering the whole file, closing the memory-guarantee gap independent of the row cap; exceeding it ends the job `status=failed` with a `row_limit_exceeded` `row_errors` entry, preserving rows already committed. Per-user concurrency is capped at one active job via a new partial unique index on `import_jobs(user_id) WHERE status IN ('pending','processing')` — same DB-constraint-not-app-check pattern as `categories(user_id, id)` on `transactions.category_id` — caught as a `409`. Also filled in `POST /import/csv`'s previously-empty Errors section (`400`/`404`/`409`/`413`), closing that bullet in §11. See `flows.md` §5, `api.md` (`POST /import/csv`, Error Responses), `architecture.md` (new CSV Import Limits section, two new env vars), `data-model.md` (`import_jobs`).

~~OAuth `email_verified` not checked~~ — done: the "no existing OAuth connection" branch of `flows.md` §3 now checks `profile.email_verified` *before* `GetUserByEmail` runs at all — an unverified email fails closed with a new `ErrOAuthEmailUnverified` / `302 .../login?error=oauth_email_unverified`, never reaching the matching or account-creation logic. The "existing OAuth connection found" branch (matched by `provider`+`provider_user_id`) is unaffected, since that identity is already proven regardless of the current verification flag. In practice this only bites GitHub — Google's userinfo endpoint always reports the returned email as verified. See `flows.md` §3, `api.md` (`GET /auth/oauth/:provider/callback`), `open-issues.md` §1.1 (same account-squatting-by-email threat model as the no-auto-merge rationale).

~~JWT algorithm pinning + `JWT_SECRET` rotation story~~ — done: verification now explicitly pins `HS256` (`jwt.WithValidMethods`) rather than trusting the token's `alg` header, and the rotation story is documented — rotating `JWT_SECRET` invalidates outstanding access-token signatures but not sessions (refresh tokens are independent, hashed, DB-stored values), so the existing 401→`POST /auth/refresh` fallback self-heals every client within one round-trip. No dual-key verification window; accepted as a v1 tradeoff given the self-healing path. See `api.md` (Authentication), `flows.md` (§4), `architecture.md` (new JWT Signing section, `JWT_SECRET` env var row).

~~Password reset/change flow~~ — done: `POST /auth/password/set`, `/change`, `/reset/request`, `/reset/confirm` added, including a new (minimal) email-delivery seam (`EMAIL_PROVIDER_URL`/`_API_KEY`/`_FROM_ADDRESS`, `internal/services/email/client.go`) and a stateless, self-invalidating reset-token design keyed to `password_hash`. Also added the previously-missing `FRONTEND_URL` env var, needed for both reset links and the pre-existing OAuth redirects (closing that bullet in §11, API specification gaps). See `architecture.md` (directory layout, env vars, rate limiting), `api.md` (`GET /users/me`, new `POST /auth/password/*` endpoints), `data-model.md` (`users.password_hash`, new Key Design Decision), `flows.md` (§2c/§2d/§2e), `open-issues.md` (§1.1 updated to resolved).

~~`category_id` ownership validation~~ — done: enforced via a `categories(user_id, id)` composite foreign key instead of an application-level check. See `data-model.md` (`categories`, `transactions`, Key Design Decisions), `api.md` (`POST`/`PUT /transactions`, Prediction Service Internal Contract), `flows.md` (§6, §7).

~~Advisory-lock/pool-connection semantics~~ — done: the classify run now pins a single dedicated connection (`pool.Acquire()`) for the lock, both repository fetches, every batch's writes, and the unlock, with the batch-atomicity tradeoff clarified rather than restructured. See `flows.md` §6, `api.md` (`POST /transactions/classify` steps 1 and 5).

~~`open-issues.md`~~ — done: created with §1.1 (OAuth account-merge rationale) and §2.1–2.4 (CSV import/job-processing gaps: no graceful shutdown, sweep scope, multi-instance race, coarse partial-failure reporting — folding in the former §9 findings). See `open-issues.md`, cross-referenced from `flows.md` §3, §5, §5a.
