package backoff

import (
	"context"
	"errors"
	"time"
)

const (
	multiplier = 2
	interval   = 1 * time.Second
)

var (
	errCanceled = errors.New("cancelled")
)

type Waiter interface {
	Wait() error
}

type Exponent struct {
	Peak time.Duration

	ctx context.Context
	n   time.Duration
}

func New(ctx context.Context) *Exponent {
	return &Exponent{
		ctx: ctx,
		n:   interval,
	}
}

func (p *Exponent) Wait() (err error) {
	n := p.n
	if p.Peak > 0 && n > p.Peak {
		n = p.Peak
	}
	select {
	case <-time.After(n):
	case <-p.ctx.Done():
		err = p.ctx.Err()
	}
	p.n *= multiplier
	return
}

type Broadcast struct {
	w    Waiter
	c    chan chan error
	quit chan error
}

func NewBroadcast(w Waiter) *Broadcast {
	b := &Broadcast{
		w:    w,
		c:    make(chan chan error, 100),
		quit: make(chan error, 1),
	}
	go b.loop()
	return b
}

func (b *Broadcast) loop() {
	var wc chan error
	var a []chan error
	for {
		select {
		case c := <-b.c:
			if wc == nil {
				wc = make(chan error, 1)
				go b.wait(wc)
			}
			a = append(a, c)
		case err := <-wc:
			for _, c := range a {
				if err != nil {
					c <- err
				}
				close(c)
			}
			a = nil
			wc = nil
		case err := <-b.quit:
			for _, c := range a {
				if err != nil {
					c <- err
				}
				close(c)
			}
			return
		}
	}
}

func (b *Broadcast) wait(c chan error) {
	if err := b.w.Wait(); err != nil {
		c <- err
	}
	close(c)
}

func (b *Broadcast) Wait() error {
	c := make(chan error, 1)
	b.c <- c
	return <-c
}

func (b *Broadcast) Cancel() {
	b.quit <- errCanceled
	close(b.quit)
}

func (b *Broadcast) Stop() {
	close(b.quit)
}
