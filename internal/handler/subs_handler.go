package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/111zxc/rss-communicator/internal/handler/dto"
	"github.com/111zxc/rss-communicator/internal/service"
)

type SubscriptionHandler struct {
	subService *service.SubscriptionService
}

func NewSubscriptionHandler(subService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subService: subService}
}

func (h *SubscriptionHandler) ListByFeed(w http.ResponseWriter, r *http.Request) {
	feedID := chi.URLParam(r, "feedID")
	if feedID == "" {
		writeError(w, 400, "INVALID_INPUT", "feedID is required", nil)
		return
	}

	subs, err := h.subService.ListByFeed(r.Context(), feedID)
	if err != nil {
		switch err {
		case service.ErrBadRequest:
			writeError(w, 400, "INVALID_INPUT", "invalid feed id", nil)
		case service.ErrNotFound:
			writeError(w, 404, "NOT_FOUND", "feed not found", nil)
		default:
			writeError(w, 500, "INTERNAL_ERROR", "failed to list subscriptions", nil)
		}
		return
	}

	out := make([]dto.SubscriptionResponse, 0, len(subs))
	for _, sub := range subs {
		out = append(out, dto.SubscriptionResponse{
			FeedID:    sub.FeedID,
			ContactID: sub.ContactID,
			Source:    string(sub.Source),
			GroupID:   sub.GroupID,
			CreatedAt: sub.CreatedAt,
		})
	}
	writeJSON(w, 200, dto.ListResponse[dto.SubscriptionResponse]{Items: out, Total: len(out)})
}

func (h *SubscriptionHandler) ListByContact(w http.ResponseWriter, r *http.Request) {
	contactID := chi.URLParam(r, "contactID")
	if contactID == "" {
		writeError(w, 400, "INVALID_INPUT", "contactID is required", nil)
		return
	}

	subs, err := h.subService.ListByContact(r.Context(), contactID)
	if err != nil {
		switch err {
		case service.ErrBadRequest:
			writeError(w, 400, "INVALID_INPUT", "invalid contact id", nil)
		case service.ErrNotFound:
			writeError(w, 404, "NOT_FOUND", "contact not found", nil)
		default:
			writeError(w, 500, "INTERNAL_ERROR", "failed to list subscriptions", nil)
		}
		return
	}

	out := make([]dto.SubscriptionResponse, 0, len(subs))
	for _, sub := range subs {
		out = append(out, dto.SubscriptionResponse{
			FeedID:    sub.FeedID,
			ContactID: sub.ContactID,
			Source:    string(sub.Source),
			GroupID:   sub.GroupID,
			CreatedAt: sub.CreatedAt,
		})
	}
	writeJSON(w, 200, dto.ListResponse[dto.SubscriptionResponse]{Items: out, Total: len(out)})
}

// POST /api/v1/feeds/{feedID}/subscriptions
func (h *SubscriptionHandler) Bind(w http.ResponseWriter, r *http.Request) {
	feedID := chi.URLParam(r, "feedID")
	if feedID == "" {
		writeError(w, 400, "INVALID_INPUT", "feedID is required", nil)
		return
	}

	var req dto.BindSubscriptionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}
	if req.ContactID == "" {
		writeError(w, 400, "INVALID_INPUT", "contact_id is required", nil)
		return
	}

	if err := h.subService.Bind(r.Context(), feedID, req.ContactID); err != nil {
		writeError(w, 500, "INTERNAL_ERROR", "failed to bind subscription", nil)
		return
	}

	writeJSON(w, 201, map[string]any{"feed_id": feedID, "contact_id": req.ContactID})
}

// DELETE /api/v1/feeds/{feedID}/subscriptions/{contactID}
func (h *SubscriptionHandler) Unbind(w http.ResponseWriter, r *http.Request) {
	feedID := chi.URLParam(r, "feedID")
	contactID := chi.URLParam(r, "contactID")
	if feedID == "" || contactID == "" {
		writeError(w, 400, "INVALID_INPUT", "feedID and contactID are required", nil)
		return
	}

	if err := h.subService.Unbind(r.Context(), feedID, contactID); err != nil {
		writeError(w, 500, "INTERNAL_ERROR", "failed to unbind subscription", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
