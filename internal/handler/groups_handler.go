package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/handler/dto"
	"github.com/111zxc/rss-communicator/internal/service"
)

type GroupHandler struct {
	groupService *service.GroupService
}

func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

func (h *GroupHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseListParams(r)
	items, total, err := h.groupService.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, 500, "INTERNAL_ERROR", "failed to list groups", nil)
		return
	}
	out := make([]dto.GroupResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapGroupResponse(item))
	}
	writeJSON(w, 200, dto.ListResponse[dto.GroupResponse]{Items: out, Total: total})
}

func (h *GroupHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.groupService.GetByID(r.Context(), chi.URLParam(r, "groupID"))
	if err != nil {
		writeServiceError(w, err, "failed to get group")
		return
	}
	writeJSON(w, 200, mapGroupResponse(item))
}

func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.GroupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}
	item, err := h.groupService.Create(r.Context(), service.GroupInput{Name: req.Name, Description: req.Description})
	if err != nil {
		writeServiceError(w, err, "failed to create group")
		return
	}
	writeJSON(w, 201, mapGroupResponse(item))
}

func (h *GroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req dto.GroupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}
	item, err := h.groupService.Update(r.Context(), chi.URLParam(r, "groupID"), service.GroupInput{Name: req.Name, Description: req.Description})
	if err != nil {
		writeServiceError(w, err, "failed to update group")
		return
	}
	writeJSON(w, 200, mapGroupResponse(item))
}

func (h *GroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.groupService.Delete(r.Context(), chi.URLParam(r, "groupID")); err != nil {
		writeServiceError(w, err, "failed to delete group")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	items, err := h.groupService.ListContacts(r.Context(), chi.URLParam(r, "groupID"))
	if err != nil {
		writeServiceError(w, err, "failed to list group contacts")
		return
	}
	out := make([]dto.ContactResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapContactResponse(item))
	}
	writeJSON(w, 200, dto.ListResponse[dto.ContactResponse]{Items: out, Total: len(out)})
}

func (h *GroupHandler) AddContact(w http.ResponseWriter, r *http.Request) {
	var req dto.GroupContactRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}
	if err := h.groupService.AddContact(r.Context(), chi.URLParam(r, "groupID"), req.ContactID); err != nil {
		writeServiceError(w, err, "failed to add group contact")
		return
	}
	writeJSON(w, 201, map[string]any{"group_id": chi.URLParam(r, "groupID"), "contact_id": req.ContactID})
}

func (h *GroupHandler) RemoveContact(w http.ResponseWriter, r *http.Request) {
	if err := h.groupService.RemoveContact(r.Context(), chi.URLParam(r, "groupID"), chi.URLParam(r, "contactID")); err != nil {
		writeServiceError(w, err, "failed to remove group contact")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *GroupHandler) ListFeeds(w http.ResponseWriter, r *http.Request) {
	items, err := h.groupService.ListFeeds(r.Context(), chi.URLParam(r, "groupID"))
	if err != nil {
		writeServiceError(w, err, "failed to list group feeds")
		return
	}
	out := make([]dto.FeedResponse, 0, len(items))
	for _, f := range items {
		out = append(out, dto.FeedResponse{
			ID:              f.ID,
			Name:            f.Name,
			URL:             f.URL,
			Enabled:         f.Enabled,
			IntervalSeconds: f.IntervalSeconds,
			BatchEnabled:    f.BatchEnabled,
			BatchWindowSecs: f.BatchWindowSecs,
			NextFetchAt:     f.NextFetchAt,
			LastFetchAt:     f.LastFetchAt,
			CreatedAt:       f.CreatedAt,
			UpdatedAt:       f.UpdatedAt,
		})
	}
	writeJSON(w, 200, dto.ListResponse[dto.FeedResponse]{Items: out, Total: len(out)})
}

func (h *GroupHandler) AddFeed(w http.ResponseWriter, r *http.Request) {
	var req dto.GroupFeedRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}
	if err := h.groupService.AddFeed(r.Context(), chi.URLParam(r, "groupID"), req.FeedID); err != nil {
		writeServiceError(w, err, "failed to add group feed")
		return
	}
	writeJSON(w, 201, map[string]any{"group_id": chi.URLParam(r, "groupID"), "feed_id": req.FeedID})
}

func (h *GroupHandler) RemoveFeed(w http.ResponseWriter, r *http.Request) {
	if err := h.groupService.RemoveFeed(r.Context(), chi.URLParam(r, "groupID"), chi.URLParam(r, "feedID")); err != nil {
		writeServiceError(w, err, "failed to remove group feed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mapGroupResponse(g domain.Group) dto.GroupResponse {
	return dto.GroupResponse{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

func parseListParams(r *http.Request) (int, int) {
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
	return limit, offset
}
