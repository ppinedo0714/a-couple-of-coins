package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
)

type TransactionFilters struct {
	AccountID    *uuid.UUID
	CategoryID   *uuid.UUID
	From         *time.Time
	To           *time.Time
	Search       *string
	Unclassified *bool
	Limit        int
	Offset       int
}

type CreateTransactionParams struct {
	UserID      uuid.UUID
	AccountID   uuid.UUID
	CategoryID  *uuid.UUID
	Amount      float64
	Description string
	Date        time.Time
	Source      string
	Classified  bool
}

type UpdateTransactionFields struct {
	CategoryID   *uuid.UUID
	ClearCategory bool
	Description  *string
	Amount       *float64
	Date         *time.Time
}

type TransactionRepository interface {
	List(ctx context.Context, userID uuid.UUID, filters TransactionFilters) ([]models.Transaction, int, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error)
	Create(ctx context.Context, tx pgx.Tx, t CreateTransactionParams) (*models.Transaction, error)
	Update(ctx context.Context, tx pgx.Tx, id, userID uuid.UUID, fields UpdateTransactionFields) (*models.Transaction, error)
	Delete(ctx context.Context, tx pgx.Tx, id, userID uuid.UUID) (*models.Transaction, error)
	GetUnclassified(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error)
	SetClassified(ctx context.Context, tx pgx.Tx, id uuid.UUID, categoryID *uuid.UUID, merchantName string) error
	SumByAccountAndDate(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, date time.Time) (float64, error)
	DeleteBalanceSnapshot(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, date time.Time) error
	SumByAccountFromDate(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, fromDate time.Time) (float64, error)
}

type pgxTransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) TransactionRepository {
	return &pgxTransactionRepository{pool: pool}
}

func scanTransaction(row pgx.Row) (*models.Transaction, error) {
	var t models.Transaction
	err := row.Scan(
		&t.ID, &t.UserID, &t.AccountID, &t.CategoryID,
		&t.Amount, &t.Description, &t.MerchantName,
		&t.Date, &t.Source, &t.Classified, &t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *pgxTransactionRepository) List(ctx context.Context, userID uuid.UUID, filters TransactionFilters) ([]models.Transaction, int, error) {
	args := []interface{}{userID}
	where := "user_id = $1"
	argIdx := 2

	if filters.AccountID != nil {
		where += fmt.Sprintf(" AND account_id = $%d", argIdx)
		args = append(args, *filters.AccountID)
		argIdx++
	}
	if filters.CategoryID != nil {
		where += fmt.Sprintf(" AND category_id = $%d", argIdx)
		args = append(args, *filters.CategoryID)
		argIdx++
	}
	if filters.From != nil {
		where += fmt.Sprintf(" AND date >= $%d", argIdx)
		args = append(args, *filters.From)
		argIdx++
	}
	if filters.To != nil {
		where += fmt.Sprintf(" AND date <= $%d", argIdx)
		args = append(args, *filters.To)
		argIdx++
	}
	if filters.Search != nil {
		where += fmt.Sprintf(" AND (description ILIKE '%%' || $%d || '%%' OR merchant_name ILIKE '%%' || $%d || '%%')", argIdx, argIdx)
		args = append(args, *filters.Search)
		argIdx++
	}
	if filters.Unclassified != nil && *filters.Unclassified {
		where += " AND classified = false"
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := filters.Offset

	query := fmt.Sprintf(`
		SELECT id, user_id, account_id, category_id, amount, description, merchant_name,
		       date, source, classified, created_at,
		       COUNT(*) OVER() AS total_count
		FROM transactions
		WHERE %s
		ORDER BY date DESC, created_at DESC
		LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	var total int
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.AccountID, &t.CategoryID,
			&t.Amount, &t.Description, &t.MerchantName,
			&t.Date, &t.Source, &t.Classified, &t.CreatedAt,
			&total,
		); err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return transactions, total, nil
}

func (r *pgxTransactionRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, account_id, category_id, amount, description, merchant_name,
		        date, source, classified, created_at
		 FROM transactions
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	return scanTransaction(row)
}

func (r *pgxTransactionRepository) Create(ctx context.Context, tx pgx.Tx, t CreateTransactionParams) (*models.Transaction, error) {
	row := tx.QueryRow(ctx,
		`INSERT INTO transactions (user_id, account_id, category_id, amount, description, date, source, classified)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, user_id, account_id, category_id, amount, description, merchant_name,
		           date, source, classified, created_at`,
		t.UserID, t.AccountID, t.CategoryID, t.Amount, t.Description, t.Date, t.Source, t.Classified,
	)
	return scanTransaction(row)
}

func (r *pgxTransactionRepository) Update(ctx context.Context, tx pgx.Tx, id, userID uuid.UUID, fields UpdateTransactionFields) (*models.Transaction, error) {
	var row pgx.Row
	if fields.ClearCategory {
		row = tx.QueryRow(ctx,
			`UPDATE transactions
			 SET category_id  = NULL,
			     description  = COALESCE($3, description),
			     amount       = COALESCE($4, amount),
			     date         = COALESCE($5, date)
			 WHERE id = $1 AND user_id = $2
			 RETURNING id, user_id, account_id, category_id, amount, description, merchant_name,
			           date, source, classified, created_at`,
			id, userID, fields.Description, fields.Amount, fields.Date,
		)
	} else {
		row = tx.QueryRow(ctx,
			`UPDATE transactions
			 SET category_id  = COALESCE($3, category_id),
			     description  = COALESCE($4, description),
			     amount       = COALESCE($5, amount),
			     date         = COALESCE($6, date)
			 WHERE id = $1 AND user_id = $2
			 RETURNING id, user_id, account_id, category_id, amount, description, merchant_name,
			           date, source, classified, created_at`,
			id, userID, fields.CategoryID, fields.Description, fields.Amount, fields.Date,
		)
	}
	return scanTransaction(row)
}

func (r *pgxTransactionRepository) Delete(ctx context.Context, tx pgx.Tx, id, userID uuid.UUID) (*models.Transaction, error) {
	row := tx.QueryRow(ctx,
		`DELETE FROM transactions
		 WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, account_id, category_id, amount, description, merchant_name,
		           date, source, classified, created_at`,
		id, userID,
	)
	return scanTransaction(row)
}

func (r *pgxTransactionRepository) GetUnclassified(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, account_id, category_id, amount, description, merchant_name,
		        date, source, classified, created_at
		 FROM transactions
		 WHERE user_id = $1 AND classified = false
		 ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.AccountID, &t.CategoryID,
			&t.Amount, &t.Description, &t.MerchantName,
			&t.Date, &t.Source, &t.Classified, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}

func (r *pgxTransactionRepository) SetClassified(ctx context.Context, tx pgx.Tx, id uuid.UUID, categoryID *uuid.UUID, merchantName string) error {
	_, err := tx.Exec(ctx,
		`UPDATE transactions
		 SET category_id = $2, merchant_name = $3, classified = true
		 WHERE id = $1`,
		id, categoryID, merchantName,
	)
	return err
}

func (r *pgxTransactionRepository) SumByAccountAndDate(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, date time.Time) (float64, error) {
	var sum float64
	err := tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0)
		 FROM transactions
		 WHERE account_id = $1 AND date = $2`,
		accountID, date,
	).Scan(&sum)
	return sum, err
}

func (r *pgxTransactionRepository) DeleteBalanceSnapshot(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, date time.Time) error {
	_, err := tx.Exec(ctx,
		`DELETE FROM account_balance_snapshots WHERE account_id = $1 AND date = $2`,
		accountID, date,
	)
	return err
}

func (r *pgxTransactionRepository) SumByAccountFromDate(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, fromDate time.Time) (float64, error) {
	var sum float64
	err := tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0)
		 FROM transactions
		 WHERE account_id = $1 AND date >= $2`,
		accountID, fromDate,
	).Scan(&sum)
	return sum, err
}
