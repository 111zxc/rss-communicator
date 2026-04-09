package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/service"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	NewHealthHandler().Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestFeedHandlerCreate(t *testing.T) {
	feedSvc := service.NewFeedService(&handlerFeedRepoStub{}, handlerClock{})
	h := NewFeedHandler(feedSvc)

	body := []byte(`{"name":"HN","url":"https://news.ycombinator.com/rss","interval_seconds":300,"enabled":true,"batch_enabled":true,"batch_window_seconds":3600}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/feeds", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var got dtoFeedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !got.BatchEnabled || got.BatchWindowSecs != 3600 {
		t.Fatalf("unexpected feed response: %+v", got)
	}
}

func TestFeedHandlerUpdateRequiresFeedID(t *testing.T) {
	feedSvc := service.NewFeedService(&handlerFeedRepoStub{}, handlerClock{})
	h := NewFeedHandler(feedSvc)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/feeds/missing", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestContactHandlerList(t *testing.T) {
	contactSvc := service.NewContactService(&handlerContactsRepoStub{
		contacts: []domain.Contact{{ID: "contact-1", Type: domain.ContactTelegram, Status: domain.ContactActive, Value: "123"}},
	})
	h := NewContactHandler(contactSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts?limit=10&offset=5", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSubscriptionHandlerBind(t *testing.T) {
	subs := &handlerSubsRepoStub{}
	subSvc := service.NewSubscriptionService(
		subs,
		&handlerFeedRepoStub{getByIDFeed: domain.Feed{ID: "feed-1"}},
		&handlerContactsRepoStub{contactByID: domain.Contact{ID: "contact-1"}},
	)
	h := NewSubscriptionHandler(subSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/feeds/feed-1/subscriptions", bytes.NewReader([]byte(`{"contact_id":"contact-1"}`)))
	req = withURLParam(req, "feedID", "feed-1")
	rec := httptest.NewRecorder()

	h.Bind(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !subs.addCalled {
		t.Fatal("expected Add to be called")
	}
}

func TestSubscriptionHandlerUnbind(t *testing.T) {
	subs := &handlerSubsRepoStub{}
	subSvc := service.NewSubscriptionService(subs, &handlerFeedRepoStub{}, &handlerContactsRepoStub{})
	h := NewSubscriptionHandler(subSvc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/feeds/feed-1/subscriptions/contact-1", nil)
	req = withURLParam(req, "feedID", "feed-1")
	req = withURLParam(req, "contactID", "contact-1")
	rec := httptest.NewRecorder()

	h.Unbind(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

type dtoFeedResponse struct {
	BatchEnabled    bool `json:"batch_enabled"`
	BatchWindowSecs int  `json:"batch_window_seconds"`
}

type handlerClock struct{}

func (handlerClock) NowUTC() time.Time {
	return time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
}

type handlerFeedRepoStub struct {
	getByIDFeed domain.Feed
}

func (r *handlerFeedRepoStub) Create(_ context.Context, f domain.Feed) (domain.Feed, error) {
	f.ID = "feed-1"
	return f, nil
}

func (r *handlerFeedRepoStub) ListDue(context.Context, time.Time, int) ([]domain.Feed, error) {
	return nil, nil
}

func (r *handlerFeedRepoStub) MarkFetched(context.Context, string, time.Time, time.Time, *string, *string) error {
	return nil
}

func (r *handlerFeedRepoStub) MarkFetchError(context.Context, string, string) error {
	return nil
}

func (r *handlerFeedRepoStub) GetByID(context.Context, string) (domain.Feed, error) {
	if r.getByIDFeed.ID == "" {
		return domain.Feed{ID: "feed-1", BatchWindowSecs: 3600}, nil
	}
	return r.getByIDFeed, nil
}

func (r *handlerFeedRepoStub) UpdateBatching(_ context.Context, feedID string, batchEnabled bool, batchWindowSecs int) (domain.Feed, error) {
	return domain.Feed{ID: feedID, BatchEnabled: batchEnabled, BatchWindowSecs: batchWindowSecs}, nil
}

func (r *handlerFeedRepoStub) List(context.Context, int, int) ([]domain.Feed, int, error) {
	return nil, 0, nil
}

func (r *handlerFeedRepoStub) Delete(context.Context, string) error {
	return nil
}

type handlerContactsRepoStub struct {
	contacts    []domain.Contact
	contactByID domain.Contact
}

func (r *handlerContactsRepoStub) UpsertTelegramActive(context.Context, string, *string, *string, time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (r *handlerContactsRepoStub) GetByTypeValue(context.Context, domain.ContactType, string) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (r *handlerContactsRepoStub) GetByID(context.Context, string) (domain.Contact, error) {
	if r.contactByID.ID == "" {
		return domain.Contact{ID: "contact-1"}, nil
	}
	return r.contactByID, nil
}

func (r *handlerContactsRepoStub) List(context.Context, int, int) ([]domain.Contact, int, error) {
	return r.contacts, len(r.contacts), nil
}

type handlerSubsRepoStub struct {
	addCalled bool
}

func (r *handlerSubsRepoStub) ListByFeed(context.Context, string) ([]domain.Subscription, error) {
	return nil, nil
}

func (r *handlerSubsRepoStub) Add(context.Context, string, string) error {
	r.addCalled = true
	return nil
}

func (r *handlerSubsRepoStub) Remove(context.Context, string, string) error {
	return nil
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	routeCtx := chi.RouteContext(r.Context())
	if routeCtx == nil {
		routeCtx = chi.NewRouteContext()
	}
	routeCtx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}
