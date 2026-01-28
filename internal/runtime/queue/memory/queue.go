package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/111zxc/rss-communicator/internal/runtime/queue"
)

type MemQueue struct {
	mu     sync.RWMutex
	subs   map[queue.Topic][]func(context.Context, queue.Message) error
	closed bool
}

func New() *MemQueue {
	return &MemQueue{subs: make(map[queue.Topic][]func(context.Context, queue.Message) error)}
}

func (q *MemQueue) Publish(ctx context.Context, topic queue.Topic, data []byte) error {
	q.mu.RLock()
	if q.closed {
		q.mu.RUnlock()
		return errors.New("queue closed")
	}
	handlers := append([]func(context.Context, queue.Message) error(nil), q.subs[topic]...)
	q.mu.RUnlock()

	msg := queue.Message{Topic: topic, Data: data}

	for _, h := range handlers {
		hh := h
		go func() { _ = hh(ctx, msg) }()
	}
	return nil
}

func (q *MemQueue) Subscribe(ctx context.Context, topic queue.Topic, handler func(context.Context, queue.Message) error) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return errors.New("queue closed")
	}
	q.subs[topic] = append(q.subs[topic], handler)
	return nil
}

func (q *MemQueue) Close() error {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	return nil
}
