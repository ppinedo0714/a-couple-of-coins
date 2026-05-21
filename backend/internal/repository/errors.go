package repository

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrHasTransactions = errors.New("account has transactions")
)
