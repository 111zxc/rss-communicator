package nats

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/111zxc/rss-communicator/internal/runtime/queue"
)

type Config struct {
	Stream      string
	SubjectRoot string
	AckWait     time.Duration
}

type Queue struct {
	conn        *nats.Conn
	js          nats.JetStreamContext
	stream      string
	subjectRoot string
	ackWait     time.Duration

	mu     sync.Mutex
	subs   []*nats.Subscription
	closed bool
}

func New(url string, cfg Config) (*Queue, error) {
	if strings.TrimSpace(url) == "" {
		url = nats.DefaultURL
	}
	if strings.TrimSpace(cfg.Stream) == "" {
		cfg.Stream = "RSS_COMMUNICATOR"
	}
	if strings.TrimSpace(cfg.SubjectRoot) == "" {
		cfg.SubjectRoot = "rss"
	}
	if cfg.AckWait <= 0 {
		cfg.AckWait = 5 * time.Minute
	}

	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if _, err := js.AddStream(&nats.StreamConfig{
		Name:      cfg.Stream,
		Subjects:  []string{cfg.SubjectRoot + ".>"},
		Retention: nats.WorkQueuePolicy,
		Storage:   nats.FileStorage,
	}); err != nil && !errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
		var apiErr *nats.APIError
		if !errors.As(err, &apiErr) || apiErr.ErrorCode != 10058 {
			conn.Close()
			return nil, err
		}
	}

	return &Queue{
		conn:        conn,
		js:          js,
		stream:      cfg.Stream,
		subjectRoot: cfg.SubjectRoot,
		ackWait:     cfg.AckWait,
	}, nil
}

func (q *Queue) Publish(_ context.Context, topic queue.Topic, data []byte) error {
	q.mu.Lock()
	closed := q.closed
	q.mu.Unlock()
	if closed {
		return errors.New("queue closed")
	}

	_, err := q.js.Publish(q.subject(topic), data)
	return err
}

func (q *Queue) Subscribe(ctx context.Context, topic queue.Topic, handler func(context.Context, queue.Message) error) error {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return errors.New("queue closed")
	}
	q.mu.Unlock()

	sub, err := q.js.QueueSubscribe(
		q.subject(topic),
		q.groupName(topic),
		func(msg *nats.Msg) {
			err := handler(ctx, queue.Message{
				Topic: topic,
				Data:  msg.Data,
			})
			if err == nil {
				_ = msg.Ack()
				return
			}
			_ = msg.Nak()
		},
		nats.BindStream(q.stream),
		nats.Durable(q.groupName(topic)),
		nats.ManualAck(),
		nats.AckWait(q.ackWait),
		nats.DeliverAll(),
	)
	if err != nil {
		return err
	}

	q.mu.Lock()
	q.subs = append(q.subs, sub)
	q.mu.Unlock()

	go func() {
		<-ctx.Done()
		_ = sub.Drain()
	}()

	return nil
}

func (q *Queue) Close() error {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return nil
	}
	q.closed = true
	subs := append([]*nats.Subscription(nil), q.subs...)
	q.mu.Unlock()

	for _, sub := range subs {
		_ = sub.Drain()
	}
	q.conn.Close()
	return nil
}

func (q *Queue) subject(topic queue.Topic) string {
	return q.subjectRoot + "." + string(topic)
}

func (q *Queue) groupName(topic queue.Topic) string {
	return fmt.Sprintf("%s-%s-workers", strings.ToLower(q.stream), string(topic))
}
