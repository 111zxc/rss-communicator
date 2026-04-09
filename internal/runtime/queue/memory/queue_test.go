package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/111zxc/rss-communicator/internal/runtime/queue"
)

func TestMemQueuePublishFanout(t *testing.T) {
	q := New()
	defer func() { _ = q.Close() }()

	var wg sync.WaitGroup
	wg.Add(2)
	received := make(chan string, 2)

	handler := func(_ context.Context, m queue.Message) error {
		received <- string(m.Data)
		wg.Done()
		return nil
	}

	if err := q.Subscribe(context.Background(), queue.TopicFetch, handler); err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}
	if err := q.Subscribe(context.Background(), queue.TopicFetch, handler); err != nil {
		t.Fatalf("Subscribe returned error: %v", err)
	}

	if err := q.Publish(context.Background(), queue.TopicFetch, []byte("job")); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for handlers")
	}

	if len(received) != 2 {
		t.Fatalf("expected both subscribers to receive the message, got %d", len(received))
	}
}

func TestMemQueueClosePreventsFurtherUse(t *testing.T) {
	q := New()
	if err := q.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if err := q.Publish(context.Background(), queue.TopicFetch, []byte("job")); err == nil {
		t.Fatal("expected publish on closed queue to fail")
	}
}
