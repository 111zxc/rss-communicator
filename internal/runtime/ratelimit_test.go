package runtime

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucketWaitWithoutLimit(t *testing.T) {
	b := NewTokenBucket(0, 0)
	if err := b.Wait(context.Background()); err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
}

func TestTokenBucketWaitRespectsContext(t *testing.T) {
	b := NewTokenBucket(1, 1)
	if err := b.Wait(context.Background()); err != nil {
		t.Fatalf("initial Wait returned error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := b.Wait(ctx); err == nil {
		t.Fatal("expected context deadline error")
	}
}
