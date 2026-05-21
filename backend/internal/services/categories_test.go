package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
)

type mockCategoryRepo struct {
	listFn     func(ctx context.Context, userID uuid.UUID) ([]models.Category, error)
	getByIDFn  func(ctx context.Context, id, userID uuid.UUID) (*models.Category, error)
	createFn   func(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID, color *string) (*models.Category, error)
	updateFn   func(ctx context.Context, id, userID uuid.UUID, fields repository.CategoryUpdateFields) (*models.Category, error)
	deleteFn   func(ctx context.Context, id, userID uuid.UUID) error
}

func (m *mockCategoryRepo) List(ctx context.Context, userID uuid.UUID) ([]models.Category, error) {
	return m.listFn(ctx, userID)
}
func (m *mockCategoryRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Category, error) {
	return m.getByIDFn(ctx, id, userID)
}
func (m *mockCategoryRepo) Create(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID, color *string) (*models.Category, error) {
	return m.createFn(ctx, userID, name, parentID, color)
}
func (m *mockCategoryRepo) Update(ctx context.Context, id, userID uuid.UUID, fields repository.CategoryUpdateFields) (*models.Category, error) {
	return m.updateFn(ctx, id, userID, fields)
}
func (m *mockCategoryRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.deleteFn(ctx, id, userID)
}

func makeCategory(userID uuid.UUID, parentID *uuid.UUID) *models.Category {
	return &models.Category{
		ID:        uuid.New(),
		UserID:    userID,
		ParentID:  parentID,
		Name:      "Test Category",
		CreatedAt: time.Now(),
	}
}

func TestCategoryServiceCreateWithValidParent(t *testing.T) {
	userID := uuid.New()
	parentID := uuid.New()
	parent := makeCategory(userID, nil) // parent is a Group (no parentID)
	parent.ID = parentID
	child := makeCategory(userID, &parentID)

	repo := &mockCategoryRepo{
		createFn: func(ctx context.Context, uid uuid.UUID, name string, pid *uuid.UUID, color *string) (*models.Category, error) {
			return child, nil
		},
	}

	svc := NewCategoryService(repo)
	req := CreateCategoryRequest{Name: "Movies", ParentID: &parentID}
	result, err := svc.Create(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ParentID == nil || *result.ParentID != parentID {
		t.Errorf("expected parent_id %v, got %v", parentID, result.ParentID)
	}
}

func TestCategoryServiceCreateParentIsCategory(t *testing.T) {
	userID := uuid.New()
	parentID := uuid.New()

	repo := &mockCategoryRepo{
		createFn: func(ctx context.Context, uid uuid.UUID, name string, pid *uuid.UUID, color *string) (*models.Category, error) {
			return nil, fmt.Errorf("parent_id must reference a group, not a category")
		},
	}

	svc := NewCategoryService(repo)
	req := CreateCategoryRequest{Name: "SubCategory", ParentID: &parentID}
	_, err := svc.Create(context.Background(), userID, req)
	if err == nil {
		t.Error("expected error when parent is a category, got nil")
	}
}

func TestCategoryServiceDeleteGroup(t *testing.T) {
	userID := uuid.New()
	groupID := uuid.New()

	deleted := false
	repo := &mockCategoryRepo{
		deleteFn: func(ctx context.Context, id, uid uuid.UUID) error {
			deleted = true
			return nil
		},
	}

	svc := NewCategoryService(repo)
	err := svc.Delete(context.Background(), groupID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("expected delete to be called")
	}
}

func TestCategoryServiceDeleteNotFound(t *testing.T) {
	repo := &mockCategoryRepo{
		deleteFn: func(ctx context.Context, id, userID uuid.UUID) error {
			return repository.ErrNotFound
		},
	}

	svc := NewCategoryService(repo)
	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCategoryServiceList(t *testing.T) {
	userID := uuid.New()
	cats := []models.Category{
		*makeCategory(userID, nil),
		*makeCategory(userID, nil),
	}

	repo := &mockCategoryRepo{
		listFn: func(ctx context.Context, uid uuid.UUID) ([]models.Category, error) {
			if uid != userID {
				t.Errorf("expected userID %v, got %v", userID, uid)
			}
			return cats, nil
		},
	}

	svc := NewCategoryService(repo)
	result, err := svc.List(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 categories, got %d", len(result))
	}
}
