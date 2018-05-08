// Backoff implements exponental backoff algorithm.
package backoff

import (
	"context"
	"math/rand"
	"time"
)

const (
	weightDiv       = 2
	multiplier      = 2
	defaultInterval = 1 * time.Second
)

var defaultRand *rand.Rand

func init() {
	defaultRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Backoff implements a variable for exponental backoff.
type Backoff struct {
	// Peak is maximum duration for Wait(). Zero is no limit.
	Peak time.Duration

	// Initial is initial duration of Wait().
	Initial time.Duration

	// Next is used to Wait() when is called if this is greater than zero.
	// It will reset to zero after consumed.
	Next time.Duration

	Limit int           // not implemented
	Age   time.Duration // not implemented

	n int
	d time.Duration
}

// weighted returns a duration in [d*0.5, d*1.5)
func weighted(d time.Duration) time.Duration {
	// must: d >= weightDiv
	w := d / weightDiv
	n := defaultRand.Int63n(w.Nanoseconds())
	return d - w + time.Duration(n)
}

// advance advances p's timers, then returns next duration of Wait().
// this method don't consider p's limitations: Peak, Limit, or Age.
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
	d := p.Next
	p.Next = 0
	if d > 0 {
		return d
	}
	return weighted(p.d)
}

// Wait waits next activation.
func (p *Backoff) Wait(ctx context.Context) error {
	d := p.advance()
	if p.Peak > 0 && d > p.Peak {
		d = p.Peak
	}
	select {
	case <-time.After(d):
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
