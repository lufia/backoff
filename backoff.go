// Backoff implements exponential backoff algorithm.
package backoff

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

const (
	jitter          = 0.5
	multiplier      = 2
	defaultInterval = 1 * time.Second
)

var defaultRand *rand.Rand

func init() {
	defaultRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Backoff implements a variable for exponential backoff.
type Backoff struct {
	// Peak is maximum duration for Wait(). Zero is no limit.
	Peak time.Duration

	// Initial is initial duration of Wait().
	Initial time.Duration

	// Limit is maximum retry count.
	Limit int

	// MaxAge is maximum time until Backoff is force expired.
	MaxAge time.Duration

	n    int           // retry count
	d    time.Duration // most recent waiting time
	next time.Duration
	t    time.Time // time when Start was called
	once sync.Once
}

// Start starts timer for detecting that alive time of p is reached the MaxAge.
func (p *Backoff) Start() {
	p.once.Do(func() {
		p.t = time.Now()
	})
}

// SetNext sets next duration to d.
func (p *Backoff) SetNext(d time.Duration) {
	p.next = d
}

// weighted returns a duration in [d*0.5, d*1.5)
func weighted(d time.Duration) time.Duration {
	w := time.Duration(float64(d) * jitter)
	n := defaultRand.Int63n(w.Nanoseconds())
	return d - w + time.Duration(n)
}

// advance returns a duration between now and next period.
// this method don't consider p's limitations: Peak, Limit, or MaxAge.
func (p *Backoff) advance() time.Duration {
	if p.n == 0 {
		p.d = defaultInterval
		if p.Initial > 0 {
			p.d = p.Initial
		}
	} else {
		p.d *= multiplier
	}
	p.n++
	d := p.next
	p.next = 0
	if d > 0 {
		return d
	}
	return weighted(p.d)
}

var (
	// ErrLimitReached is the error returned when retrying count is reached to the limit.
	ErrLimitReached = errors.New("retry limit reached")

	// ErrExpired is the error returned when backoff's deadline is exceeded.
	ErrExpired = errors.New("operation is expired")
)

func (p *Backoff) age() time.Duration {
	return time.Since(p.t)
}

// Advance advances p's timer, returns a duration between now and next period.
// If p reaches any limits in current cycle, Advance returns an error.
func (p *Backoff) Advance() (time.Duration, error) {
	p.Start()
	if p.Limit > 0 && p.n >= p.Limit {
		return 0, ErrLimitReached
	}
	d := p.advance()
	if p.Peak > 0 && d > p.Peak {
		d = p.Peak
	}
	age := p.age() + d
	if p.MaxAge > 0 && age >= p.MaxAge {
		return 0, ErrExpired
	}
	return d, nil
}

// Wait blocks until next activation available, or ctx is cancelled.
func (p *Backoff) Wait(ctx context.Context) error {
	d, err := p.Advance()
	if err != nil {
		return err
	}
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
