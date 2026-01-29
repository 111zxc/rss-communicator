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
