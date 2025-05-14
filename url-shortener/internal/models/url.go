package models

import "time"

type URL struct {
	ID          int       `json:"id" db:"id"`
	OriginalURL string    `json:"original_url" db:"original_url"`
	ShortURL    string    `json:"short_url,omitempty" db:"short_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
