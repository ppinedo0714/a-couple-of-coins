package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
)

type CategoryUpdateFields struct {
	Name  *string
	Color *string
}

type CategoryRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]models.Category, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Category, error)
	Create(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID, color *string) (*models.Category, error)
	Update(ctx context.Context, id, userID uuid.UUID, fields CategoryUpdateFields) (*models.Category, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type pgxCategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) CategoryRepository {
	return &pgxCategoryRepository{pool: pool}
}

func (r *pgxCategoryRepository) List(ctx context.Context, userID uuid.UUID) ([]models.Category, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, parent_id, name, color, created_at
		 FROM categories
		 WHERE user_id = $1
		 ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.ParentID, &c.Name, &c.Color, &c.CreatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *pgxCategoryRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Category, error) {
	var c models.Category
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, parent_id, name, color, created_at
		 FROM categories
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&c.ID, &c.UserID, &c.ParentID, &c.Name, &c.Color, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *pgxCategoryRepository) Create(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID, color *string) (*models.Category, error) {
	if parentID != nil {
		var parentParentID *uuid.UUID
		err := r.pool.QueryRow(ctx,
			`SELECT parent_id FROM categories WHERE id = $1 AND user_id = $2`,
			*parentID, userID,
		).Scan(&parentParentID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("parent category not found")
		}
		if err != nil {
			return nil, err
		}
		if parentParentID != nil {
			return nil, fmt.Errorf("parent_id must reference a group, not a category")
		}
	}

	var c models.Category
	err := r.pool.QueryRow(ctx,
		`INSERT INTO categories (user_id, parent_id, name, color)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, parent_id, name, color, created_at`,
		userID, parentID, name, color,
	).Scan(&c.ID, &c.UserID, &c.ParentID, &c.Name, &c.Color, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *pgxCategoryRepository) Update(ctx context.Context, id, userID uuid.UUID, fields CategoryUpdateFields) (*models.Category, error) {
	var c models.Category
	err := r.pool.QueryRow(ctx,
		`UPDATE categories
		 SET
		   name  = COALESCE($3, name),
		   color = CASE WHEN parent_id IS NULL THEN COALESCE($4, color) ELSE color END
		 WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, parent_id, name, color, created_at`,
		id, userID, fields.Name, fields.Color,
	).Scan(&c.ID, &c.UserID, &c.ParentID, &c.Name, &c.Color, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *pgxCategoryRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM categories WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
