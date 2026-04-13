package handler

import "github.com/111zxc/rss-communicator/internal/service"

type Handler struct {
	Feed         *FeedHandler
	Contact      *ContactHandler
	Subscription *SubscriptionHandler
	Group        *GroupHandler
	RegCode      *RegistrationCodeHandler
	Health       *HealthHandler
}

func New(feed *service.FeedService, contact *service.ContactService, contactDelivery *service.ContactDeliveryService, sub *service.SubscriptionService, group *service.GroupService, regCode *service.RegistrationCodeService) *Handler {
	return &Handler{
		Feed:         NewFeedHandler(feed),
		Contact:      NewContactHandler(contact, contactDelivery),
		Subscription: NewSubscriptionHandler(sub),
		Group:        NewGroupHandler(group),
		RegCode:      NewRegistrationCodeHandler(regCode),
		Health:       NewHealthHandler(),
	}
}
