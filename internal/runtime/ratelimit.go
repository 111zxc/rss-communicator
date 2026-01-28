package runtime

import (
	"context"
	"time"
)

type TokenBucket struct {
	ch chan struct{}
}

func NewTokenBucket(rps float64, burst int) *TokenBucket {
	if rps <= 0 {
		return &TokenBucket{ch: nil}
	}
	if burst < 1 {
		burst = 1
	}
	ch := make(chan struct{}, burst)

	for i := 0; i < burst; i++ {
		ch <- struct{}{}
	}

	interval := time.Duration(float64(time.Second) / rps)
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for range ticker.C {
			select {
			case ch <- struct{}{}:
			default:
				// bucket full
			}
		}
	}()

	return &TokenBucket{ch: ch}
}

func (b *TokenBucket) Wait(ctx context.Context) error {
	if b.ch == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.ch:
		return nil
	}
}
