package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
)

type CreateCategoryRequest struct {
	Name     string     `json:"name"`
	ParentID *uuid.UUID `json:"parent_id"`
	Color    *string    `json:"color"`
}

type UpdateCategoryRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

type CategoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List(ctx context.Context, userID uuid.UUID) ([]models.Category, error) {
	return s.repo.List(ctx, userID)
}

func (s *CategoryService) Get(ctx context.Context, id, userID uuid.UUID) (*models.Category, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *CategoryService) Create(ctx context.Context, userID uuid.UUID, req CreateCategoryRequest) (*models.Category, error) {
	return s.repo.Create(ctx, userID, req.Name, req.ParentID, req.Color)
}

func (s *CategoryService) Update(ctx context.Context, id, userID uuid.UUID, req UpdateCategoryRequest) (*models.Category, error) {
	fields := repository.CategoryUpdateFields{
		Name:  req.Name,
		Color: req.Color,
	}
	return s.repo.Update(ctx, id, userID, fields)
}

func (s *CategoryService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}
