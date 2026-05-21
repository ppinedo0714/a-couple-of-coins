package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	AccountID    uuid.UUID  `json:"account_id"`
	CategoryID   *uuid.UUID `json:"category_id"`
	Amount       float64    `json:"amount"`
	Description  string     `json:"description"`
	MerchantName *string    `json:"merchant_name"`
	Date         time.Time  `json:"date"`
	Source       string     `json:"source"`
	Classified   bool       `json:"classified"`
	CreatedAt    time.Time  `json:"created_at"`
}

type CreateTransactionRequest struct {
	AccountID   uuid.UUID  `json:"account_id"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Amount      float64    `json:"amount"`
	Description string     `json:"description"`
	Date        string     `json:"date"` // YYYY-MM-DD
}

type UpdateTransactionRequest struct {
	CategoryID    *uuid.UUID `json:"category_id"` // explicit null removes category
	ClearCategory bool       `json:"-"`            // set by handler when category_id was explicitly null in JSON
	Description   *string    `json:"description"`
	Amount        *float64   `json:"amount"`
	Date          *string    `json:"date"`
}

type ClassifyResult struct {
	Classified int `json:"classified"`
	Failed     int `json:"failed"`
}
