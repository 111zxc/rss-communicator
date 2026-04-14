package queuefactory

import (
	"fmt"
	"strings"
	"time"

	"github.com/111zxc/rss-communicator/internal/runtime/queue"
	"github.com/111zxc/rss-communicator/internal/runtime/queue/memory"
	natsqueue "github.com/111zxc/rss-communicator/internal/runtime/queue/nats"
)

type NATSConfig struct {
	URL         string
	Stream      string
	SubjectRoot string
	AckWait     time.Duration
}

func Open(driver string, cfg NATSConfig) (queue.Queue, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "memory", "inmemory", "in-proc", "inproc":
		return memory.New(), nil
	case "nats", "jetstream", "nats-jetstream":
		return natsqueue.New(cfg.URL, natsqueue.Config{
			Stream:      cfg.Stream,
			SubjectRoot: cfg.SubjectRoot,
			AckWait:     cfg.AckWait,
		})
	default:
		return nil, fmt.Errorf("unsupported queue driver %q", driver)
	}
}
