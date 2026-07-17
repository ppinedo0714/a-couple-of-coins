package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/auth"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
)

const testJWTSecret = "test-secret-that-is-at-least-32-chars-long"

// mockUserRepo is a controllable in-memory implementation of UserRepository.
type mockUserRepo struct {
	createFn              func(ctx context.Context, email, passwordHash string) (*models.User, error)
	getByEmailFn          func(ctx context.Context, email string) (*models.User, error)
	getByIDFn             func(ctx context.Context, id uuid.UUID) (*models.User, error)
	getByOAuthProviderFn  func(ctx context.Context, provider, providerUserID string) (*models.User, error)
	updateFn              func(ctx context.Context, id uuid.UUID, email string) (*models.User, error)
	createOAuthConnFn     func(ctx context.Context, userID uuid.UUID, provider, providerUserID string) error
}

func (m *mockUserRepo) Create(ctx context.Context, email, passwordHash string) (*models.User, error) {
	return m.createFn(ctx, email, passwordHash)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.getByEmailFn(ctx, email)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockUserRepo) GetByOAuthProvider(ctx context.Context, provider, providerUserID string) (*models.User, error) {
	return m.getByOAuthProviderFn(ctx, provider, providerUserID)
}
func (m *mockUserRepo) Update(ctx context.Context, id uuid.UUID, email string) (*models.User, error) {
	return m.updateFn(ctx, id, email)
}
func (m *mockUserRepo) CreateOAuthConnection(ctx context.Context, userID uuid.UUID, provider, providerUserID string) error {
	return m.createOAuthConnFn(ctx, userID, provider, providerUserID)
}

func newTestAuthHandler(repo repository.UserRepository) *AuthHandler {
	return NewAuthHandler(repo, testJWTSecret, "http://localhost:3000", nil, nil)
}

func makeUser(email string) *models.User {
	hash, _ := auth.Hash("password123")
	return &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: &hash,
		CreatedAt:    time.Now(),
	}
}

func TestRegisterHappyPath(t *testing.T) {
	user := makeUser("test@example.com")
	repo := &mockUserRepo{
		createFn: func(ctx context.Context, email, passwordHash string) (*models.User, error) {
			return user, nil
		},
	}

	h := newTestAuthHandler(repo)
	body, _ := json.Marshal(map[string]string{"email": "test@example.com", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	// Cookie must be set
	cookies := rec.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "token" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("auth cookie not set in response")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(ctx context.Context, email, passwordHash string) (*models.User, error) {
			// Simulate a unique_violation from Postgres
			return nil, &pgconn.PgError{Code: "23505"}
		},
	}

	h := newTestAuthHandler(repo)
	body, _ := json.Marshal(map[string]string{"email": "dup@example.com", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	hash, _ := auth.Hash("correctpassword")
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: &hash,
		CreatedAt:    time.Now(),
	}

	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return user, nil
		},
	}

	h := newTestAuthHandler(repo)
	body, _ := json.Marshal(map[string]string{"email": "test@example.com", "password": "wrongpassword"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestLoginUnknownEmail(t *testing.T) {
	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return nil, repository.ErrNotFound
		},
	}

	h := newTestAuthHandler(repo)
	body, _ := json.Marshal(map[string]string{"email": "nobody@example.com", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestLogout(t *testing.T) {
	r := chi.NewRouter()
	r.Use(auth.Middleware(testJWTSecret))

	userID := uuid.New()
	token, _ := auth.Issue(testJWTSecret, userID, auth.TokenTTL)

	repo := &mockUserRepo{}
	h := newTestAuthHandler(repo)
	r.Post("/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	// Cookie should be cleared (MaxAge=0 or negative)
	cookies := rec.Result().Cookies()
	var cleared bool
	for _, c := range cookies {
		if c.Name == "token" && c.MaxAge == 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Error("auth cookie was not cleared on logout")
	}
}

func TestLoginHappyPath(t *testing.T) {
	password := "password123"
	hash, _ := auth.Hash(password)
	user := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: &hash,
		CreatedAt:    time.Now(),
	}

	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return user, nil
		},
	}

	h := newTestAuthHandler(repo)
	body, _ := json.Marshal(map[string]string{"email": "test@example.com", "password": password})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	cookies := rec.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "token" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("auth cookie not set on login")
	}
}

func TestRegisterInvalidEmail(t *testing.T) {
	repo := &mockUserRepo{}
	h := newTestAuthHandler(repo)

	body, _ := json.Marshal(map[string]string{"email": "notanemail", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRegisterShortPassword(t *testing.T) {
	repo := &mockUserRepo{}
	h := newTestAuthHandler(repo)

	body, _ := json.Marshal(map[string]string{"email": "test@example.com", "password": "short"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// Verify the error responses use the standard shape.
func TestErrorResponseShape(t *testing.T) {
	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return nil, repository.ErrNotFound
		},
	}

	h := newTestAuthHandler(repo)
	body, _ := json.Marshal(map[string]string{"email": "x@x.com", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := resp["error"]; !ok {
		t.Error("response body missing 'error' key")
	}
}

// Ensure missing required field returns 401 from login (not a panic).
func TestLoginEmptyBody(t *testing.T) {
	repo := &mockUserRepo{}
	h := newTestAuthHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRegisterPasswordTooShort(t *testing.T) {
	repo := &mockUserRepo{}
	h := newTestAuthHandler(repo)

	body, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "1234567"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// Register with exactly 8-char password should succeed.
func TestRegisterPasswordExactMinLength(t *testing.T) {
	user := makeUser("a@b.com")
	repo := &mockUserRepo{
		createFn: func(ctx context.Context, email, passwordHash string) (*models.User, error) {
			return user, nil
		},
	}

	h := newTestAuthHandler(repo)
	body, _ := json.Marshal(map[string]string{"email": "a@b.com", "password": "12345678"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestLoginWithNoPasswordHash(t *testing.T) {
	// OAuth-only user has no password hash
	user := &models.User{
		ID:           uuid.New(),
		Email:        "oauth@example.com",
		PasswordHash: nil,
		CreatedAt:    time.Now(),
	}

	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return user, nil
		},
	}

	h := newTestAuthHandler(repo)
	body, _ := json.Marshal(map[string]string{"email": "oauth@example.com", "password": "anypassword"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d; OAuth-only user should not be able to login with password", rec.Code, http.StatusUnauthorized)
	}
}

// Ensure errors in Create are classified correctly
func TestRegisterInternalError(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(ctx context.Context, email, passwordHash string) (*models.User, error) {
			return nil, errors.New("unexpected db error")
		},
	}

	h := newTestAuthHandler(repo)
	body, _ := json.Marshal(map[string]string{"email": "test@example.com", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
