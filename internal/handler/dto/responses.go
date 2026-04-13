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
	ID          string                `json:"id"`
	Type        string                `json:"type"`
	Value       string                `json:"value"`
	Username    *string               `json:"username,omitempty"`
	Email       *EmailContactResponse `json:"email,omitempty"`
	DisplayName *string               `json:"display_name,omitempty"`
	HTTP        *HTTPContactResponse  `json:"http,omitempty"`
	Status      string                `json:"status"`
	VerifiedAt  *time.Time            `json:"verified_at,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

type EmailContactResponse struct {
	Format string `json:"format"`
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
	Source    string    `json:"source,omitempty"`
	GroupID   *string   `json:"group_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type GroupResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RegistrationCodeResponse struct {
	ID          string     `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Enabled     bool       `json:"enabled"`
	MaxUses     *int       `json:"max_uses,omitempty"`
	UseCount    int        `json:"use_count"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
