package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/111zxc/rss-communicator/internal/handler"
)

func NewRouter(h *handler.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Compress(5))

	r.Get("/healthz", h.Health.Health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/feeds", func(r chi.Router) {
			r.Get("/", h.Feed.List)
			r.Post("/", h.Feed.Create)
			r.Patch("/{feedID}", h.Feed.Update)
			r.Delete("/{feedID}", h.Feed.Delete)

			r.Route("/{feedID}/subscriptions", func(r chi.Router) {
				r.Get("/", h.Subscription.ListByFeed)
				r.Post("/", h.Subscription.Bind)
				r.Delete("/{contactID}", h.Subscription.Unbind)
			})
		})

		r.Route("/contacts", func(r chi.Router) {
			r.Get("/", h.Contact.List)
			r.Get("/{contactID}", h.Contact.Get)
			r.Get("/{contactID}/subscriptions", h.Subscription.ListByContact)
			r.Post("/telegram", h.Contact.CreateTelegram)
			r.Post("/email", h.Contact.CreateEmail)
			r.Post("/http", h.Contact.CreateHTTP)
			r.Put("/telegram/{contactID}", h.Contact.UpdateTelegram)
			r.Put("/email/{contactID}", h.Contact.UpdateEmail)
			r.Put("/http/{contactID}", h.Contact.UpdateHTTP)
			r.Post("/{contactID}/test-send", h.Contact.TestSend)
			r.Delete("/{contactID}", h.Contact.Delete)
		})

		r.Route("/groups", func(r chi.Router) {
			r.Get("/", h.Group.List)
			r.Post("/", h.Group.Create)
			r.Get("/{groupID}", h.Group.Get)
			r.Put("/{groupID}", h.Group.Update)
			r.Delete("/{groupID}", h.Group.Delete)
			r.Get("/{groupID}/contacts", h.Group.ListContacts)
			r.Post("/{groupID}/contacts", h.Group.AddContact)
			r.Delete("/{groupID}/contacts/{contactID}", h.Group.RemoveContact)
			r.Get("/{groupID}/feeds", h.Group.ListFeeds)
			r.Post("/{groupID}/feeds", h.Group.AddFeed)
			r.Delete("/{groupID}/feeds/{feedID}", h.Group.RemoveFeed)
		})

		r.Route("/registration-codes", func(r chi.Router) {
			r.Get("/", h.RegCode.List)
			r.Post("/", h.RegCode.Create)
			r.Get("/{codeID}", h.RegCode.Get)
			r.Put("/{codeID}", h.RegCode.Update)
			r.Delete("/{codeID}", h.RegCode.Delete)
			r.Get("/{codeID}/groups", h.RegCode.ListGroups)
			r.Post("/{codeID}/groups", h.RegCode.AddGroup)
			r.Delete("/{codeID}/groups/{groupID}", h.RegCode.RemoveGroup)
		})
	})

	return r
}
