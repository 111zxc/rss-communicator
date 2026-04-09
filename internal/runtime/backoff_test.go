package runtime

import (
	"testing"
	"time"
)

func TestBackoffDelayUsesBaseAndJitterBounds(t *testing.T) {
	b := Backoff{Base: 10 * time.Second, Max: 2 * time.Minute}
	d := b.Delay(0)

	if d < 8*time.Second || d > 12*time.Second {
		t.Fatalf("expected jittered base delay within bounds, got %v", d)
	}
}

func TestBackoffDelayCapsAtMax(t *testing.T) {
	b := Backoff{Base: 10 * time.Second, Max: 30 * time.Second}
	d := b.Delay(10)

	if d < 24*time.Second || d > 36*time.Second {
		t.Fatalf("expected capped delay around max, got %v", d)
	}
}
