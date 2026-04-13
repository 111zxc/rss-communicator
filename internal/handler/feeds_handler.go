package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/111zxc/rss-communicator/internal/handler/dto"
	"github.com/111zxc/rss-communicator/internal/service"
)

type FeedHandler struct {
	feedService *service.FeedService
}

func NewFeedHandler(feedService *service.FeedService) *FeedHandler {
	return &FeedHandler{feedService: feedService}
}

func (h *FeedHandler) List(w http.ResponseWriter, r *http.Request) {
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

	feeds, total, err := h.feedService.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, 500, "INTERNAL_ERROR", "failed to list feeds", nil)
		return
	}

	out := make([]dto.FeedResponse, 0, len(feeds))
	for _, f := range feeds {
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

	writeJSON(w, 200, dto.ListResponse[dto.FeedResponse]{Items: out, Total: total})
}

func (h *FeedHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateFeedRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}

	f, err := h.feedService.Create(r.Context(), service.CreateFeedInput{
		Name:            req.Name,
		URL:             req.URL,
		IntervalSeconds: req.IntervalSeconds,
		Enabled:         boolOrDefault(req.Enabled, true),
		BatchEnabled:    boolOrDefault(req.BatchEnabled, false),
		BatchWindowSecs: req.BatchWindowSecs,
	})
	if err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid feed fields", nil)
		return
	}

	writeJSON(w, 201, dto.FeedResponse{
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

func boolOrDefault(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func (h *FeedHandler) Update(w http.ResponseWriter, r *http.Request) {
	feedID := chi.URLParam(r, "feedID")
	if feedID == "" {
		writeError(w, 400, "INVALID_INPUT", "feedID is required", nil)
		return
	}

	var req dto.UpdateFeedRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid request body", nil)
		return
	}

	f, err := h.feedService.Update(r.Context(), feedID, service.UpdateFeedInput{
		BatchEnabled:    req.BatchEnabled,
		BatchWindowSecs: req.BatchWindowSecs,
	})
	if err != nil {
		writeError(w, 400, "INVALID_INPUT", "invalid feed fields", nil)
		return
	}

	writeJSON(w, 200, dto.FeedResponse{
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

func (h *FeedHandler) Delete(w http.ResponseWriter, r *http.Request) {
	feedID := chi.URLParam(r, "feedID")
	if feedID == "" {
		writeError(w, 400, "INVALID_INPUT", "feedID is required", nil)
		return
	}

	if err := h.feedService.Delete(r.Context(), feedID); err != nil {
		writeError(w, 500, "INTERNAL_ERROR", "failed to delete feed", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
