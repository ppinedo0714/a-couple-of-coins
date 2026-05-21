package models

import (
	"time"

	"github.com/google/uuid"
)

type ImportJob struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	Status       string     `json:"status"`
	SourceType   string     `json:"source_type"`
	FileName     string     `json:"file_name"`
	RowsTotal    *int       `json:"rows_total"`
	RowsImported int        `json:"rows_imported"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at"`
}
