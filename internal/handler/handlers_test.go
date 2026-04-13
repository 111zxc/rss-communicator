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

func TestFeedHandlerCreateDefaultsEnabledTrue(t *testing.T) {
	repo := &handlerFeedRepoStub{}
	feedSvc := service.NewFeedService(repo, handlerClock{})
	h := NewFeedHandler(feedSvc)

	body := []byte(`{"name":"HN","url":"https://news.ycombinator.com/rss","interval_seconds":300}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/feeds", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !repo.created.Enabled {
		t.Fatal("expected omitted enabled to default to true")
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
		contacts: []domain.Contact{{
			ID:       "contact-1",
			Type:     domain.ContactTelegram,
			Status:   domain.ContactActive,
			Value:    "123",
			Telegram: &domain.TelegramContactConfig{Username: strPtr("alice")},
		}},
	})
	h := NewContactHandler(contactSvc, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts?limit=10&offset=5", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestContactHandlerCreateHTTP(t *testing.T) {
	repo := &handlerContactsRepoStub{
		createdHTTP: domain.Contact{
			ID:     "contact-http-1",
			Type:   domain.ContactHTTP,
			Status: domain.ContactActive,
			Value:  "https://example.com/hook",
			HTTP: &domain.HTTPContactConfig{
				Method:       "POST",
				URL:          "https://example.com/hook",
				Headers:      map[string]string{"Content-Type": "application/json"},
				BodyTemplate: strPtr("{\"text\": {json_text}}"),
			},
		},
	}
	contactSvc := service.NewContactService(repo)
	h := NewContactHandler(contactSvc, nil)

	body := []byte(`{"display_name":"Webhook","status":"active","method":"post","url":"https://example.com/hook","headers":{"Content-Type":"application/json"},"body_template":"{\"text\": {json_text}}"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/contacts/http", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.CreateHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	if repo.createHTTPInput.Method != "POST" {
		t.Fatalf("expected normalized POST method, got %q", repo.createHTTPInput.Method)
	}
	if repo.createHTTPStatus != domain.ContactActive {
		t.Fatalf("expected active status, got %s", repo.createHTTPStatus)
	}
}

func TestContactHandlerCreateTelegram(t *testing.T) {
	repo := &handlerContactsRepoStub{
		createdTelegram: domain.Contact{
			ID:     "contact-tg-1",
			Type:   domain.ContactTelegram,
			Status: domain.ContactActive,
			Value:  "123456",
			Telegram: &domain.TelegramContactConfig{
				Username: strPtr("alice"),
			},
		},
	}
	h := NewContactHandler(service.NewContactService(repo), nil)

	body := []byte(`{"chat_id":"123456","username":"alice","display_name":"Alice","status":"active"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/contacts/telegram", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.CreateTelegram(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	if repo.createTelegramChatID != "123456" {
		t.Fatalf("unexpected chat id: %s", repo.createTelegramChatID)
	}
}

func TestContactHandlerUpdateHTTP(t *testing.T) {
	repo := &handlerContactsRepoStub{
		updatedHTTP: domain.Contact{
			ID:     "contact-http-1",
			Type:   domain.ContactHTTP,
			Status: domain.ContactDisabled,
			Value:  "https://example.com/updated",
			HTTP: &domain.HTTPContactConfig{
				Method:  "PUT",
				URL:     "https://example.com/updated",
				Headers: map[string]string{"X-Test": "1"},
			},
		},
	}
	h := NewContactHandler(service.NewContactService(repo), nil)

	body := []byte(`{"display_name":"Webhook","status":"disabled","method":"put","url":"https://example.com/updated","headers":{"X-Test":"1"}}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/contacts/http/contact-http-1", bytes.NewReader(body))
	req = withURLParam(req, "contactID", "contact-http-1")
	rec := httptest.NewRecorder()

	h.UpdateHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if repo.updateHTTPContactID != "contact-http-1" || repo.updateHTTPInput.Method != "PUT" {
		t.Fatalf("unexpected update call: id=%s method=%s", repo.updateHTTPContactID, repo.updateHTTPInput.Method)
	}
}

func TestContactHandlerUpdateTelegram(t *testing.T) {
	repo := &handlerContactsRepoStub{
		updatedTelegram: domain.Contact{
			ID:     "contact-tg-1",
			Type:   domain.ContactTelegram,
			Status: domain.ContactDisabled,
			Value:  "999",
			Telegram: &domain.TelegramContactConfig{
				Username: strPtr("bob"),
			},
		},
	}
	h := NewContactHandler(service.NewContactService(repo), nil)

	body := []byte(`{"chat_id":"999","username":"bob","display_name":"Bob","status":"disabled"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/contacts/telegram/contact-tg-1", bytes.NewReader(body))
	req = withURLParam(req, "contactID", "contact-tg-1")
	rec := httptest.NewRecorder()

	h.UpdateTelegram(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if repo.updateTelegramContactID != "contact-tg-1" || repo.updateTelegramStatus != domain.ContactDisabled {
		t.Fatalf("unexpected update call: id=%s status=%s", repo.updateTelegramContactID, repo.updateTelegramStatus)
	}
}

func TestContactHandlerDelete(t *testing.T) {
	repo := &handlerContactsRepoStub{}
	h := NewContactHandler(service.NewContactService(repo), nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/contacts/contact-1", nil)
	req = withURLParam(req, "contactID", "contact-1")
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if repo.deletedContactID != "contact-1" {
		t.Fatalf("unexpected deleted id: %s", repo.deletedContactID)
	}
}

func TestSubscriptionHandlerBind(t *testing.T) {
	subs := &handlerSubsRepoStub{}
	subSvc := service.NewSubscriptionService(
		subs,
		&handlerFeedRepoStub{getByIDFeed: domain.Feed{ID: "feed-1"}},
		&handlerContactsRepoStub{contactByID: domain.Contact{ID: "contact-1", Type: domain.ContactHTTP}},
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

func TestSubscriptionHandlerListByContact(t *testing.T) {
	subs := &handlerSubsRepoStub{
		listByContact: []domain.Subscription{{
			FeedID:    "feed-1",
			ContactID: "contact-1",
			CreatedAt: time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC),
		}},
	}
	subSvc := service.NewSubscriptionService(subs, &handlerFeedRepoStub{}, &handlerContactsRepoStub{
		contactByID: domain.Contact{ID: "contact-1", Type: domain.ContactHTTP},
	})
	h := NewSubscriptionHandler(subSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/contacts/contact-1/subscriptions", nil)
	req = withURLParam(req, "contactID", "contact-1")
	rec := httptest.NewRecorder()

	h.ListByContact(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !subs.listByContactCalled {
		t.Fatal("expected ListByContact to be called")
	}
}

func TestSubscriptionHandlerListByFeed(t *testing.T) {
	subs := &handlerSubsRepoStub{
		listByFeed: []domain.Subscription{{
			FeedID:    "feed-1",
			ContactID: "contact-1",
			CreatedAt: time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC),
		}},
	}
	subSvc := service.NewSubscriptionService(subs, &handlerFeedRepoStub{getByIDFeed: domain.Feed{ID: "feed-1"}}, &handlerContactsRepoStub{})
	h := NewSubscriptionHandler(subSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/feeds/feed-1/subscriptions", nil)
	req = withURLParam(req, "feedID", "feed-1")
	rec := httptest.NewRecorder()

	h.ListByFeed(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !subs.listByFeedCalled {
		t.Fatal("expected ListByFeed to be called")
	}
}

func TestContactHandlerTestSend(t *testing.T) {
	repo := &handlerContactsRepoStub{
		contactByID: domain.Contact{ID: "contact-http-1", Type: domain.ContactHTTP, Status: domain.ContactActive},
	}
	deliverySvc := service.NewContactDeliveryService(repo, handlerContactSenderStub{})
	h := NewContactHandler(service.NewContactService(repo), deliverySvc)

	body := []byte(`{"feed_name":"Test feed","title":"Hello","link":"https://example.com/test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/contacts/contact-http-1/test-send", bytes.NewReader(body))
	req = withURLParam(req, "contactID", "contact-http-1")
	rec := httptest.NewRecorder()

	h.TestSend(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGroupHandlerCreate(t *testing.T) {
	groupSvc := service.NewGroupService(&handlerGroupsRepoStub{group: domain.Group{ID: "group-1", Name: "test"}}, &handlerFeedRepoStub{}, &handlerContactsRepoStub{}, &handlerSubsRepoStub{})
	h := NewGroupHandler(groupSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader([]byte(`{"name":"test"}`)))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegistrationCodeHandlerCreate(t *testing.T) {
	codeSvc := service.NewRegistrationCodeService(&handlerRegistrationCodesRepoStub{}, &handlerGroupsRepoStub{group: domain.Group{ID: "group-1", Name: "g"}})
	h := NewRegistrationCodeHandler(codeSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/registration-codes", bytes.NewReader([]byte(`{"code":"ABC123","name":"Promo","enabled":true}`)))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
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
	created     domain.Feed
	getByIDFeed domain.Feed
}

func (r *handlerFeedRepoStub) Create(_ context.Context, f domain.Feed) (domain.Feed, error) {
	r.created = f
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
	contacts                []domain.Contact
	contactByID             domain.Contact
	createdHTTP             domain.Contact
	updatedHTTP             domain.Contact
	createdTelegram         domain.Contact
	updatedTelegram         domain.Contact
	createHTTPInput         domain.HTTPContactConfig
	createHTTPValue         string
	createHTTPName          *string
	createHTTPStatus        domain.ContactStatus
	updateHTTPInput         domain.HTTPContactConfig
	updateHTTPValue         string
	updateHTTPName          *string
	updateHTTPStatus        domain.ContactStatus
	updateHTTPContactID     string
	createTelegramChatID    string
	createTelegramUser      *string
	createTelegramName      *string
	createTelegramStatus    domain.ContactStatus
	updateTelegramChatID    string
	updateTelegramUser      *string
	updateTelegramName      *string
	updateTelegramStatus    domain.ContactStatus
	updateTelegramContactID string
	deletedContactID        string
}

func (r *handlerContactsRepoStub) UpsertTelegramActive(context.Context, string, *string, *string, time.Time) (domain.Contact, error) {
	return domain.Contact{}, nil
}

func (r *handlerContactsRepoStub) CreateTelegram(_ context.Context, chatID string, username *string, displayName *string, status domain.ContactStatus, _ *time.Time) (domain.Contact, error) {
	r.createTelegramChatID = chatID
	r.createTelegramUser = username
	r.createTelegramName = displayName
	r.createTelegramStatus = status
	return r.createdTelegram, nil
}

func (r *handlerContactsRepoStub) UpdateTelegram(_ context.Context, contactID string, chatID string, username *string, displayName *string, status domain.ContactStatus, _ *time.Time) (domain.Contact, error) {
	r.updateTelegramContactID = contactID
	r.updateTelegramChatID = chatID
	r.updateTelegramUser = username
	r.updateTelegramName = displayName
	r.updateTelegramStatus = status
	return r.updatedTelegram, nil
}

func (r *handlerContactsRepoStub) CreateHTTP(_ context.Context, value string, displayName *string, status domain.ContactStatus, cfg domain.HTTPContactConfig, _ *time.Time) (domain.Contact, error) {
	r.createHTTPValue = value
	r.createHTTPInput = cfg
	r.createHTTPName = displayName
	r.createHTTPStatus = status
	return r.createdHTTP, nil
}

func (r *handlerContactsRepoStub) UpdateHTTP(_ context.Context, contactID string, value string, displayName *string, status domain.ContactStatus, cfg domain.HTTPContactConfig, _ *time.Time) (domain.Contact, error) {
	r.updateHTTPContactID = contactID
	r.updateHTTPValue = value
	r.updateHTTPName = displayName
	r.updateHTTPStatus = status
	r.updateHTTPInput = cfg
	return r.updatedHTTP, nil
}

func (r *handlerContactsRepoStub) GetHTTPConfig(context.Context, string) (domain.HTTPContactConfig, error) {
	return domain.HTTPContactConfig{}, nil
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

func (r *handlerContactsRepoStub) Delete(_ context.Context, id string) error {
	r.deletedContactID = id
	return nil
}

type handlerSubsRepoStub struct {
	addCalled           bool
	listByFeed          []domain.Subscription
	listByFeedCalled    bool
	listByContact       []domain.Subscription
	listByContactCalled bool
}

func (r *handlerSubsRepoStub) ListByFeed(context.Context, string) ([]domain.Subscription, error) {
	r.listByFeedCalled = true
	return r.listByFeed, nil
}

func (r *handlerSubsRepoStub) ListByContact(context.Context, string) ([]domain.Subscription, error) {
	r.listByContactCalled = true
	return r.listByContact, nil
}

func (r *handlerSubsRepoStub) Add(context.Context, string, string) error {
	r.addCalled = true
	return nil
}

func (r *handlerSubsRepoStub) Remove(context.Context, string, string) error {
	return nil
}

func (r *handlerSubsRepoStub) AddGroup(context.Context, string, string, string) error {
	return nil
}

func (r *handlerSubsRepoStub) RemoveGroupByFeed(context.Context, string, string) error {
	return nil
}

func (r *handlerSubsRepoStub) RemoveGroupByContact(context.Context, string, string) error {
	return nil
}

type handlerContactSenderStub struct{}

func (handlerContactSenderStub) Send(context.Context, domain.Contact, domain.Feed, []domain.Item) error {
	return nil
}

type handlerGroupsRepoStub struct {
	group domain.Group
}

func (r *handlerGroupsRepoStub) Create(context.Context, domain.Group) (domain.Group, error) {
	return r.group, nil
}
func (r *handlerGroupsRepoStub) Update(context.Context, string, string, *string) (domain.Group, error) {
	return r.group, nil
}
func (r *handlerGroupsRepoStub) GetByID(context.Context, string) (domain.Group, error) {
	return r.group, nil
}
func (r *handlerGroupsRepoStub) List(context.Context, int, int) ([]domain.Group, int, error) {
	return []domain.Group{r.group}, 1, nil
}
func (r *handlerGroupsRepoStub) Delete(context.Context, string) error { return nil }
func (r *handlerGroupsRepoStub) ListContacts(context.Context, string) ([]domain.Contact, error) {
	return nil, nil
}
func (r *handlerGroupsRepoStub) AddContact(context.Context, string, string) error { return nil }
func (r *handlerGroupsRepoStub) RemoveContact(context.Context, string, string) error {
	return nil
}
func (r *handlerGroupsRepoStub) ListFeeds(context.Context, string) ([]domain.Feed, error) {
	return nil, nil
}
func (r *handlerGroupsRepoStub) AddFeed(context.Context, string, string) error    { return nil }
func (r *handlerGroupsRepoStub) RemoveFeed(context.Context, string, string) error { return nil }
func (r *handlerGroupsRepoStub) ListFeedIDs(context.Context, string) ([]string, error) {
	return nil, nil
}
func (r *handlerGroupsRepoStub) ListContactIDs(context.Context, string) ([]string, error) {
	return nil, nil
}

type handlerRegistrationCodesRepoStub struct{}

func (r *handlerRegistrationCodesRepoStub) Create(context.Context, domain.RegistrationCode) (domain.RegistrationCode, error) {
	return domain.RegistrationCode{ID: "code-1", Code: "ABC123", Name: "Promo", Enabled: true}, nil
}
func (r *handlerRegistrationCodesRepoStub) Update(context.Context, string, string, string, *string, bool, *int, *time.Time) (domain.RegistrationCode, error) {
	return domain.RegistrationCode{}, nil
}
func (r *handlerRegistrationCodesRepoStub) GetByID(context.Context, string) (domain.RegistrationCode, error) {
	return domain.RegistrationCode{ID: "code-1", Code: "ABC123", Name: "Promo", Enabled: true}, nil
}
func (r *handlerRegistrationCodesRepoStub) GetByCode(context.Context, string) (domain.RegistrationCode, error) {
	return domain.RegistrationCode{ID: "code-1", Code: "ABC123", Name: "Promo", Enabled: true}, nil
}
func (r *handlerRegistrationCodesRepoStub) List(context.Context, int, int) ([]domain.RegistrationCode, int, error) {
	return nil, 0, nil
}
func (r *handlerRegistrationCodesRepoStub) Delete(context.Context, string) error { return nil }
func (r *handlerRegistrationCodesRepoStub) ListGroups(context.Context, string) ([]domain.Group, error) {
	return nil, nil
}
func (r *handlerRegistrationCodesRepoStub) AddGroup(context.Context, string, string) error {
	return nil
}
func (r *handlerRegistrationCodesRepoStub) RemoveGroup(context.Context, string, string) error {
	return nil
}
func (r *handlerRegistrationCodesRepoStub) IncrementUse(context.Context, string) error { return nil }

func withURLParam(r *http.Request, key, value string) *http.Request {
	routeCtx := chi.RouteContext(r.Context())
	if routeCtx == nil {
		routeCtx = chi.NewRouteContext()
	}
	routeCtx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
}

func strPtr(v string) *string { return &v }
