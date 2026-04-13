package dto

import "time"

type CreateFeedRequest struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	IntervalSeconds int    `json:"interval_seconds"`
	Enabled         *bool  `json:"enabled"`
	BatchEnabled    *bool  `json:"batch_enabled"`
	BatchWindowSecs int    `json:"batch_window_seconds"`
}

type UpdateFeedRequest struct {
	BatchEnabled    *bool `json:"batch_enabled"`
	BatchWindowSecs *int  `json:"batch_window_seconds"`
}

type BindSubscriptionRequest struct {
	ContactID string `json:"contact_id"`
}

type CreateHTTPContactRequest struct {
	DisplayName  *string           `json:"display_name"`
	Status       string            `json:"status"`
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	BodyTemplate *string           `json:"body_template"`
}

type CreateTelegramContactRequest struct {
	ChatID      string  `json:"chat_id"`
	Username    *string `json:"username"`
	DisplayName *string `json:"display_name"`
	Status      string  `json:"status"`
}

type CreateEmailContactRequest struct {
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
	Status      string  `json:"status"`
	Format      string  `json:"format"`
}

type UpdateTelegramContactRequest struct {
	ChatID      string  `json:"chat_id"`
	Username    *string `json:"username"`
	DisplayName *string `json:"display_name"`
	Status      string  `json:"status"`
}

type UpdateEmailContactRequest struct {
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
	Status      string  `json:"status"`
	Format      string  `json:"format"`
}

type UpdateHTTPContactRequest struct {
	DisplayName  *string           `json:"display_name"`
	Status       string            `json:"status"`
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      map[string]string `json:"headers"`
	BodyTemplate *string           `json:"body_template"`
}

type ContactTestSendRequest struct {
	FeedName string  `json:"feed_name"`
	FeedURL  string  `json:"feed_url"`
	Title    string  `json:"title"`
	Link     string  `json:"link"`
	Summary  *string `json:"summary"`
	Author   *string `json:"author"`
}

type GroupRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type GroupContactRequest struct {
	ContactID string `json:"contact_id"`
}

type GroupFeedRequest struct {
	FeedID string `json:"feed_id"`
}

type RegistrationCodeRequest struct {
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Enabled     bool       `json:"enabled"`
	MaxUses     *int       `json:"max_uses"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type RegistrationCodeGroupRequest struct {
	GroupID string `json:"group_id"`
}
