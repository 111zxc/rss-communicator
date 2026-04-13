package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/handler/dto"
	"github.com/111zxc/rss-communicator/internal/service"
)

type ContactHandler struct {
	contactService         *service.ContactService
	contactDeliveryService *service.ContactDeliveryService
}

func NewContactHandler(contactService *service.ContactService, contactDeliveryService *service.ContactDeliveryService) *ContactHandler {
	return &ContactHandler{
		contactService:         contactService,
		contactDeliveryService: contactDeliveryService,
	}
}

func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	cs, total, err := h.contactService.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, 500, "INTERNAL_ERROR", "failed to list contacts", nil)
		return
	}

	out := make([]dto.ContactResponse, 0, len(cs))
	for _, c := range cs {
		out = append(out, mapContactResponse(c))
	}

	writeJSON(w, 200, dto.ListResponse[dto.ContactResponse]{Items: out, Total: total})
}

func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	contactID := chi.URLParam(r, "contactID")
	contact, err := h.contactService.GetByID(r.Context(), contactID)
	if err != nil {
		writeServiceError(w, err, "failed to get contact")
		return
	}
	writeJSON(w, 200, mapContactResponse(contact))
}

func (h *ContactHandler) CreateTelegram(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTelegramContactRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}

	contact, err := h.contactService.CreateTelegram(r.Context(), service.CreateTelegramContactInput{
		ChatID:      req.ChatID,
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Status:      req.Status,
	})
	if err != nil {
		writeServiceError(w, err, "failed to create telegram contact")
		return
	}

	writeJSON(w, 201, mapContactResponse(contact))
}

func (h *ContactHandler) CreateHTTP(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateHTTPContactRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}

	contact, err := h.contactService.CreateHTTP(r.Context(), service.CreateHTTPContactInput{
		DisplayName:  req.DisplayName,
		Status:       req.Status,
		Method:       req.Method,
		URL:          req.URL,
		Headers:      req.Headers,
		BodyTemplate: req.BodyTemplate,
	})
	if err != nil {
		writeServiceError(w, err, "failed to create http contact")
		return
	}

	writeJSON(w, 201, mapContactResponse(contact))
}

func (h *ContactHandler) UpdateTelegram(w http.ResponseWriter, r *http.Request) {
	contactID := chi.URLParam(r, "contactID")

	var req dto.UpdateTelegramContactRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}

	contact, err := h.contactService.UpdateTelegram(r.Context(), contactID, service.UpdateTelegramContactInput{
		ChatID:      req.ChatID,
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Status:      req.Status,
	})
	if err != nil {
		writeServiceError(w, err, "failed to update telegram contact")
		return
	}

	writeJSON(w, 200, mapContactResponse(contact))
}

func (h *ContactHandler) UpdateHTTP(w http.ResponseWriter, r *http.Request) {
	contactID := chi.URLParam(r, "contactID")

	var req dto.UpdateHTTPContactRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}

	contact, err := h.contactService.UpdateHTTP(r.Context(), contactID, service.UpdateHTTPContactInput{
		DisplayName:  req.DisplayName,
		Status:       req.Status,
		Method:       req.Method,
		URL:          req.URL,
		Headers:      req.Headers,
		BodyTemplate: req.BodyTemplate,
	})
	if err != nil {
		writeServiceError(w, err, "failed to update http contact")
		return
	}

	writeJSON(w, 200, mapContactResponse(contact))
}

func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	contactID := chi.URLParam(r, "contactID")
	if err := h.contactService.Delete(r.Context(), contactID); err != nil {
		writeServiceError(w, err, "failed to delete contact")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ContactHandler) TestSend(w http.ResponseWriter, r *http.Request) {
	contactID := chi.URLParam(r, "contactID")

	var req dto.ContactTestSendRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}

	if h.contactDeliveryService == nil {
		writeError(w, 500, "INTERNAL_ERROR", "test send is not configured", nil)
		return
	}

	err := h.contactDeliveryService.TestSend(r.Context(), contactID, service.TestSendInput{
		FeedName: req.FeedName,
		FeedURL:  req.FeedURL,
		Title:    req.Title,
		Link:     req.Link,
		Summary:  req.Summary,
		Author:   req.Author,
	})
	if err != nil {
		writeServiceError(w, err, "failed to send test message")
		return
	}

	writeJSON(w, 200, map[string]any{"ok": true})
}

func mapContactResponse(c domain.Contact) dto.ContactResponse {
	resp := dto.ContactResponse{
		ID:          c.ID,
		Type:        string(c.Type),
		Value:       c.Value,
		DisplayName: c.DisplayName,
		Status:      string(c.Status),
		VerifiedAt:  c.VerifiedAt,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
	if c.Telegram != nil {
		resp.Username = c.Telegram.Username
	}
	if c.HTTP != nil {
		headers := make(map[string]string, len(c.HTTP.Headers))
		for k, v := range c.HTTP.Headers {
			headers[k] = v
		}
		resp.HTTP = &dto.HTTPContactResponse{
			Method:       c.HTTP.Method,
			URL:          c.HTTP.URL,
			Headers:      headers,
			BodyTemplate: c.HTTP.BodyTemplate,
		}
	}
	return resp
}

func writeServiceError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrBadRequest):
		writeError(w, 400, "INVALID_INPUT", fallback, nil)
	case errors.Is(err, service.ErrNotFound):
		writeError(w, 404, "NOT_FOUND", fallback, nil)
	default:
		writeError(w, 500, "INTERNAL_ERROR", fallback, nil)
	}
}
