package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
)

type mockAccountRepo struct {
	listFn                  func(ctx context.Context, userID uuid.UUID) ([]models.Account, error)
	getByIDFn               func(ctx context.Context, id, userID uuid.UUID) (*models.Account, error)
	createFn                func(ctx context.Context, userID uuid.UUID, name, typ, currency string, balance float64) (*models.Account, error)
	updateFn                func(ctx context.Context, id, userID uuid.UUID, fields repository.AccountUpdateFields) (*models.Account, error)
	deleteFn                func(ctx context.Context, id, userID uuid.UUID) error
	updateBalanceFn         func(ctx context.Context, tx pgx.Tx, id uuid.UUID, delta float64) error
	listBalanceSnapshotsFn  func(ctx context.Context, accountIDs []uuid.UUID, from, to time.Time, interval string) ([]models.BalanceSnapshot, error)
	upsertBalanceSnapshotFn func(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, date time.Time, balance float64) error
}

func (m *mockAccountRepo) List(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	return m.listFn(ctx, userID)
}
func (m *mockAccountRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*models.Account, error) {
	return m.getByIDFn(ctx, id, userID)
}
func (m *mockAccountRepo) Create(ctx context.Context, userID uuid.UUID, name, typ, currency string, balance float64) (*models.Account, error) {
	return m.createFn(ctx, userID, name, typ, currency, balance)
}
func (m *mockAccountRepo) Update(ctx context.Context, id, userID uuid.UUID, fields repository.AccountUpdateFields) (*models.Account, error) {
	return m.updateFn(ctx, id, userID, fields)
}
func (m *mockAccountRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.deleteFn(ctx, id, userID)
}
func (m *mockAccountRepo) UpdateBalance(ctx context.Context, tx pgx.Tx, id uuid.UUID, delta float64) error {
	return m.updateBalanceFn(ctx, tx, id, delta)
}
func (m *mockAccountRepo) ListBalanceSnapshots(ctx context.Context, accountIDs []uuid.UUID, from, to time.Time, interval string) ([]models.BalanceSnapshot, error) {
	return m.listBalanceSnapshotsFn(ctx, accountIDs, from, to, interval)
}
func (m *mockAccountRepo) UpsertBalanceSnapshot(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, date time.Time, balance float64) error {
	return m.upsertBalanceSnapshotFn(ctx, tx, accountID, date, balance)
}

func (m *mockAccountRepo) UpdateBalanceDirect(_ context.Context, _ uuid.UUID, _ float64) error {
	return nil
}

func (m *mockAccountRepo) UpsertBalanceSnapshotDirect(_ context.Context, _ uuid.UUID, _ time.Time, _ float64) error {
	return nil
}

func makeAccount(userID uuid.UUID) *models.Account {
	return &models.Account{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      "Test Checking",
		Type:      "checking",
		Balance:   100.00,
		Currency:  "USD",
		CreatedAt: time.Now(),
	}
}

func TestAccountServiceList(t *testing.T) {
	userID := uuid.New()
	account := makeAccount(userID)

	repo := &mockAccountRepo{
		listFn: func(ctx context.Context, uid uuid.UUID) ([]models.Account, error) {
			if uid != userID {
				t.Errorf("expected userID %v, got %v", userID, uid)
			}
			return []models.Account{*account}, nil
		},
	}

	svc := NewAccountService(repo)
	accounts, err := svc.List(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 1 {
		t.Errorf("expected 1 account, got %d", len(accounts))
	}
}

func TestAccountServiceGetNotFound(t *testing.T) {
	repo := &mockAccountRepo{
		getByIDFn: func(ctx context.Context, id, userID uuid.UUID) (*models.Account, error) {
			return nil, repository.ErrNotFound
		},
	}

	svc := NewAccountService(repo)
	_, err := svc.Get(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAccountServiceCreateInvalidType(t *testing.T) {
	repo := &mockAccountRepo{}
	svc := NewAccountService(repo)

	req := CreateAccountRequest{
		Name:    "Test",
		Type:    "invalid-type",
		Balance: 0,
	}
	_, err := svc.Create(context.Background(), uuid.New(), req)
	if err == nil {
		t.Error("expected error for invalid type, got nil")
	}
}

func TestAccountServiceCreateValidType(t *testing.T) {
	userID := uuid.New()
	account := makeAccount(userID)

	repo := &mockAccountRepo{
		createFn: func(ctx context.Context, uid uuid.UUID, name, typ, currency string, balance float64) (*models.Account, error) {
			return account, nil
		},
	}

	svc := NewAccountService(repo)
	for _, validType := range []string{"checking", "savings", "credit", "investment"} {
		req := CreateAccountRequest{Name: "Test", Type: validType}
		_, err := svc.Create(context.Background(), userID, req)
		if err != nil {
			t.Errorf("unexpected error for type %q: %v", validType, err)
		}
	}
}

func TestAccountServiceDeletePropagatesErrHasTransactions(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	account := &models.Account{ID: accountID, UserID: userID}

	repo := &mockAccountRepo{
		getByIDFn: func(ctx context.Context, id, uid uuid.UUID) (*models.Account, error) {
			return account, nil
		},
		deleteFn: func(ctx context.Context, id, uid uuid.UUID) error {
			return repository.ErrHasTransactions
		},
	}

	svc := NewAccountService(repo)
	err := svc.Delete(context.Background(), accountID, userID)
	if !errors.Is(err, repository.ErrHasTransactions) {
		t.Errorf("expected ErrHasTransactions, got %v", err)
	}
}

func TestAccountServiceGetHistoryMissingDates(t *testing.T) {
	repo := &mockAccountRepo{}
	svc := NewAccountService(repo)

	req := HistoryRequest{
		// From and To are zero values
	}
	_, err := svc.GetHistory(context.Background(), uuid.New(), req)
	if err == nil {
		t.Error("expected error when from/to are missing, got nil")
	}
}

func TestAccountServiceGetHistoryFiltersAccountByUser(t *testing.T) {
	userID := uuid.New()
	otherAccountID := uuid.New()

	repo := &mockAccountRepo{
		getByIDFn: func(ctx context.Context, id, uid uuid.UUID) (*models.Account, error) {
			// Simulate the account not belonging to this user
			return nil, repository.ErrNotFound
		},
	}

	svc := NewAccountService(repo)
	req := HistoryRequest{
		AccountIDs: []uuid.UUID{otherAccountID},
		From:       time.Now().Add(-24 * time.Hour),
		To:         time.Now(),
	}
	_, err := svc.GetHistory(context.Background(), userID, req)
	if err == nil {
		t.Error("expected error for account not owned by user, got nil")
	}
}

func TestAccountServiceGetHistorySuccess(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	now := time.Now()

	repo := &mockAccountRepo{
		getByIDFn: func(ctx context.Context, id, uid uuid.UUID) (*models.Account, error) {
			return &models.Account{ID: accountID, UserID: userID}, nil
		},
		listBalanceSnapshotsFn: func(ctx context.Context, accountIDs []uuid.UUID, from, to time.Time, interval string) ([]models.BalanceSnapshot, error) {
			return []models.BalanceSnapshot{
				{AccountID: accountID, Date: now, Balance: 500.00},
			}, nil
		},
	}

	svc := NewAccountService(repo)
	req := HistoryRequest{
		AccountIDs: []uuid.UUID{accountID},
		From:       now.Add(-24 * time.Hour),
		To:         now,
		Interval:   "day",
	}
	snapshots, err := svc.GetHistory(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snapshots) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(snapshots))
	}
}
