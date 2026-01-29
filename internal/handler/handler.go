package handler

import "github.com/111zxc/rss-communicator/internal/service"

type Handler struct {
	Feed         *FeedHandler
	Contact      *ContactHandler
	Subscription *SubscriptionHandler
	Health       *HealthHandler
}

func New(feed *service.FeedService, contact *service.ContactService, sub *service.SubscriptionService) *Handler {
	return &Handler{
		Feed:         NewFeedHandler(feed),
		Contact:      NewContactHandler(contact),
		Subscription: NewSubscriptionHandler(sub),
		Health:       NewHealthHandler(),
	}
}
