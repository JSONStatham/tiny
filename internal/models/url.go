package models

import "time"

type URL struct {
	ID        int       `json:"id" db:"id"`
	URL       string    `json:"url" db:"url"`
	Alias     string    `json:"alias,omitempty" db:"alias"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
