package handler

import (
	"net/http"
	"strconv"

	"github.com/111zxc/rss-communicator/internal/handler/dto"
	"github.com/111zxc/rss-communicator/internal/service"
)

type ContactHandler struct {
	contactService *service.ContactService
}

func NewContactHandler(contactService *service.ContactService) *ContactHandler {
	return &ContactHandler{contactService: contactService}
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
		out = append(out, dto.ContactResponse{
			ID:          c.ID,
			Type:        string(c.Type),
			Value:       c.Value,
			DisplayName: c.DisplayName,
			Status:      string(c.Status),
			VerifiedAt:  c.VerifiedAt,
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
		})
	}

	writeJSON(w, 200, dto.ListResponse[dto.ContactResponse]{Items: out, Total: total})
}
