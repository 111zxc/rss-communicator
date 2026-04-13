package dto

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

type UpdateTelegramContactRequest struct {
	ChatID      string  `json:"chat_id"`
	Username    *string `json:"username"`
	DisplayName *string `json:"display_name"`
	Status      string  `json:"status"`
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
