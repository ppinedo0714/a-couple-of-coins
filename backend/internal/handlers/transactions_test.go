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
	"github.com/ppinedo/a-couple-of-coins/backend/internal/auth"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
)

// --- mock service ---

type mockTransactionService struct {
	listFn     func(ctx context.Context, userID uuid.UUID, filters repository.TransactionFilters) ([]models.Transaction, int, error)
	getFn      func(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error)
	createFn   func(ctx context.Context, userID uuid.UUID, req models.CreateTransactionRequest) (*models.Transaction, error)
	updateFn   func(ctx context.Context, id, userID uuid.UUID, req models.UpdateTransactionRequest) (*models.Transaction, error)
	deleteFn   func(ctx context.Context, id, userID uuid.UUID) error
	classifyFn func(ctx context.Context, userID uuid.UUID) (models.ClassifyResult, error)

	capturedFilters repository.TransactionFilters
	capturedUpdate  models.UpdateTransactionRequest
}

func (m *mockTransactionService) List(ctx context.Context, userID uuid.UUID, filters repository.TransactionFilters) ([]models.Transaction, int, error) {
	m.capturedFilters = filters
	if m.listFn != nil {
		return m.listFn(ctx, userID, filters)
	}
	return nil, 0, nil
}
func (m *mockTransactionService) Get(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id, userID)
	}
	return nil, nil
}
func (m *mockTransactionService) Create(ctx context.Context, userID uuid.UUID, req models.CreateTransactionRequest) (*models.Transaction, error) {
	if m.createFn != nil {
		return m.createFn(ctx, userID, req)
	}
	return nil, nil
}
func (m *mockTransactionService) Update(ctx context.Context, id, userID uuid.UUID, req models.UpdateTransactionRequest) (*models.Transaction, error) {
	m.capturedUpdate = req
	if m.updateFn != nil {
		return m.updateFn(ctx, id, userID, req)
	}
	return nil, nil
}
func (m *mockTransactionService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, userID)
	}
	return nil
}
func (m *mockTransactionService) ClassifyUnclassified(ctx context.Context, userID uuid.UUID) (models.ClassifyResult, error) {
	if m.classifyFn != nil {
		return m.classifyFn(ctx, userID)
	}
	return models.ClassifyResult{}, nil
}

// --- helpers ---

func buildTransactionRouter(svc transactionServiceIface) (*chi.Mux, uuid.UUID, string) {
	userID := uuid.New()
	token, _ := auth.Issue(testJWTSecret, userID, auth.TokenTTL)

	h := &TransactionsHandler{svc: svc}
	r := chi.NewRouter()
	r.Use(auth.Middleware(testJWTSecret))
	// classify must be before /{id}
	r.Post("/transactions/classify", h.Classify)
	r.Get("/transactions", h.List)
	r.Post("/transactions", h.Create)
	r.Get("/transactions/{id}", h.Get)
	r.Put("/transactions/{id}", h.Update)
	r.Delete("/transactions/{id}", h.Delete)
	return r, userID, token
}

func sampleTransaction(userID uuid.UUID) *models.Transaction {
	return &models.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		AccountID:   uuid.New(),
		Amount:      -42.50,
		Description: "Test transaction",
		Date:        time.Now(),
		Source:      "manual",
		Classified:  true,
		CreatedAt:   time.Now(),
	}
}

// --- tests ---

func TestTransactionList_NoFilters_Returns200WithPaginationFields(t *testing.T) {
	svc := &mockTransactionService{
		listFn: func(_ context.Context, uid uuid.UUID, _ repository.TransactionFilters) ([]models.Transaction, int, error) {
			return []models.Transaction{*sampleTransaction(uid)}, 1, nil
		},
	}
	r, _, token := buildTransactionRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodGet, "/transactions", nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, field := range []string{"transactions", "total", "limit", "offset"} {
		if _, ok := resp[field]; !ok {
			t.Errorf("want field %q in response", field)
		}
	}
}

func TestTransactionList_WithFilters_PassesFiltersToService(t *testing.T) {
	accountID := uuid.New()
	svc := &mockTransactionService{}
	r, _, token := buildTransactionRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodGet, "/transactions?account_id="+accountID.String()+"&limit=10&offset=5", nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
	if svc.capturedFilters.AccountID == nil || *svc.capturedFilters.AccountID != accountID {
		t.Errorf("want AccountID=%v in filters, got %v", accountID, svc.capturedFilters.AccountID)
	}
	if svc.capturedFilters.Limit != 10 {
		t.Errorf("want limit=10, got %d", svc.capturedFilters.Limit)
	}
	if svc.capturedFilters.Offset != 5 {
		t.Errorf("want offset=5, got %d", svc.capturedFilters.Offset)
	}
}

func TestTransactionCreate_ValidRequest_Returns201(t *testing.T) {
	svc := &mockTransactionService{
		createFn: func(_ context.Context, uid uuid.UUID, _ models.CreateTransactionRequest) (*models.Transaction, error) {
			return sampleTransaction(uid), nil
		},
	}
	r, _, token := buildTransactionRouter(svc)

	body, _ := json.Marshal(models.CreateTransactionRequest{
		AccountID:   uuid.New(),
		Amount:      -42.50,
		Description: "Whole Foods",
		Date:        "2024-01-15",
	})
	req := withToken(httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body)), token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTransactionCreate_InvalidDate_Returns400(t *testing.T) {
	svc := &mockTransactionService{
		createFn: func(_ context.Context, _ uuid.UUID, _ models.CreateTransactionRequest) (*models.Transaction, error) {
			return nil, &handlerValidationError{"invalid date, use YYYY-MM-DD"}
		},
	}
	r, _, token := buildTransactionRouter(svc)

	body, _ := json.Marshal(map[string]interface{}{
		"account_id":  uuid.New().String(),
		"amount":      -10.0,
		"description": "test",
		"date":        "not-a-date",
	})
	req := withToken(httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body)), token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestTransactionGet_Found_Returns200(t *testing.T) {
	svc := &mockTransactionService{}
	r, userID, token := buildTransactionRouter(svc)
	txn := sampleTransaction(userID)
	svc.getFn = func(_ context.Context, _, _ uuid.UUID) (*models.Transaction, error) {
		return txn, nil
	}

	req := withToken(httptest.NewRequest(http.MethodGet, "/transactions/"+txn.ID.String(), nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
}

func TestTransactionGet_NotFound_Returns404(t *testing.T) {
	svc := &mockTransactionService{
		getFn: func(_ context.Context, _, _ uuid.UUID) (*models.Transaction, error) {
			return nil, repository.ErrNotFound
		},
	}
	r, _, token := buildTransactionRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodGet, "/transactions/"+uuid.New().String(), nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
}

func TestTransactionUpdate_ValidRequest_Returns200(t *testing.T) {
	svc := &mockTransactionService{}
	r, userID, token := buildTransactionRouter(svc)
	txn := sampleTransaction(userID)
	svc.updateFn = func(_ context.Context, _, _ uuid.UUID, _ models.UpdateTransactionRequest) (*models.Transaction, error) {
		return txn, nil
	}

	body := []byte(`{"description": "Updated description"}`)
	req := withToken(httptest.NewRequest(http.MethodPut, "/transactions/"+txn.ID.String(), bytes.NewReader(body)), token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTransactionUpdate_ExplicitNullCategoryID_SetsClearCategory(t *testing.T) {
	svc := &mockTransactionService{}
	r, userID, token := buildTransactionRouter(svc)
	txn := sampleTransaction(userID)
	svc.updateFn = func(_ context.Context, _, _ uuid.UUID, _ models.UpdateTransactionRequest) (*models.Transaction, error) {
		return txn, nil
	}

	body := []byte(`{"category_id": null}`)
	req := withToken(httptest.NewRequest(http.MethodPut, "/transactions/"+txn.ID.String(), bytes.NewReader(body)), token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !svc.capturedUpdate.ClearCategory {
		t.Error("want ClearCategory=true when category_id is explicitly null")
	}
}

func TestTransactionDelete_Found_Returns204(t *testing.T) {
	svc := &mockTransactionService{
		deleteFn: func(_ context.Context, _, _ uuid.UUID) error { return nil },
	}
	r, _, token := buildTransactionRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodDelete, "/transactions/"+uuid.New().String(), nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("want 204, got %d", rec.Code)
	}
}

func TestTransactionClassify_Returns200WithCounts(t *testing.T) {
	svc := &mockTransactionService{
		classifyFn: func(_ context.Context, _ uuid.UUID) (models.ClassifyResult, error) {
			return models.ClassifyResult{Classified: 5, Failed: 1}, nil
		},
	}
	r, _, token := buildTransactionRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodPost, "/transactions/classify", nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}

	var result models.ClassifyResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Classified != 5 {
		t.Errorf("want classified=5, got %d", result.Classified)
	}
	if result.Failed != 1 {
		t.Errorf("want failed=1, got %d", result.Failed)
	}
}

// handlerValidationError is a simple error for service-level validation failures.
type handlerValidationError struct{ msg string }

func (e *handlerValidationError) Error() string { return e.msg }
