package telegram

import (
	"strings"
	"testing"

	"github.com/111zxc/rss-communicator/internal/domain"
)

func TestRenderMessageSingleItem(t *testing.T) {
	msg := renderMessage(domain.Feed{Name: "Feed"}, []domain.Item{
		{Title: "Item 1", Link: "https://example.com/1"},
	})

	if msg != "📰 Item 1\nhttps://example.com/1" {
		t.Fatalf("unexpected single-item message: %q", msg)
	}
}

func TestRenderMessageBatch(t *testing.T) {
	msg := renderMessage(domain.Feed{Name: "Feed"}, []domain.Item{
		{Title: "Item 1", Link: "https://example.com/1"},
		{Title: "Item 2", Link: "https://example.com/2"},
	})

	for _, want := range []string{"📰 Feed", "2 new items", "1. Item 1", "2. Item 2"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("expected %q in message %q", want, msg)
		}
	}
}

func TestParseChatID(t *testing.T) {
	id, err := parseChatID("12345")
	if err != nil {
		t.Fatalf("parseChatID returned error: %v", err)
	}
	if id != 12345 {
		t.Fatalf("expected 12345, got %d", id)
	}
}
