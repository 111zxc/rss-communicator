package dto

type CreateFeedRequest struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	IntervalSeconds int    `json:"interval_seconds"`
	Enabled         bool   `json:"enabled"`
}

type BindSubscriptionRequest struct {
	ContactID string `json:"contact_id"`
}
