# Key Flows

Sequence diagrams for the most important request paths through the backend.

---

## 1. Email/Password Registration

```mermaid
sequenceDiagram
    participant C as Client
    participant H as Handler (auth.go)
    participant S as Auth Service
    participant R as User Repository
    participant DB as PostgreSQL

    C->>H: POST /auth/register { email, password }
    H->>H: Validate input (email format, password length)
    H->>S: Register(email, password)
    S->>S: bcrypt.Hash(password)
    S->>R: CreateUser(email, passwordHash)
    R->>DB: INSERT INTO users ...
    DB-->>R: user row
    R-->>S: User{}
    S->>S: jwt.Sign(userID)
    S-->>H: token, User{}
    H->>H: Set-Cookie: token=<jwt>; HttpOnly; Secure; SameSite=Lax
    H-->>C: 201 { user } + auth cookie
```

---

## 2. Email/Password Login

```mermaid
sequenceDiagram
    participant C as Client
    participant H as Handler (auth.go)
    participant S as Auth Service
    participant R as User Repository
    participant DB as PostgreSQL

    C->>H: POST /auth/login { email, password }
    H->>H: Validate input
    H->>S: Login(email, password)
    S->>R: GetUserByEmail(email)
    R->>DB: SELECT * FROM users WHERE email = $1
    DB-->>R: user row (or not found)
    R-->>S: User{} or ErrNotFound
    alt user not found
        S-->>H: ErrInvalidCredentials
        H-->>C: 401 Unauthorized
    else user found
        S->>S: bcrypt.Compare(password, hash)
        alt wrong password
            S-->>H: ErrInvalidCredentials
            H-->>C: 401 Unauthorized
        else password matches
            S->>S: jwt.Sign(userID)
            S-->>H: token, User{}
            H->>H: Set-Cookie: token=<jwt>; HttpOnly; Secure; SameSite=Lax
            H-->>C: 200 { user } + auth cookie
        end
    end
```

---

## 3. OAuth Login (Google or GitHub)

```mermaid
sequenceDiagram
    participant C as Client (browser)
    participant H as Handler (auth.go)
    participant S as Auth Service
    participant O as OAuth Provider
    participant R as Repository
    participant DB as PostgreSQL

    C->>H: GET /auth/oauth/google
    H->>H: Generate state token (CSRF protection)
    H-->>C: 302 redirect → Google consent screen

    C->>O: Browser follows redirect
    O-->>C: Consent screen shown
    C->>O: User grants permission
    O-->>C: Redirect to /auth/oauth/google/callback?code=...&state=...

    C->>H: GET /auth/oauth/google/callback
    H->>H: Validate state token
    H->>S: OAuthCallback(provider, code)
    S->>O: Exchange code for access token
    O-->>S: Access token
    S->>O: Fetch user profile (email, provider ID)
    O-->>S: Profile data

    S->>R: GetUserByOAuthProvider(provider, providerUserID)
    R->>DB: SELECT user via JOIN oauth_connections
    DB-->>R: User{} or not found

    alt existing user
        S->>S: jwt.Sign(userID)
    else new user
        S->>R: CreateUser(email, noPassword)
        R->>DB: INSERT INTO users ...
        S->>R: CreateOAuthConnection(userID, provider, providerUserID)
        R->>DB: INSERT INTO oauth_connections ...
        S->>S: jwt.Sign(newUserID)
    end

    S-->>H: token, User{}
    H->>H: Set-Cookie: token=<jwt>; HttpOnly; Secure; SameSite=Lax
    H-->>C: 302 redirect → <frontend>/login?oauth=success
```

---

## 4. Authenticated Request (JWT Middleware)

Every protected endpoint runs this flow before the handler is called.

```mermaid
sequenceDiagram
    participant C as Client
    participant M as Auth Middleware
    participant H as Handler
    participant R as Repository

    C->>M: GET /accounts (Cookie: token=<jwt>)
    M->>M: Read token from "token" cookie
    M->>M: jwt.Parse(token, secret)
    alt invalid or expired token
        M-->>C: 401 Unauthorized
    else valid token
        M->>M: Extract userID from claims
        M->>M: Set userID on request context
        M->>H: Next(w, r) — pass to handler
        H->>R: GetAccounts(ctx.UserID)
        R-->>H: []Account
        H-->>C: 200 [ ...accounts ]
    end
```

---

## 5. CSV Import

```mermaid
sequenceDiagram
    participant C as Client
    participant H as Handler (imports.go)
    participant IS as Importer Service
    participant TR as Transaction Repository
    participant IR as ImportJob Repository
    participant DB as PostgreSQL

    C->>H: POST /import/csv (multipart: file, account_id)
    H->>H: Validate account belongs to user
    H->>IR: CreateImportJob(userID, fileName)
    IR->>DB: INSERT INTO import_jobs (status=pending)
    DB-->>IR: ImportJob{}
    H-->>C: 202 { job_id, status: "pending" }

    Note over H,IS: Processing continues asynchronously

    H->>IS: ProcessCSV(jobID, file, accountID)
    IS->>IR: UpdateJob(jobID, status=processing)
    IS->>IS: Parse CSV rows
    IS->>IR: UpdateJob(jobID, rows_total=N)

    loop for each batch of rows
        IS->>TR: BulkInsertTransactions(rows)
        TR->>DB: INSERT INTO transactions (classified=false) ...
        IS->>IR: UpdateJob(jobID, rows_imported += batch_size)
    end

    IS->>IR: UpdateJob(jobID, status=done, completed_at=now)
```

---

## 6. Transaction Classification (Prediction Service)

```mermaid
sequenceDiagram
    participant C as Client
    participant H as Handler (transactions.go)
    participant TS as Transaction Service
    participant PC as Predictor Client
    participant TR as Transaction Repository
    participant PS as Python Prediction Service
    participant DB as PostgreSQL

    C->>H: POST /transactions/classify
    H->>TS: ClassifyUnclassified(userID)
    TS->>TR: GetUnclassifiedTransactions(userID)
    TR->>DB: SELECT * FROM transactions WHERE user_id=$1 AND classified=false
    DB-->>TR: []Transaction
    TR-->>TS: []Transaction

    TS->>PC: Classify([]Transaction)
    PC->>PS: POST /classify { transactions: [...] }
    PS->>PS: Run ML model on each description
    PS-->>PC: { predictions: [{ id, category_name, merchant_name }, ...] }
    PC-->>TS: []Prediction

    loop for each prediction
        TS->>TR: UpdateTransaction(id, categoryName, merchantName, classified=true)
        TR->>DB: UPDATE transactions SET ...
    end

    TS-->>H: { classified: N, failed: M }
    H-->>C: 200 { classified: N, failed: M }
```

---

## 7. Manual Transaction Creation

```mermaid
sequenceDiagram
    participant C as Client
    participant H as Handler (transactions.go)
    participant TS as Transaction Service
    participant TR as Transaction Repository
    participant AR as Account Repository
    participant DB as PostgreSQL

    C->>H: POST /transactions { account_id, amount, description, date, category_id? }
    H->>H: Validate input
    H->>TS: CreateTransaction(userID, input)
    TS->>AR: GetAccount(accountID) — verify ownership
    AR->>DB: SELECT * FROM accounts WHERE id=$1 AND user_id=$2
    DB-->>AR: Account{}

    TS->>DB: BEGIN transaction
    TS->>TR: InsertTransaction(...)
    TR->>DB: INSERT INTO transactions (source=manual, classified=true)
    TS->>AR: UpdateBalance(accountID, amount)
    AR->>DB: UPDATE accounts SET balance = balance + $1
    TS->>DB: COMMIT

    TS-->>H: Transaction{}
    H-->>C: 201 { ...transaction }
```
