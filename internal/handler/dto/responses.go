package dto

import "time"

type FeedResponse struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	URL             string     `json:"url"`
	Enabled         bool       `json:"enabled"`
	IntervalSeconds int        `json:"interval_seconds"`
	BatchEnabled    bool       `json:"batch_enabled"`
	BatchWindowSecs int        `json:"batch_window_seconds"`
	NextFetchAt     *time.Time `json:"next_fetch_at,omitempty"`
	LastFetchAt     *time.Time `json:"last_fetch_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ListResponse[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

type ContactResponse struct {
	ID          string               `json:"id"`
	Type        string               `json:"type"`
	Value       string               `json:"value"`
	Username    *string              `json:"username,omitempty"`
	DisplayName *string              `json:"display_name,omitempty"`
	HTTP        *HTTPContactResponse `json:"http,omitempty"`
	Status      string               `json:"status"`
	VerifiedAt  *time.Time           `json:"verified_at,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type HTTPContactResponse struct {
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	BodyTemplate *string           `json:"body_template,omitempty"`
}

type SubscriptionResponse struct {
	FeedID    string    `json:"feed_id"`
	ContactID string    `json:"contact_id"`
	CreatedAt time.Time `json:"created_at"`
}
