package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/runtime/worker"
)

func TestSenderSendRendersRequest(t *testing.T) {
	var gotMethod, gotPath, gotAuth, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotBody = string(body)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	bodyTemplate := `{"text": {json_text}, "items": {items_json}}`
	sender := New(&contactRepoStub{}, srv.Client())
	err := sender.Send(context.Background(), domain.Contact{
		ID:     "contact-1",
		Type:   domain.ContactHTTP,
		Status: domain.ContactActive,
		HTTP: &domain.HTTPContactConfig{
			Method:       "POST",
			URL:          srv.URL + "/hooks/{feed_name}",
			Headers:      map[string]string{"Authorization": "Bearer {feed_id}"},
			BodyTemplate: &bodyTemplate,
		},
	}, domain.Feed{ID: "feed-1", Name: "news"}, []domain.Item{{Title: "Item 1", Link: "https://example.com/1"}})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/hooks/news" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotAuth != "Bearer feed-1" {
		t.Fatalf("unexpected auth header: %s", gotAuth)
	}
	if !strings.Contains(gotBody, `"text": "📰 Item 1\nhttps://example.com/1"`) {
		t.Fatalf("unexpected body: %s", gotBody)
	}
}

func TestSenderSendMarksClientErrorsPermanent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	sender := New(&contactRepoStub{}, srv.Client())
	err := sender.Send(context.Background(), domain.Contact{
		ID:   "contact-1",
		Type: domain.ContactHTTP,
		HTTP: &domain.HTTPContactConfig{
			Method:  "POST",
			URL:     srv.URL,
			Headers: map[string]string{},
		},
	}, domain.Feed{}, []domain.Item{{Title: "Item 1", Link: "https://example.com/1"}})
	if err == nil {
		t.Fatal("expected error")
	}

	var permanent *worker.PermanentError
	if !errors.As(err, &permanent) {
		t.Fatalf("expected PermanentError, got %T", err)
	}
}

type contactRepoStub struct{}

func (contactRepoStub) GetHTTPConfig(context.Context, string) (domain.HTTPContactConfig, error) {
	return domain.HTTPContactConfig{}, nil
}
