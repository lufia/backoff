// Backoff implements exponental backoff algorithm.
package backoff

import (
	"context"
	"time"
)

const (
	multiplier      = 2
	defaultInterval = 1 * time.Second
)

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
	if p.Next > 0 {
		return p.Next
	}
	return p.d
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
