package render

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/111zxc/rss-communicator/internal/domain"
)

func Message(feed domain.Feed, items []domain.Item) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) == 1 {
		return fmt.Sprintf("📰 %s\n%s", items[0].Title, items[0].Link)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📰 %s\n%d new items\n\n", feed.Name, len(items))
	for i, item := range items {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "%d. %s\n%s", i+1, item.Title, item.Link)
	}
	return b.String()
}

func Template(input string, feed domain.Feed, items []domain.Item) string {
	if input == "" {
		return ""
	}

	first := domain.Item{}
	if len(items) > 0 {
		first = items[0]
	}

	values := map[string]string{
		"{feed_id}":           feed.ID,
		"{feed_name}":         feed.Name,
		"{feed_url}":          feed.URL,
		"{item_count}":        fmt.Sprintf("%d", len(items)),
		"{text}":              Message(feed, items),
		"{json_feed_id}":      jsonString(feed.ID),
		"{json_feed_name}":    jsonString(feed.Name),
		"{json_feed_url}":     jsonString(feed.URL),
		"{json_item_count}":   fmt.Sprintf("%d", len(items)),
		"{json_text}":         jsonString(Message(feed, items)),
		"{title}":             first.Title,
		"{link}":              first.Link,
		"{summary}":           deref(first.Summary),
		"{author}":            deref(first.Author),
		"{published_at}":      formatTime(first.PublishedAt),
		"{json_title}":        jsonString(first.Title),
		"{json_link}":         jsonString(first.Link),
		"{json_summary}":      jsonString(deref(first.Summary)),
		"{json_author}":       jsonString(deref(first.Author)),
		"{json_published_at}": jsonString(formatTime(first.PublishedAt)),
		"{items_json}":        itemsJSON(items),
	}

	out := input
	for token, value := range values {
		out = strings.ReplaceAll(out, token, value)
	}
	return out
}

func jsonString(v string) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func itemsJSON(items []domain.Item) string {
	type payloadItem struct {
		Title       string  `json:"title"`
		Link        string  `json:"link"`
		Summary     *string `json:"summary,omitempty"`
		Author      *string `json:"author,omitempty"`
		PublishedAt *string `json:"published_at,omitempty"`
	}

	payload := make([]payloadItem, 0, len(items))
	for _, item := range items {
		var publishedAt *string
		if item.PublishedAt != nil {
			s := item.PublishedAt.UTC().Format(timeLayout)
			publishedAt = &s
		}
		payload = append(payload, payloadItem{
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			Author:      item.Author,
			PublishedAt: publishedAt,
		})
	}

	b, _ := json.Marshal(payload)
	return string(b)
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

const timeLayout = "2006-01-02T15:04:05Z07:00"

func formatTime(v *time.Time) string {
	if v == nil {
		return ""
	}
	return v.UTC().Format(timeLayout)
}
