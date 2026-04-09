package rss

import "testing"

func TestParseRSSItems(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Example Feed</title>
    <item>
      <guid>item-1</guid>
      <title>First item</title>
      <link>https://example.com/first</link>
      <description>Summary</description>
      <author>author@example.com (Alice)</author>
      <pubDate>Tue, 09 Apr 2026 12:00:00 GMT</pubDate>
    </item>
  </channel>
</rss>`)

	items, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ExternalID == nil || *items[0].ExternalID != "item-1" {
		t.Fatalf("unexpected external ID: %+v", items[0].ExternalID)
	}
	if items[0].Title != "First item" || items[0].Link != "https://example.com/first" {
		t.Fatalf("unexpected parsed item: %+v", items[0])
	}
	if items[0].PublishedAt == nil {
		t.Fatal("expected PublishedAt to be set")
	}
	if items[0].UniqKey == "" {
		t.Fatal("expected uniq key to be set")
	}
}

func TestParseInvalidFeed(t *testing.T) {
	if _, err := Parse([]byte("<rss>")); err == nil {
		t.Fatal("expected parse error for invalid feed")
	}
}
