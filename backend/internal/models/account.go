package models

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

type BalanceSnapshot struct {
	AccountID uuid.UUID `json:"account_id"`
	Date      time.Time `json:"date"`
	Balance   float64   `json:"balance"`
}
