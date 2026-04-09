package dto

type CreateFeedRequest struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	IntervalSeconds int    `json:"interval_seconds"`
	Enabled         bool   `json:"enabled"`
	BatchEnabled    bool   `json:"batch_enabled"`
	BatchWindowSecs int    `json:"batch_window_seconds"`
}

type UpdateFeedRequest struct {
	BatchEnabled    *bool `json:"batch_enabled"`
	BatchWindowSecs *int  `json:"batch_window_seconds"`
}

type BindSubscriptionRequest struct {
	ContactID string `json:"contact_id"`
}
