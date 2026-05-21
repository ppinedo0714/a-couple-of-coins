package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	ParentID  *uuid.UUID `json:"parent_id"`
	Name      string     `json:"name"`
	Color     *string    `json:"color"`
	CreatedAt time.Time  `json:"created_at"`
}
