package rss

import (
	"testing"
	"time"
)

func TestComputeUniqKeyDeterministic(t *testing.T) {
	pub := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)

	a := ComputeUniqKey("title", "https://example.com/1", &pub)
	b := ComputeUniqKey("title", "https://example.com/1", &pub)
	c := ComputeUniqKey("title", "https://example.com/2", &pub)

	if a != b {
		t.Fatal("expected identical input to produce identical hash")
	}
	if a == c {
		t.Fatal("expected different input to produce different hash")
	}
}
