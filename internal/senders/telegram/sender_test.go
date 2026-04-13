package telegram

import (
	"testing"

	"github.com/111zxc/rss-communicator/internal/domain"
	"github.com/111zxc/rss-communicator/internal/senders/render"
)

func TestRenderMessageSingleItem(t *testing.T) {
	msg := render.Message(domain.Feed{Name: "Feed"}, []domain.Item{
		{Title: "Item 1", Link: "https://example.com/1"},
	})

	if msg != "📰 Item 1\nhttps://example.com/1" {
		t.Fatalf("unexpected single-item message: %q", msg)
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
