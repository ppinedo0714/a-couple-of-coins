package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/services/predictor"
)

type TransactionService interface {
	List(ctx context.Context, userID uuid.UUID, filters repository.TransactionFilters) ([]models.Transaction, int, error)
	Get(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error)
	Create(ctx context.Context, userID uuid.UUID, req models.CreateTransactionRequest) (*models.Transaction, error)
	Update(ctx context.Context, id, userID uuid.UUID, req models.UpdateTransactionRequest) (*models.Transaction, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	ClassifyUnclassified(ctx context.Context, userID uuid.UUID) (models.ClassifyResult, error)
}

type transactionService struct {
	pool         *pgxpool.Pool
	txRepo       repository.TransactionRepository
	accountRepo  repository.AccountRepository
	categoryRepo repository.CategoryRepository
	predictor    predictor.PredictorClient
}

func NewTransactionService(
	pool *pgxpool.Pool,
	txRepo repository.TransactionRepository,
	accountRepo repository.AccountRepository,
	categoryRepo repository.CategoryRepository,
	pred predictor.PredictorClient,
) TransactionService {
	return &transactionService{
		pool:         pool,
		txRepo:       txRepo,
		accountRepo:  accountRepo,
		categoryRepo: categoryRepo,
		predictor:    pred,
	}
}

func (s *transactionService) List(ctx context.Context, userID uuid.UUID, filters repository.TransactionFilters) ([]models.Transaction, int, error) {
	return s.txRepo.List(ctx, userID, filters)
}

func (s *transactionService) Get(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error) {
	return s.txRepo.GetByID(ctx, id, userID)
}

func (s *transactionService) Create(ctx context.Context, userID uuid.UUID, req models.CreateTransactionRequest) (*models.Transaction, error) {
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if req.AccountID == uuid.Nil {
		return nil, fmt.Errorf("account_id is required")
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date, use YYYY-MM-DD")
	}

	account, err := s.accountRepo.GetByID(ctx, req.AccountID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	newBalance := account.Balance + req.Amount

	dbTx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer dbTx.Rollback(ctx)

	created, err := s.txRepo.Create(ctx, dbTx, repository.CreateTransactionParams{
		UserID:      userID,
		AccountID:   req.AccountID,
		CategoryID:  req.CategoryID,
		Amount:      req.Amount,
		Description: req.Description,
		Date:        date,
		Source:      "manual",
		Classified:  true,
	})
	if err != nil {
		return nil, err
	}

	if err := s.accountRepo.UpdateBalance(ctx, dbTx, req.AccountID, req.Amount); err != nil {
		return nil, err
	}

	if err := s.accountRepo.UpsertBalanceSnapshot(ctx, dbTx, req.AccountID, date, newBalance); err != nil {
		return nil, err
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, err
	}

	return created, nil
}

func (s *transactionService) Update(ctx context.Context, id, userID uuid.UUID, req models.UpdateTransactionRequest) (*models.Transaction, error) {
	old, err := s.txRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	var newDate *time.Time
	if req.Date != nil {
		parsed, parseErr := time.Parse("2006-01-02", *req.Date)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid date, use YYYY-MM-DD")
		}
		newDate = &parsed
	}

	amountChanged := req.Amount != nil && *req.Amount != old.Amount
	dateChanged := newDate != nil && !newDate.Equal(old.Date)
	financialChange := amountChanged || dateChanged

	dbTx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer dbTx.Rollback(ctx)

	fields := repository.UpdateTransactionFields{
		CategoryID:    req.CategoryID,
		ClearCategory: req.ClearCategory,
		Description:   req.Description,
		Amount:        req.Amount,
		Date:          newDate,
	}

	updated, err := s.txRepo.Update(ctx, dbTx, id, userID, fields)
	if err != nil {
		return nil, err
	}

	if !financialChange {
		// Case 1: only non-financial fields changed
		if err := dbTx.Commit(ctx); err != nil {
			return nil, err
		}
		return updated, nil
	}

	// For cases 2, 3, 4 we need the account balance (pre-update, pre-delta).
	account, err := s.accountRepo.GetByID(ctx, old.AccountID, userID)
	if err != nil {
		return nil, err
	}

	effectiveNewAmount := old.Amount
	if req.Amount != nil {
		effectiveNewAmount = *req.Amount
	}
	delta := effectiveNewAmount - old.Amount

	if delta != 0 {
		if err := s.accountRepo.UpdateBalance(ctx, dbTx, old.AccountID, delta); err != nil {
			return nil, err
		}
	}

	// Account balance after delta (used for snapshot calculations).
	newAccountBalance := account.Balance + delta

	if !dateChanged {
		// Case 2: amount changed, date unchanged.
		if err := s.accountRepo.UpsertBalanceSnapshot(ctx, dbTx, old.AccountID, old.Date, newAccountBalance); err != nil {
			return nil, err
		}
	} else {
		// Case 3 or 4: date changed.
		// The transaction row is already updated in DB (at the new date), so
		// SumByAccountAndDate on old.Date now reflects only the remaining transactions.
		oldDateSum, err := s.txRepo.SumByAccountAndDate(ctx, dbTx, old.AccountID, old.Date)
		if err != nil {
			return nil, err
		}

		if oldDateSum != 0 {
			// prior_day_balance = newAccountBalance - SUM(all txns on account with date >= old_date)
			sumFromOldDate, err := s.txRepo.SumByAccountFromDate(ctx, dbTx, old.AccountID, old.Date)
			if err != nil {
				return nil, err
			}
			priorDayBalance := newAccountBalance - sumFromOldDate
			if err := s.accountRepo.UpsertBalanceSnapshot(ctx, dbTx, old.AccountID, old.Date, priorDayBalance+oldDateSum); err != nil {
				return nil, err
			}
		} else {
			if err := s.txRepo.DeleteBalanceSnapshot(ctx, dbTx, old.AccountID, old.Date); err != nil {
				return nil, err
			}
		}

		// Apply new-date snapshot.
		if err := s.accountRepo.UpsertBalanceSnapshot(ctx, dbTx, old.AccountID, *newDate, newAccountBalance); err != nil {
			return nil, err
		}
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *transactionService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	old, err := s.txRepo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	account, err := s.accountRepo.GetByID(ctx, old.AccountID, userID)
	if err != nil {
		return err
	}

	dbTx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer dbTx.Rollback(ctx)

	if _, err := s.txRepo.Delete(ctx, dbTx, id, userID); err != nil {
		return err
	}

	if err := s.accountRepo.UpdateBalance(ctx, dbTx, old.AccountID, -old.Amount); err != nil {
		return err
	}

	// Account balance after reversal.
	newAccountBalance := account.Balance - old.Amount

	remaining, err := s.txRepo.SumByAccountAndDate(ctx, dbTx, old.AccountID, old.Date)
	if err != nil {
		return err
	}

	if remaining != 0 {
		sumFromDate, err := s.txRepo.SumByAccountFromDate(ctx, dbTx, old.AccountID, old.Date)
		if err != nil {
			return err
		}
		priorDayBalance := newAccountBalance - sumFromDate
		if err := s.accountRepo.UpsertBalanceSnapshot(ctx, dbTx, old.AccountID, old.Date, priorDayBalance+remaining); err != nil {
			return err
		}
	} else {
		if err := s.txRepo.DeleteBalanceSnapshot(ctx, dbTx, old.AccountID, old.Date); err != nil {
			return err
		}
	}

	return dbTx.Commit(ctx)
}

func (s *transactionService) ClassifyUnclassified(ctx context.Context, userID uuid.UUID) (models.ClassifyResult, error) {
	txns, err := s.txRepo.GetUnclassified(ctx, userID)
	if err != nil {
		return models.ClassifyResult{}, err
	}

	if len(txns) == 0 {
		return models.ClassifyResult{}, nil
	}

	cats, err := s.categoryRepo.List(ctx, userID)
	if err != nil {
		return models.ClassifyResult{}, err
	}

	predictions, err := s.predictor.Classify(ctx, txns, cats)
	if err != nil {
		return models.ClassifyResult{}, err
	}

	predictionByID := make(map[uuid.UUID]predictor.Prediction, len(predictions))
	for _, p := range predictions {
		predictionByID[p.TransactionID] = p
	}

	var classified, failed int
	for _, t := range txns {
		p, ok := predictionByID[t.ID]
		if !ok {
			failed++
			continue
		}

		dbTx, err := s.pool.Begin(ctx)
		if err != nil {
			failed++
			continue
		}

		if err := s.txRepo.SetClassified(ctx, dbTx, t.ID, p.CategoryID, p.MerchantName); err != nil {
			dbTx.Rollback(ctx)
			failed++
			continue
		}

		if err := dbTx.Commit(ctx); err != nil {
			failed++
			continue
		}

		classified++
	}

	return models.ClassifyResult{
		Classified: classified,
		Failed:     failed,
	}, nil
}
