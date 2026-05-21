package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
)

var validAccountTypes = map[string]bool{
	"checking":   true,
	"savings":    true,
	"credit":     true,
	"investment": true,
}

type CreateAccountRequest struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

type UpdateAccountRequest struct {
	Name     *string  `json:"name"`
	Type     *string  `json:"type"`
	Currency *string  `json:"currency"`
	Balance  *float64 `json:"balance"`
}

type HistoryRequest struct {
	AccountIDs []uuid.UUID
	From       time.Time
	To         time.Time
	Interval   string
}

type AccountService struct {
	repo repository.AccountRepository
}

func NewAccountService(repo repository.AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) List(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	return s.repo.List(ctx, userID)
}

func (s *AccountService) Get(ctx context.Context, id, userID uuid.UUID) (*models.Account, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *AccountService) Create(ctx context.Context, userID uuid.UUID, req CreateAccountRequest) (*models.Account, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if !validAccountTypes[req.Type] {
		return nil, fmt.Errorf("type must be one of: checking, savings, credit, investment")
	}
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	return s.repo.Create(ctx, userID, req.Name, req.Type, currency, req.Balance)
}

func (s *AccountService) Update(ctx context.Context, id, userID uuid.UUID, req UpdateAccountRequest) (*models.Account, error) {
	if req.Type != nil && !validAccountTypes[*req.Type] {
		return nil, fmt.Errorf("type must be one of: checking, savings, credit, investment")
	}
	fields := repository.AccountUpdateFields{
		Name:     req.Name,
		Type:     req.Type,
		Currency: req.Currency,
		Balance:  req.Balance,
	}
	return s.repo.Update(ctx, id, userID, fields)
}

func (s *AccountService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	// Verify ownership before attempting delete
	_, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id, userID)
}

func (s *AccountService) GetHistory(ctx context.Context, userID uuid.UUID, req HistoryRequest) ([]models.BalanceSnapshot, error) {
	if req.From.IsZero() || req.To.IsZero() {
		return nil, fmt.Errorf("from and to dates are required")
	}
	if req.From.After(req.To) {
		return nil, fmt.Errorf("from must be before or equal to to")
	}

	accountIDs := req.AccountIDs
	if len(accountIDs) == 0 {
		accounts, err := s.repo.List(ctx, userID)
		if err != nil {
			return nil, err
		}
		for _, a := range accounts {
			accountIDs = append(accountIDs, a.ID)
		}
	} else {
		for _, id := range accountIDs {
			if _, err := s.repo.GetByID(ctx, id, userID); err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return nil, fmt.Errorf("account %s not found", id)
				}
				return nil, err
			}
		}
	}

	if len(accountIDs) == 0 {
		return []models.BalanceSnapshot{}, nil
	}

	interval := req.Interval
	if interval == "" {
		interval = "day"
	}

	return s.repo.ListBalanceSnapshots(ctx, accountIDs, req.From, req.To, interval)
}
