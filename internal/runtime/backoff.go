package runtime

import (
	"math/rand"
	"time"
)

type Backoff struct {
	Base time.Duration
	Max  time.Duration
}

func (b Backoff) Delay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	// экспонента: Base * 2^attempt
	d := b.Base << attempt
	if d > b.Max {
		d = b.Max
	}
	// джиттер 0.8..1.2
	j := 0.8 + rand.Float64()*0.4
	return time.Duration(float64(d) * j)
}
