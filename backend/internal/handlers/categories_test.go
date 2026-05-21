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

type mockCategoryService struct {
	listFn   func(ctx context.Context, userID uuid.UUID) ([]models.Category, error)
	getFn    func(ctx context.Context, id, userID uuid.UUID) (*models.Category, error)
	createFn func(ctx context.Context, userID uuid.UUID, req services.CreateCategoryRequest) (*models.Category, error)
	updateFn func(ctx context.Context, id, userID uuid.UUID, req services.UpdateCategoryRequest) (*models.Category, error)
	deleteFn func(ctx context.Context, id, userID uuid.UUID) error
}

func (m *mockCategoryService) List(ctx context.Context, userID uuid.UUID) ([]models.Category, error) {
	return m.listFn(ctx, userID)
}
func (m *mockCategoryService) Get(ctx context.Context, id, userID uuid.UUID) (*models.Category, error) {
	return m.getFn(ctx, id, userID)
}
func (m *mockCategoryService) Create(ctx context.Context, userID uuid.UUID, req services.CreateCategoryRequest) (*models.Category, error) {
	return m.createFn(ctx, userID, req)
}
func (m *mockCategoryService) Update(ctx context.Context, id, userID uuid.UUID, req services.UpdateCategoryRequest) (*models.Category, error) {
	return m.updateFn(ctx, id, userID, req)
}
func (m *mockCategoryService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.deleteFn(ctx, id, userID)
}

func newMockCategoriesHandler(svc categoryServiceIface) *CategoriesHandler {
	return &CategoriesHandler{svc: svc}
}

func buildCategoryRouter(svc categoryServiceIface) (*chi.Mux, uuid.UUID, string) {
	userID := uuid.New()
	token, _ := auth.Issue(testJWTSecret, userID, auth.TokenTTL)

	h := newMockCategoriesHandler(svc)
	r := chi.NewRouter()
	r.Use(auth.Middleware(testJWTSecret))
	r.Get("/categories", h.List)
	r.Post("/categories", h.Create)
	r.Put("/categories/{id}", h.Update)
	r.Delete("/categories/{id}", h.Delete)
	return r, userID, token
}

func sampleCategory(userID uuid.UUID, parentID *uuid.UUID) *models.Category {
	return &models.Category{
		ID:        uuid.New(),
		UserID:    userID,
		ParentID:  parentID,
		Name:      "Entertainment",
		CreatedAt: time.Now(),
	}
}

func TestCategoryHandlerList(t *testing.T) {
	svc := &mockCategoryService{
		listFn: func(ctx context.Context, uid uuid.UUID) ([]models.Category, error) {
			return []models.Category{}, nil
		},
	}
	r, _, token := buildCategoryRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodGet, "/categories", nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestCategoryHandlerCreateGroup(t *testing.T) {
	color := "#EC407A"
	svc := &mockCategoryService{
		createFn: func(ctx context.Context, uid uuid.UUID, req services.CreateCategoryRequest) (*models.Category, error) {
			cat := sampleCategory(uid, nil)
			cat.Color = &color
			return cat, nil
		},
	}
	r, _, token := buildCategoryRouter(svc)

	body, _ := json.Marshal(map[string]interface{}{
		"name":  "Entertainment",
		"color": color,
	})
	req := withToken(httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body)), token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestCategoryHandlerCreateCategory(t *testing.T) {
	parentID := uuid.New()
	svc := &mockCategoryService{
		createFn: func(ctx context.Context, uid uuid.UUID, req services.CreateCategoryRequest) (*models.Category, error) {
			cat := sampleCategory(uid, req.ParentID)
			cat.Name = "Movies"
			return cat, nil
		},
	}
	r, _, token := buildCategoryRouter(svc)

	body, _ := json.Marshal(map[string]interface{}{
		"name":      "Movies",
		"parent_id": parentID.String(),
	})
	req := withToken(httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body)), token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestCategoryHandlerCreateInvalidParent(t *testing.T) {
	svc := &mockCategoryService{
		createFn: func(ctx context.Context, uid uuid.UUID, req services.CreateCategoryRequest) (*models.Category, error) {
			return nil, errors.New("parent_id must reference a group, not a category")
		},
	}
	r, _, token := buildCategoryRouter(svc)

	parentID := uuid.New()
	body, _ := json.Marshal(map[string]interface{}{
		"name":      "SubSub",
		"parent_id": parentID.String(),
	})
	req := withToken(httptest.NewRequest(http.MethodPost, "/categories", bytes.NewReader(body)), token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCategoryHandlerDelete(t *testing.T) {
	catID := uuid.New()
	svc := &mockCategoryService{
		deleteFn: func(ctx context.Context, id, uid uuid.UUID) error {
			return nil
		},
	}
	r, _, token := buildCategoryRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodDelete, "/categories/"+catID.String(), nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestCategoryHandlerDeleteNotFound(t *testing.T) {
	svc := &mockCategoryService{
		deleteFn: func(ctx context.Context, id, uid uuid.UUID) error {
			return repository.ErrNotFound
		},
	}
	r, _, token := buildCategoryRouter(svc)

	req := withToken(httptest.NewRequest(http.MethodDelete, "/categories/"+uuid.New().String(), nil), token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
