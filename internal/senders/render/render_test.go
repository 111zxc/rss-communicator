package render

import (
	"strings"
	"testing"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

func TestMessageSingleItem(t *testing.T) {
	msg := Message(domain.Feed{Name: "Feed"}, []domain.Item{
		{Title: "Item 1", Link: "https://example.com/1"},
	})

	if msg != "📰 Item 1\nhttps://example.com/1" {
		t.Fatalf("unexpected single-item message: %q", msg)
	}
}

func TestMessageBatch(t *testing.T) {
	msg := Message(domain.Feed{Name: "Feed"}, []domain.Item{
		{Title: "Item 1", Link: "https://example.com/1"},
		{Title: "Item 2", Link: "https://example.com/2"},
	})

	for _, want := range []string{"📰 Feed", "2 new items", "1. Item 1", "2. Item 2"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("expected %q in message %q", want, msg)
		}
	}
}

func TestTemplateSupportsJSONAndItems(t *testing.T) {
	publishedAt := time.Date(2026, 4, 13, 10, 0, 0, 0, time.UTC)
	got := Template(
		`{"feed": {json_feed_name}, "text": {json_text}, "items": {items_json}, "first_title": {json_title}}`,
		domain.Feed{Name: "Feed"},
		[]domain.Item{{Title: `Quote "test"`, Link: "https://example.com/1", PublishedAt: &publishedAt}},
	)

	for _, want := range []string{`"Feed"`, `Quote \"test\"`, `"items": [`} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in rendered template %q", want, got)
		}
	}
}
