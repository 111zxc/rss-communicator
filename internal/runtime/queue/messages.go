package queue

type FetchJob struct {
	FeedID string `json:"feed_id"`
}

type DeliverJob struct {
	DeliveryID string `json:"delivery_id"`
}
