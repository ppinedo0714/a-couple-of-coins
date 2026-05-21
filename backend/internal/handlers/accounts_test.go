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
	"github.com/ppinedo/a-couple-of-coins/backend/internal/auth"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/services"
)

type mockAccountService struct {
	listFn       func(ctx context.Context, userID uuid.UUID) ([]models.Account, error)
	getFn        func(ctx context.Context, id, userID uuid.UUID) (*models.Account, error)
	createFn     func(ctx context.Context, userID uuid.UUID, req services.CreateAccountRequest) (*models.Account, error)
	updateFn     func(ctx context.Context, id, userID uuid.UUID, req services.UpdateAccountRequest) (*models.Account, error)
	deleteFn     func(ctx context.Context, id, userID uuid.UUID) error
	getHistoryFn func(ctx context.Context, userID uuid.UUID, req services.HistoryRequest) ([]models.BalanceSnapshot, error)
}

func (m *mockAccountService) List(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	return m.listFn(ctx, userID)
}
func (m *mockAccountService) Get(ctx context.Context, id, userID uuid.UUID) (*models.Account, error) {
	return m.getFn(ctx, id, userID)
}
func (m *mockAccountService) Create(ctx context.Context, userID uuid.UUID, req services.CreateAccountRequest) (*models.Account, error) {
	return m.createFn(ctx, userID, req)
}
func (m *mockAccountService) Update(ctx context.Context, id, userID uuid.UUID, req services.UpdateAccountRequest) (*models.Account, error) {
	return m.updateFn(ctx, id, userID, req)
}
func (m *mockAccountService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.deleteFn(ctx, id, userID)
}
func (m *mockAccountService) GetHistory(ctx context.Context, userID uuid.UUID, req services.HistoryRequest) ([]models.BalanceSnapshot, error) {
	return m.getHistoryFn(ctx, userID, req)
}

func newMockAccountsHandler(svc accountServiceIface) *AccountsHandler {
	return &AccountsHandler{svc: svc}
}

// buildAccountRouter wires up an AccountsHandler behind real JWT middleware.
func buildAccountRouter(svc accountServiceIface) (*chi.Mux, uuid.UUID, string) {
	userID := uuid.New()
	token, _ := auth.Issue(testJWTSecret, userID, auth.TokenTTL)

	h := newMockAccountsHandler(svc)
	r := chi.NewRouter()
	r.Use(auth.Middleware(testJWTSecret))
	r.Get("/accounts/history", h.History)
	r.Get("/accounts", h.List)
	r.Post("/accounts", h.Create)
	r.Get("/accounts/{id}", h.Get)
	r.Put("/accounts/{id}", h.Update)
	r.Delete("/accounts/{id}", h.Delete)
	return r, userID, token
}

func withToken(req *http.Request, token string) *http.Request {
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	return req
}

func sampleAccount(userID uuid.UUID) *models.Account {
	return &models.Account{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      "Test Checking",
		Type:      "checking",
		Balance:   100.00,
		Currency:  "USD",
		CreatedAt: time.Now(),
	}
}

func TestAccountHandlerList(t *testing.T) {
	svc := &mockAccountService{
		listFn: func(ctx context.Context, uid uuid.UUID) ([]models.Account, error) {
			return []models.Account{}, nil
		},
	}
	r, _, token := buildAccountRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodGet, "/accounts", nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAccountHandlerCreateValid(t *testing.T) {
	svc := &mockAccountService{
		createFn: func(ctx context.Context, uid uuid.UUID, req services.CreateAccountRequest) (*models.Account, error) {
			return sampleAccount(uid), nil
		},
	}
	r, _, token := buildAccountRouter(svc)

	body, _ := json.Marshal(map[string]interface{}{
		"name": "My Checking",
		"type": "checking",
	})
	req := withToken(httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body)), token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestAccountHandlerCreateInvalidType(t *testing.T) {
	svc := &mockAccountService{
		createFn: func(ctx context.Context, uid uuid.UUID, req services.CreateAccountRequest) (*models.Account, error) {
			return nil, errors.New("type must be one of: checking, savings, credit, investment")
		},
	}
	r, _, token := buildAccountRouter(svc)

	body, _ := json.Marshal(map[string]interface{}{
		"name": "Bad Account",
		"type": "invalid",
	})
	req := withToken(httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body)), token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAccountHandlerGetFound(t *testing.T) {
	accountID := uuid.New()
	svc := &mockAccountService{
		getFn: func(ctx context.Context, id, uid uuid.UUID) (*models.Account, error) {
			return &models.Account{ID: accountID, UserID: uid, Type: "savings"}, nil
		},
	}
	r, _, token := buildAccountRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodGet, "/accounts/"+accountID.String(), nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAccountHandlerGetNotFound(t *testing.T) {
	svc := &mockAccountService{
		getFn: func(ctx context.Context, id, uid uuid.UUID) (*models.Account, error) {
			return nil, repository.ErrNotFound
		},
	}
	r, _, token := buildAccountRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodGet, "/accounts/"+uuid.New().String(), nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestAccountHandlerDeleteHasTransactions(t *testing.T) {
	svc := &mockAccountService{
		deleteFn: func(ctx context.Context, id, uid uuid.UUID) error {
			return repository.ErrHasTransactions
		},
	}
	r, _, token := buildAccountRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodDelete, "/accounts/"+uuid.New().String(), nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestAccountHandlerHistoryMissingDates(t *testing.T) {
	svc := &mockAccountService{}
	r, _, token := buildAccountRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodGet, "/accounts/history", nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAccountHandlerHistorySuccess(t *testing.T) {
	accountID := uuid.New()
	now := time.Now()

	svc := &mockAccountService{
		getHistoryFn: func(ctx context.Context, uid uuid.UUID, req services.HistoryRequest) ([]models.BalanceSnapshot, error) {
			return []models.BalanceSnapshot{
				{AccountID: accountID, Date: now, Balance: 500.00},
			}, nil
		},
	}
	r, _, token := buildAccountRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodGet, "/accounts/history?from=2025-01-01&to=2025-01-31", nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAccountHandlerRequiresAuth(t *testing.T) {
	svc := &mockAccountService{}
	r, _, _ := buildAccountRouter(svc)

	// No token cookie — middleware should reject.
	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
