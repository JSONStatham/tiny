package models

import "time"

type URL struct {
	ID        int       `json:"id"`
	URL       string    `json:"url"`
	Alias     string    `json:"alias,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
