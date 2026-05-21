package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
)

type ImportJobRepository interface {
	Create(ctx context.Context, userID uuid.UUID, fileName string) (*models.ImportJob, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*models.ImportJob, error)
	List(ctx context.Context, userID uuid.UUID) ([]models.ImportJob, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateRowsTotal(ctx context.Context, id uuid.UUID, rowsTotal int) error
	IncrementRowsImported(ctx context.Context, id uuid.UUID, count int) error
	Complete(ctx context.Context, id uuid.UUID, status string) error
}

type pgxImportJobRepository struct {
	pool *pgxpool.Pool
}

func NewImportJobRepository(pool *pgxpool.Pool) ImportJobRepository {
	return &pgxImportJobRepository{pool: pool}
}

func scanImportJob(row pgx.Row) (*models.ImportJob, error) {
	var j models.ImportJob
	err := row.Scan(
		&j.ID, &j.UserID, &j.Status, &j.SourceType,
		&j.FileName, &j.RowsTotal, &j.RowsImported,
		&j.CreatedAt, &j.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *pgxImportJobRepository) Create(ctx context.Context, userID uuid.UUID, fileName string) (*models.ImportJob, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO import_jobs (user_id, file_name)
		 VALUES ($1, $2)
		 RETURNING id, user_id, status, source_type, file_name, rows_total, rows_imported, created_at, completed_at`,
		userID, fileName,
	)
	return scanImportJob(row)
}

func (r *pgxImportJobRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.ImportJob, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, status, source_type, file_name, rows_total, rows_imported, created_at, completed_at
		 FROM import_jobs
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	return scanImportJob(row)
}

func (r *pgxImportJobRepository) List(ctx context.Context, userID uuid.UUID) ([]models.ImportJob, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, status, source_type, file_name, rows_total, rows_imported, created_at, completed_at
		 FROM import_jobs
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.ImportJob
	for rows.Next() {
		var j models.ImportJob
		if err := rows.Scan(
			&j.ID, &j.UserID, &j.Status, &j.SourceType,
			&j.FileName, &j.RowsTotal, &j.RowsImported,
			&j.CreatedAt, &j.CompletedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (r *pgxImportJobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE import_jobs SET status = $2 WHERE id = $1`,
		id, status,
	)
	return err
}

func (r *pgxImportJobRepository) UpdateRowsTotal(ctx context.Context, id uuid.UUID, rowsTotal int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE import_jobs SET rows_total = $2 WHERE id = $1`,
		id, rowsTotal,
	)
	return err
}

func (r *pgxImportJobRepository) IncrementRowsImported(ctx context.Context, id uuid.UUID, count int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE import_jobs SET rows_imported = rows_imported + $2 WHERE id = $1`,
		id, count,
	)
	return err
}

func (r *pgxImportJobRepository) Complete(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE import_jobs SET status = $2, completed_at = $3 WHERE id = $1`,
		id, status, time.Now().UTC(),
	)
	return err
}
