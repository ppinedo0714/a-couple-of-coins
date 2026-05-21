package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/services/predictor"
)

// --- fakeTx implements pgx.Tx with no-op methods ---

type fakeTx struct {
	committed  bool
	rolledBack bool
}

func (f *fakeTx) Begin(ctx context.Context) (pgx.Tx, error)   { return f, nil }
func (f *fakeTx) Commit(ctx context.Context) error             { f.committed = true; return nil }
func (f *fakeTx) Rollback(ctx context.Context) error           { f.rolledBack = true; return nil }
func (f *fakeTx) CopyFrom(_ context.Context, _ pgx.Identifier, _ []string, _ pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (f *fakeTx) SendBatch(_ context.Context, _ *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeTx) LargeObjects() pgx.LargeObjects                              { return pgx.LargeObjects{} }
func (f *fakeTx) Prepare(_ context.Context, _, _ string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (f *fakeTx) Exec(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (f *fakeTx) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) { return nil, nil }
func (f *fakeTx) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row        { return nil }
func (f *fakeTx) Conn() *pgx.Conn                                               { return nil }

// --- mock repositories ---

type mockTransactionRepository struct {
	getByIDFn          func(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error)
	createFn           func(ctx context.Context, tx pgx.Tx, t repository.CreateTransactionParams) (*models.Transaction, error)
	updateFn           func(ctx context.Context, tx pgx.Tx, id, userID uuid.UUID, fields repository.UpdateTransactionFields) (*models.Transaction, error)
	deleteFn           func(ctx context.Context, tx pgx.Tx, id, userID uuid.UUID) (*models.Transaction, error)
	getUnclassifiedFn  func(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error)
	setClassifiedFn    func(ctx context.Context, tx pgx.Tx, id uuid.UUID, categoryID *uuid.UUID, merchantName string) error
	sumByDateFn        func(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, date time.Time) (float64, error)
	deleteSnapshotFn   func(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, date time.Time) error
	sumFromDateFn      func(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, fromDate time.Time) (float64, error)

	setClassifiedCalls []setClassifiedCall
}

type setClassifiedCall struct {
	id           uuid.UUID
	categoryID   *uuid.UUID
	merchantName string
}

func (m *mockTransactionRepository) List(_ context.Context, _ uuid.UUID, _ repository.TransactionFilters) ([]models.Transaction, int, error) {
	return nil, 0, nil
}
func (m *mockTransactionRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Transaction, error) {
	return m.getByIDFn(ctx, id, userID)
}
func (m *mockTransactionRepository) Create(ctx context.Context, tx pgx.Tx, t repository.CreateTransactionParams) (*models.Transaction, error) {
	return m.createFn(ctx, tx, t)
}
func (m *mockTransactionRepository) Update(ctx context.Context, tx pgx.Tx, id, userID uuid.UUID, fields repository.UpdateTransactionFields) (*models.Transaction, error) {
	return m.updateFn(ctx, tx, id, userID, fields)
}
func (m *mockTransactionRepository) Delete(ctx context.Context, tx pgx.Tx, id, userID uuid.UUID) (*models.Transaction, error) {
	return m.deleteFn(ctx, tx, id, userID)
}
func (m *mockTransactionRepository) GetUnclassified(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	return m.getUnclassifiedFn(ctx, userID)
}
func (m *mockTransactionRepository) SetClassified(ctx context.Context, tx pgx.Tx, id uuid.UUID, categoryID *uuid.UUID, merchantName string) error {
	m.setClassifiedCalls = append(m.setClassifiedCalls, setClassifiedCall{id, categoryID, merchantName})
	if m.setClassifiedFn != nil {
		return m.setClassifiedFn(ctx, tx, id, categoryID, merchantName)
	}
	return nil
}
func (m *mockTransactionRepository) SumByAccountAndDate(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, date time.Time) (float64, error) {
	if m.sumByDateFn != nil {
		return m.sumByDateFn(ctx, tx, accountID, date)
	}
	return 0, nil
}
func (m *mockTransactionRepository) DeleteBalanceSnapshot(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, date time.Time) error {
	if m.deleteSnapshotFn != nil {
		return m.deleteSnapshotFn(ctx, tx, accountID, date)
	}
	return nil
}
func (m *mockTransactionRepository) SumByAccountFromDate(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, fromDate time.Time) (float64, error) {
	if m.sumFromDateFn != nil {
		return m.sumFromDateFn(ctx, tx, accountID, fromDate)
	}
	return 0, nil
}

type mockAccountRepository struct {
	getByIDFn           func(ctx context.Context, id, userID uuid.UUID) (*models.Account, error)
	updateBalanceCalls  []updateBalanceCall
	upsertSnapshotCalls []upsertSnapshotCall
}

type updateBalanceCall struct {
	id    uuid.UUID
	delta float64
}
type upsertSnapshotCall struct {
	accountID uuid.UUID
	date      time.Time
	balance   float64
}

func (m *mockAccountRepository) List(_ context.Context, _ uuid.UUID) ([]models.Account, error) {
	return nil, nil
}
func (m *mockAccountRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Account, error) {
	return m.getByIDFn(ctx, id, userID)
}
func (m *mockAccountRepository) Create(_ context.Context, _ uuid.UUID, _, _, _ string, _ float64) (*models.Account, error) {
	return nil, nil
}
func (m *mockAccountRepository) Update(_ context.Context, _, _ uuid.UUID, _ repository.AccountUpdateFields) (*models.Account, error) {
	return nil, nil
}
func (m *mockAccountRepository) Delete(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockAccountRepository) UpdateBalance(_ context.Context, _ pgx.Tx, id uuid.UUID, delta float64) error {
	m.updateBalanceCalls = append(m.updateBalanceCalls, updateBalanceCall{id, delta})
	return nil
}
func (m *mockAccountRepository) ListBalanceSnapshots(_ context.Context, _ []uuid.UUID, _, _ time.Time, _ string) ([]models.BalanceSnapshot, error) {
	return nil, nil
}
func (m *mockAccountRepository) UpsertBalanceSnapshot(_ context.Context, _ pgx.Tx, accountID uuid.UUID, date time.Time, balance float64) error {
	m.upsertSnapshotCalls = append(m.upsertSnapshotCalls, upsertSnapshotCall{accountID, date, balance})
	return nil
}

type mockCategoryRepository struct{}

func (m *mockCategoryRepository) List(_ context.Context, _ uuid.UUID) ([]models.Category, error) {
	return []models.Category{}, nil
}
func (m *mockCategoryRepository) GetByID(_ context.Context, _, _ uuid.UUID) (*models.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepository) Create(_ context.Context, _ uuid.UUID, _ string, _ *uuid.UUID, _ *string) (*models.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepository) Update(_ context.Context, _, _ uuid.UUID, _ repository.CategoryUpdateFields) (*models.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepository) Delete(_ context.Context, _, _ uuid.UUID) error { return nil }

type mockPredictorClient struct {
	classifyFn func(ctx context.Context, transactions []models.Transaction, categories []models.Category) ([]predictor.Prediction, error)
}

func (m *mockPredictorClient) Classify(ctx context.Context, transactions []models.Transaction, categories []models.Category) ([]predictor.Prediction, error) {
	return m.classifyFn(ctx, transactions, categories)
}

// --- testable service that accepts a beginTx function ---

// testableTransactionService mirrors transactionService but uses injectable beginTx.
type testableTransactionService struct {
	txRepo       repository.TransactionRepository
	accountRepo  repository.AccountRepository
	categoryRepo repository.CategoryRepository
	predictor    predictor.PredictorClient
	beginTx      func(ctx context.Context) (pgx.Tx, error)
}

func newTestSvc(
	txRepo repository.TransactionRepository,
	accountRepo repository.AccountRepository,
	categoryRepo repository.CategoryRepository,
	pred predictor.PredictorClient,
	beginTx func(ctx context.Context) (pgx.Tx, error),
) *testableTransactionService {
	return &testableTransactionService{txRepo, accountRepo, categoryRepo, pred, beginTx}
}

func (s *testableTransactionService) Create(ctx context.Context, userID uuid.UUID, req models.CreateTransactionRequest) (*models.Transaction, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, err
	}
	account, err := s.accountRepo.GetByID(ctx, req.AccountID, userID)
	if err != nil {
		return nil, err
	}
	newBalance := account.Balance + req.Amount

	tx, err := s.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	created, err := s.txRepo.Create(ctx, tx, repository.CreateTransactionParams{
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
	if err := s.accountRepo.UpdateBalance(ctx, tx, req.AccountID, req.Amount); err != nil {
		return nil, err
	}
	if err := s.accountRepo.UpsertBalanceSnapshot(ctx, tx, req.AccountID, date, newBalance); err != nil {
		return nil, err
	}
	return created, tx.Commit(ctx)
}

func (s *testableTransactionService) Update(ctx context.Context, id, userID uuid.UUID, req models.UpdateTransactionRequest) (*models.Transaction, error) {
	old, err := s.txRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	var newDate *time.Time
	if req.Date != nil {
		parsed, parseErr := time.Parse("2006-01-02", *req.Date)
		if parseErr != nil {
			return nil, parseErr
		}
		newDate = &parsed
	}

	amountChanged := req.Amount != nil && *req.Amount != old.Amount
	dateChanged := newDate != nil && !newDate.Equal(old.Date)
	financialChange := amountChanged || dateChanged

	dbTx, err := s.beginTx(ctx)
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
		return updated, dbTx.Commit(ctx)
	}

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

	newAccountBalance := account.Balance + delta

	if !dateChanged {
		if err := s.accountRepo.UpsertBalanceSnapshot(ctx, dbTx, old.AccountID, old.Date, newAccountBalance); err != nil {
			return nil, err
		}
	} else {
		oldDateSum, err := s.txRepo.SumByAccountAndDate(ctx, dbTx, old.AccountID, old.Date)
		if err != nil {
			return nil, err
		}
		if oldDateSum != 0 {
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
		if err := s.accountRepo.UpsertBalanceSnapshot(ctx, dbTx, old.AccountID, *newDate, newAccountBalance); err != nil {
			return nil, err
		}
	}

	return updated, dbTx.Commit(ctx)
}

func (s *testableTransactionService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	old, err := s.txRepo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}
	account, err := s.accountRepo.GetByID(ctx, old.AccountID, userID)
	if err != nil {
		return err
	}

	dbTx, err := s.beginTx(ctx)
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

func (s *testableTransactionService) ClassifyUnclassified(ctx context.Context, userID uuid.UUID) (models.ClassifyResult, error) {
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

	predByID := make(map[uuid.UUID]predictor.Prediction, len(predictions))
	for _, p := range predictions {
		predByID[p.TransactionID] = p
	}

	var classified, failed int
	for _, t := range txns {
		p, ok := predByID[t.ID]
		if !ok {
			failed++
			continue
		}
		dbTx, err := s.beginTx(ctx)
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

	return models.ClassifyResult{Classified: classified, Failed: failed}, nil
}

// --- helpers ---

func mustParseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func makeBeginFn(tx pgx.Tx) func(ctx context.Context) (pgx.Tx, error) {
	return func(ctx context.Context) (pgx.Tx, error) { return tx, nil }
}

// --- tests ---

func TestCreate_UpdatesBalanceAndUpsertSnapshot(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	txID := uuid.New()
	date := mustParseDate("2024-01-15")
	initialBalance := 100.0
	amount := -42.50

	accountRepo := &mockAccountRepository{
		getByIDFn: func(_ context.Context, id, _ uuid.UUID) (*models.Account, error) {
			return &models.Account{ID: accountID, UserID: userID, Balance: initialBalance}, nil
		},
	}
	txRepo := &mockTransactionRepository{
		createFn: func(_ context.Context, _ pgx.Tx, p repository.CreateTransactionParams) (*models.Transaction, error) {
			return &models.Transaction{ID: txID, Amount: p.Amount, Date: p.Date, Source: p.Source, Classified: p.Classified}, nil
		},
	}

	ft := &fakeTx{}
	svc := newTestSvc(txRepo, accountRepo, &mockCategoryRepository{}, &mockPredictorClient{}, makeBeginFn(ft))

	created, err := svc.Create(context.Background(), userID, models.CreateTransactionRequest{
		AccountID: accountID, Amount: amount, Description: "Whole Foods", Date: "2024-01-15",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != txID {
		t.Errorf("want ID %v, got %v", txID, created.ID)
	}
	if created.Source != "manual" {
		t.Errorf("want source manual, got %v", created.Source)
	}
	if !created.Classified {
		t.Error("want classified=true")
	}

	if len(accountRepo.updateBalanceCalls) != 1 || accountRepo.updateBalanceCalls[0].delta != amount {
		t.Errorf("want UpdateBalance delta=%v, got %+v", amount, accountRepo.updateBalanceCalls)
	}

	if len(accountRepo.upsertSnapshotCalls) != 1 {
		t.Fatalf("want 1 UpsertSnapshot, got %d", len(accountRepo.upsertSnapshotCalls))
	}
	want := initialBalance + amount
	if accountRepo.upsertSnapshotCalls[0].balance != want {
		t.Errorf("want snapshot balance=%v, got %v", want, accountRepo.upsertSnapshotCalls[0].balance)
	}
	if !accountRepo.upsertSnapshotCalls[0].date.Equal(date) {
		t.Errorf("want snapshot date=%v, got %v", date, accountRepo.upsertSnapshotCalls[0].date)
	}
	if !ft.committed {
		t.Error("want tx committed")
	}
}

func TestUpdate_AmountOnly_DeltaAndSnapshot(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	txID := uuid.New()
	date := mustParseDate("2024-01-15")
	oldAmount := -42.50
	newAmount := -45.00
	accountBalance := 100.0

	accountRepo := &mockAccountRepository{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.Account, error) {
			return &models.Account{ID: accountID, Balance: accountBalance}, nil
		},
	}
	txRepo := &mockTransactionRepository{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.Transaction, error) {
			return &models.Transaction{ID: txID, AccountID: accountID, Amount: oldAmount, Date: date}, nil
		},
		updateFn: func(_ context.Context, _ pgx.Tx, _, _ uuid.UUID, _ repository.UpdateTransactionFields) (*models.Transaction, error) {
			return &models.Transaction{ID: txID, Amount: newAmount, Date: date}, nil
		},
	}

	ft := &fakeTx{}
	svc := newTestSvc(txRepo, accountRepo, &mockCategoryRepository{}, &mockPredictorClient{}, makeBeginFn(ft))

	_, err := svc.Update(context.Background(), txID, userID, models.UpdateTransactionRequest{Amount: &newAmount})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDelta := newAmount - oldAmount
	if len(accountRepo.updateBalanceCalls) != 1 || accountRepo.updateBalanceCalls[0].delta != expectedDelta {
		t.Errorf("want delta=%v, got %+v", expectedDelta, accountRepo.updateBalanceCalls)
	}

	if len(accountRepo.upsertSnapshotCalls) != 1 {
		t.Fatalf("want 1 UpsertSnapshot, got %d", len(accountRepo.upsertSnapshotCalls))
	}
	want := accountBalance + expectedDelta
	if accountRepo.upsertSnapshotCalls[0].balance != want {
		t.Errorf("want snapshot balance=%v, got %v", want, accountRepo.upsertSnapshotCalls[0].balance)
	}
	if !accountRepo.upsertSnapshotCalls[0].date.Equal(date) {
		t.Errorf("want snapshot date=%v, got %v", date, accountRepo.upsertSnapshotCalls[0].date)
	}
}

func TestUpdate_DateOnly_OldSnapshotDeleted_NewSnapshotSet(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	txID := uuid.New()
	oldDate := mustParseDate("2024-01-15")
	newDate := mustParseDate("2024-01-20")
	amount := -42.50
	accountBalance := 100.0
	newDateStr := "2024-01-20"

	accountRepo := &mockAccountRepository{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.Account, error) {
			return &models.Account{ID: accountID, Balance: accountBalance}, nil
		},
	}

	deleteSnapshotDates := []time.Time{}
	txRepo := &mockTransactionRepository{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.Transaction, error) {
			return &models.Transaction{ID: txID, AccountID: accountID, Amount: amount, Date: oldDate}, nil
		},
		updateFn: func(_ context.Context, _ pgx.Tx, _, _ uuid.UUID, _ repository.UpdateTransactionFields) (*models.Transaction, error) {
			return &models.Transaction{ID: txID, Amount: amount, Date: newDate}, nil
		},
		sumByDateFn: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ time.Time) (float64, error) {
			return 0, nil // no remaining transactions on old date
		},
		deleteSnapshotFn: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, d time.Time) error {
			deleteSnapshotDates = append(deleteSnapshotDates, d)
			return nil
		},
	}

	ft := &fakeTx{}
	svc := newTestSvc(txRepo, accountRepo, &mockCategoryRepository{}, &mockPredictorClient{}, makeBeginFn(ft))

	_, err := svc.Update(context.Background(), txID, userID, models.UpdateTransactionRequest{Date: &newDateStr})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deleteSnapshotDates) != 1 || !deleteSnapshotDates[0].Equal(oldDate) {
		t.Errorf("want DeleteBalanceSnapshot called with %v, got %v", oldDate, deleteSnapshotDates)
	}

	if len(accountRepo.upsertSnapshotCalls) != 1 {
		t.Fatalf("want 1 UpsertSnapshot call (new date), got %d", len(accountRepo.upsertSnapshotCalls))
	}
	if !accountRepo.upsertSnapshotCalls[0].date.Equal(newDate) {
		t.Errorf("want upsert for %v, got %v", newDate, accountRepo.upsertSnapshotCalls[0].date)
	}
	// delta=0, so balance unchanged
	if accountRepo.upsertSnapshotCalls[0].balance != accountBalance {
		t.Errorf("want balance=%v, got %v", accountBalance, accountRepo.upsertSnapshotCalls[0].balance)
	}
	// No UpdateBalance call (delta=0)
	if len(accountRepo.updateBalanceCalls) != 0 {
		t.Errorf("want no UpdateBalance calls, got %d", len(accountRepo.updateBalanceCalls))
	}
}

func TestUpdate_AmountAndDate_BothEffectsApplied(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	txID := uuid.New()
	oldDate := mustParseDate("2024-01-15")
	newDate := mustParseDate("2024-01-20")
	oldAmount := -42.50
	newAmount := -55.00
	accountBalance := 200.0
	newDateStr := "2024-01-20"

	accountRepo := &mockAccountRepository{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.Account, error) {
			return &models.Account{ID: accountID, Balance: accountBalance}, nil
		},
	}

	deleteSnapshotCalled := false
	txRepo := &mockTransactionRepository{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.Transaction, error) {
			return &models.Transaction{ID: txID, AccountID: accountID, Amount: oldAmount, Date: oldDate}, nil
		},
		updateFn: func(_ context.Context, _ pgx.Tx, _, _ uuid.UUID, _ repository.UpdateTransactionFields) (*models.Transaction, error) {
			return &models.Transaction{ID: txID, Amount: newAmount, Date: newDate}, nil
		},
		sumByDateFn: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ time.Time) (float64, error) {
			return 0, nil
		},
		deleteSnapshotFn: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ time.Time) error {
			deleteSnapshotCalled = true
			return nil
		},
	}

	ft := &fakeTx{}
	svc := newTestSvc(txRepo, accountRepo, &mockCategoryRepository{}, &mockPredictorClient{}, makeBeginFn(ft))

	_, err := svc.Update(context.Background(), txID, userID, models.UpdateTransactionRequest{Amount: &newAmount, Date: &newDateStr})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDelta := newAmount - oldAmount
	if len(accountRepo.updateBalanceCalls) != 1 || accountRepo.updateBalanceCalls[0].delta != expectedDelta {
		t.Errorf("want delta=%v, got %+v", expectedDelta, accountRepo.updateBalanceCalls)
	}
	if !deleteSnapshotCalled {
		t.Error("want old-date snapshot deleted")
	}
	if len(accountRepo.upsertSnapshotCalls) != 1 {
		t.Fatalf("want 1 UpsertSnapshot (new date), got %d", len(accountRepo.upsertSnapshotCalls))
	}
	if !accountRepo.upsertSnapshotCalls[0].date.Equal(newDate) {
		t.Errorf("want upsert for %v, got %v", newDate, accountRepo.upsertSnapshotCalls[0].date)
	}
}

func TestDelete_BalanceReversed_SnapshotDeletedWhenNoRemaining(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	txID := uuid.New()
	date := mustParseDate("2024-01-15")
	amount := -42.50
	accountBalance := 57.50

	oldTxn := &models.Transaction{ID: txID, AccountID: accountID, Amount: amount, Date: date}

	accountRepo := &mockAccountRepository{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.Account, error) {
			return &models.Account{ID: accountID, Balance: accountBalance}, nil
		},
	}

	deleteSnapshotCalled := false
	txRepo := &mockTransactionRepository{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.Transaction, error) {
			return oldTxn, nil
		},
		deleteFn: func(_ context.Context, _ pgx.Tx, _, _ uuid.UUID) (*models.Transaction, error) {
			return oldTxn, nil
		},
		sumByDateFn: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ time.Time) (float64, error) {
			return 0, nil
		},
		deleteSnapshotFn: func(_ context.Context, _ pgx.Tx, _ uuid.UUID, d time.Time) error {
			deleteSnapshotCalled = true
			if !d.Equal(date) {
				t.Errorf("want delete for %v, got %v", date, d)
			}
			return nil
		},
	}

	ft := &fakeTx{}
	svc := newTestSvc(txRepo, accountRepo, &mockCategoryRepository{}, &mockPredictorClient{}, makeBeginFn(ft))

	err := svc.Delete(context.Background(), txID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(accountRepo.updateBalanceCalls) != 1 || accountRepo.updateBalanceCalls[0].delta != -amount {
		t.Errorf("want UpdateBalance delta=%v, got %+v", -amount, accountRepo.updateBalanceCalls)
	}
	if !deleteSnapshotCalled {
		t.Error("want snapshot deleted when no remaining transactions")
	}
	if len(accountRepo.upsertSnapshotCalls) != 0 {
		t.Errorf("want no UpsertSnapshot calls, got %d", len(accountRepo.upsertSnapshotCalls))
	}
}

func TestClassifyUnclassified_AllPredicted(t *testing.T) {
	userID := uuid.New()
	tx1 := models.Transaction{ID: uuid.New(), UserID: userID}
	tx2 := models.Transaction{ID: uuid.New(), UserID: userID}
	catID := uuid.New()

	txRepo := &mockTransactionRepository{
		getUnclassifiedFn: func(_ context.Context, _ uuid.UUID) ([]models.Transaction, error) {
			return []models.Transaction{tx1, tx2}, nil
		},
	}

	ft := &fakeTx{}
	pred := &mockPredictorClient{
		classifyFn: func(_ context.Context, txns []models.Transaction, _ []models.Category) ([]predictor.Prediction, error) {
			return []predictor.Prediction{
				{TransactionID: tx1.ID, CategoryID: &catID, MerchantName: "Whole Foods"},
				{TransactionID: tx2.ID, CategoryID: &catID, MerchantName: "Amazon"},
			}, nil
		},
	}

	svc := newTestSvc(txRepo, &mockAccountRepository{}, &mockCategoryRepository{}, pred, makeBeginFn(ft))

	result, err := svc.ClassifyUnclassified(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Classified != 2 {
		t.Errorf("want 2 classified, got %d", result.Classified)
	}
	if result.Failed != 0 {
		t.Errorf("want 0 failed, got %d", result.Failed)
	}
	if len(txRepo.setClassifiedCalls) != 2 {
		t.Errorf("want 2 SetClassified calls, got %d", len(txRepo.setClassifiedCalls))
	}
}

func TestClassifyUnclassified_PartialPredictions(t *testing.T) {
	userID := uuid.New()
	tx1 := models.Transaction{ID: uuid.New(), UserID: userID}
	tx2 := models.Transaction{ID: uuid.New(), UserID: userID}
	tx3 := models.Transaction{ID: uuid.New(), UserID: userID}
	catID := uuid.New()

	txRepo := &mockTransactionRepository{
		getUnclassifiedFn: func(_ context.Context, _ uuid.UUID) ([]models.Transaction, error) {
			return []models.Transaction{tx1, tx2, tx3}, nil
		},
	}

	ft := &fakeTx{}
	pred := &mockPredictorClient{
		classifyFn: func(_ context.Context, _ []models.Transaction, _ []models.Category) ([]predictor.Prediction, error) {
			return []predictor.Prediction{
				{TransactionID: tx1.ID, CategoryID: &catID, MerchantName: "Whole Foods"},
				{TransactionID: tx3.ID, CategoryID: &catID, MerchantName: "Amazon"},
				// tx2 is absent — should count as failed
			}, nil
		},
	}

	svc := newTestSvc(txRepo, &mockAccountRepository{}, &mockCategoryRepository{}, pred, makeBeginFn(ft))

	result, err := svc.ClassifyUnclassified(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Classified != 2 {
		t.Errorf("want 2 classified, got %d", result.Classified)
	}
	if result.Failed != 1 {
		t.Errorf("want 1 failed, got %d", result.Failed)
	}
}

// Verify ErrNotFound is handled correctly.
func TestCreate_AccountNotFound(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()

	accountRepo := &mockAccountRepository{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*models.Account, error) {
			return nil, repository.ErrNotFound
		},
	}
	txRepo := &mockTransactionRepository{}
	ft := &fakeTx{}

	svc := newTestSvc(txRepo, accountRepo, &mockCategoryRepository{}, &mockPredictorClient{}, makeBeginFn(ft))

	_, err := svc.Create(context.Background(), userID, models.CreateTransactionRequest{
		AccountID: accountID, Amount: -10, Description: "test", Date: "2024-01-01",
	})
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}
