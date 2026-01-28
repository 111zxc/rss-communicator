package queue

import "context"

type Topic string

const (
	TopicFetch   Topic = "fetch"
	TopicDeliver Topic = "deliver"
)

type Message struct {
	Topic Topic
	Data  []byte
}

type Queue interface {
	Publish(ctx context.Context, topic Topic, data []byte) error
	Subscribe(ctx context.Context, topic Topic, handler func(context.Context, Message) error) error
	Close() error
}
