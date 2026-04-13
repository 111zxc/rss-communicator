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
			r.Post("/http", h.Contact.CreateHTTP)
			r.Put("/telegram/{contactID}", h.Contact.UpdateTelegram)
			r.Put("/http/{contactID}", h.Contact.UpdateHTTP)
			r.Post("/{contactID}/test-send", h.Contact.TestSend)
			r.Delete("/{contactID}", h.Contact.Delete)
		})
	})

	return r
}
