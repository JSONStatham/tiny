package models

import "time"

type UrlEvent struct {
	EventType  string    `json:"event_type"`
	ShortURL   string    `json:"short_url"`
	OriginaUrl string    `json:"original_url,omitempty"`
	UserID     string    `json:"user_id"`
	EventTime  time.Time `json:"event_time"`
	RequestMeta
}

type RequestMeta struct {
	IPAddress  string `json:"ip_address,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
	Referrer   string `json:"referrer,omitempty"`
	Country    string `json:"country,omitempty"`
	Region     string `json:"region,omitempty"`
	City       string `json:"city,omitempty"`
	Browser    string `json:"browser,omitempty"`
	OS         string `json:"os,omitempty"`
	DeviceType string `json:"device_type,omitempty"`
}
