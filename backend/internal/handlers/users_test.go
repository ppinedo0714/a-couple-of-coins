package handlers

import (
	"bytes"
	"context"
	"encoding/json"
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

func newTestRouter(repo repository.UserRepository) *chi.Mux {
	usersHandler := NewUsersHandler(repo, testJWTSecret)
	r := chi.NewRouter()
	r.Use(auth.Middleware(testJWTSecret))
	r.Get("/users/me", usersHandler.GetMe)
	r.Put("/users/me", usersHandler.UpdateMe)
	return r
}

func makeAuthRequest(t *testing.T, method, path string, body []byte, userID uuid.UUID) *http.Request {
	t.Helper()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")

	token, err := auth.Issue(testJWTSecret, userID, auth.TokenTTL)
	if err != nil {
		t.Fatalf("failed to issue test token: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	return req
}

func TestGetMeSuccess(t *testing.T) {
	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		Email:     "me@example.com",
		CreatedAt: time.Now(),
	}

	repo := &mockUserRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
			if id != userID {
				t.Errorf("GetByID called with wrong id: %v", id)
			}
			return user, nil
		},
	}

	r := newTestRouter(repo)
	req := makeAuthRequest(t, http.MethodGet, "/users/me", nil, userID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["email"] != user.Email {
		t.Errorf("email = %v, want %v", resp["email"], user.Email)
	}
	// password_hash must not be in response
	if _, ok := resp["password_hash"]; ok {
		t.Error("response should not contain password_hash")
	}
}

func TestGetMeNoAuth(t *testing.T) {
	repo := &mockUserRepo{}
	r := newTestRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUpdateMeSuccess(t *testing.T) {
	userID := uuid.New()
	updatedUser := &models.User{
		ID:        userID,
		Email:     "new@example.com",
		CreatedAt: time.Now(),
	}

	repo := &mockUserRepo{
		updateFn: func(ctx context.Context, id uuid.UUID, email string) (*models.User, error) {
			return updatedUser, nil
		},
	}

	r := newTestRouter(repo)
	body, _ := json.Marshal(map[string]string{"email": "new@example.com"})
	req := makeAuthRequest(t, http.MethodPut, "/users/me", body, userID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["email"] != "new@example.com" {
		t.Errorf("email = %v, want new@example.com", resp["email"])
	}
}

func TestUpdateMeEmailTaken(t *testing.T) {
	userID := uuid.New()

	repo := &mockUserRepo{
		updateFn: func(ctx context.Context, id uuid.UUID, email string) (*models.User, error) {
			return nil, &pgconn.PgError{Code: "23505"}
		},
	}

	r := newTestRouter(repo)
	body, _ := json.Marshal(map[string]string{"email": "taken@example.com"})
	req := makeAuthRequest(t, http.MethodPut, "/users/me", body, userID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestUpdateMeInvalidEmail(t *testing.T) {
	userID := uuid.New()
	repo := &mockUserRepo{}

	r := newTestRouter(repo)
	body, _ := json.Marshal(map[string]string{"email": "notvalid"})
	req := makeAuthRequest(t, http.MethodPut, "/users/me", body, userID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateMeNotFound(t *testing.T) {
	userID := uuid.New()

	repo := &mockUserRepo{
		updateFn: func(ctx context.Context, id uuid.UUID, email string) (*models.User, error) {
			return nil, repository.ErrNotFound
		},
	}

	r := newTestRouter(repo)
	body, _ := json.Marshal(map[string]string{"email": "new@example.com"})
	req := makeAuthRequest(t, http.MethodPut, "/users/me", body, userID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
