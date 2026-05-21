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

type AccountUpdateFields struct {
	Name     *string
	Type     *string
	Currency *string
	Balance  *float64
}

type AccountRepository interface {
	List(ctx context.Context, userID uuid.UUID) ([]models.Account, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Account, error)
	Create(ctx context.Context, userID uuid.UUID, name, typ, currency string, balance float64) (*models.Account, error)
	Update(ctx context.Context, id, userID uuid.UUID, fields AccountUpdateFields) (*models.Account, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	UpdateBalance(ctx context.Context, tx pgx.Tx, id uuid.UUID, delta float64) error
	UpdateBalanceDirect(ctx context.Context, id uuid.UUID, delta float64) error
	ListBalanceSnapshots(ctx context.Context, accountIDs []uuid.UUID, from, to time.Time, interval string) ([]models.BalanceSnapshot, error)
	UpsertBalanceSnapshot(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, date time.Time, balance float64) error
	UpsertBalanceSnapshotDirect(ctx context.Context, accountID uuid.UUID, date time.Time, balance float64) error
}

type pgxAccountRepository struct {
	pool *pgxpool.Pool
}

func NewAccountRepository(pool *pgxpool.Pool) AccountRepository {
	return &pgxAccountRepository{pool: pool}
}

func (r *pgxAccountRepository) List(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, type, balance, currency, created_at
		 FROM accounts
		 WHERE user_id = $1
		 ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (r *pgxAccountRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Account, error) {
	var a models.Account
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, type, balance, currency, created_at
		 FROM accounts
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *pgxAccountRepository) Create(ctx context.Context, userID uuid.UUID, name, typ, currency string, balance float64) (*models.Account, error) {
	var a models.Account
	err := r.pool.QueryRow(ctx,
		`INSERT INTO accounts (user_id, name, type, currency, balance)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, name, type, balance, currency, created_at`,
		userID, name, typ, currency, balance,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *pgxAccountRepository) Update(ctx context.Context, id, userID uuid.UUID, fields AccountUpdateFields) (*models.Account, error) {
	var a models.Account
	err := r.pool.QueryRow(ctx,
		`UPDATE accounts
		 SET
		   name     = COALESCE($3, name),
		   type     = COALESCE($4, type),
		   currency = COALESCE($5, currency),
		   balance  = COALESCE($6, balance)
		 WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, name, type, balance, currency, created_at`,
		id, userID, fields.Name, fields.Type, fields.Currency, fields.Balance,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *pgxAccountRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM transactions WHERE account_id = $1`,
		id,
	).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrHasTransactions
	}

	tag, err := r.pool.Exec(ctx,
		`DELETE FROM accounts WHERE id = $1 AND user_id = $2`,
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

func (r *pgxAccountRepository) UpdateBalance(ctx context.Context, tx pgx.Tx, id uuid.UUID, delta float64) error {
	_, err := tx.Exec(ctx,
		`UPDATE accounts SET balance = balance + $2 WHERE id = $1`,
		id, delta,
	)
	return err
}

func (r *pgxAccountRepository) ListBalanceSnapshots(ctx context.Context, accountIDs []uuid.UUID, from, to time.Time, interval string) ([]models.BalanceSnapshot, error) {
	var query string
	switch interval {
	case "week":
		query = `
			SELECT DISTINCT ON (account_id, iso_week)
			       account_id,
			       date,
			       balance
			FROM (
			    SELECT account_id,
			           date,
			           balance,
			           to_char(date, 'IYYY-IW') AS iso_week
			    FROM account_balance_snapshots
			    WHERE account_id = ANY($1)
			      AND date >= $2
			      AND date <= $3
			) sub
			ORDER BY account_id, iso_week, date DESC`
	case "month":
		query = `
			SELECT DISTINCT ON (account_id, month)
			       account_id,
			       date,
			       balance
			FROM (
			    SELECT account_id,
			           date,
			           balance,
			           to_char(date, 'YYYY-MM') AS month
			    FROM account_balance_snapshots
			    WHERE account_id = ANY($1)
			      AND date >= $2
			      AND date <= $3
			) sub
			ORDER BY account_id, month, date DESC`
	default:
		query = `
			SELECT account_id, date, balance
			FROM account_balance_snapshots
			WHERE account_id = ANY($1)
			  AND date >= $2
			  AND date <= $3
			ORDER BY account_id, date ASC`
	}

	rows, err := r.pool.Query(ctx, query, accountIDs, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.BalanceSnapshot
	for rows.Next() {
		var s models.BalanceSnapshot
		if err := rows.Scan(&s.AccountID, &s.Date, &s.Balance); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

func (r *pgxAccountRepository) UpsertBalanceSnapshot(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, date time.Time, balance float64) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO account_balance_snapshots (account_id, date, balance)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (account_id, date) DO UPDATE SET balance = EXCLUDED.balance`,
		accountID, date, balance,
	)
	return err
}

func (r *pgxAccountRepository) UpdateBalanceDirect(ctx context.Context, id uuid.UUID, delta float64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE accounts SET balance = balance + $2 WHERE id = $1`,
		id, delta,
	)
	return err
}

func (r *pgxAccountRepository) UpsertBalanceSnapshotDirect(ctx context.Context, accountID uuid.UUID, date time.Time, balance float64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO account_balance_snapshots (account_id, date, balance)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (account_id, date) DO UPDATE SET balance = EXCLUDED.balance`,
		accountID, date, balance,
	)
	return err
}
