package backoff

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

const (
	tAccuracy = 10 * time.Millisecond
)

func TestWait(t *testing.T) {
	tab := [][]time.Duration{
		{interval},
		{interval, interval * multiplier},
		{interval, interval * multiplier, interval * multiplier * 2},
	}
	for _, a := range tab {
		w := New(context.Background())
		for i, d := range a {
			t0 := time.Now()
			if err := w.Wait(); err != nil {
				t.Fatalf("Wait() = %v", err)
			}
			t1 := time.Now()
			exp := t0.Add(d)
			if !testAfterAccuracy(t1, exp, tAccuracy) {
				t.Errorf("try %d: watis %v; but want %v", i+1, t1.Sub(t0), exp.Sub(t0))
			}
		}
	}
}

func TestWaitWithPeak(t *testing.T) {
	const peak = interval + interval/2
	tab := [][]time.Duration{
		{interval},
		{interval, peak},
		{interval, peak, peak},
	}
	for _, a := range tab {
		w := New(context.Background())
		w.Peak = peak
		for i, d := range a {
			t0 := time.Now()
			if err := w.Wait(); err != nil {
				t.Fatalf("Wait() = %v", err)
			}
			t1 := time.Now()
			exp := t0.Add(d)
			if !testAfterAccuracy(t1, exp, tAccuracy) {
				t.Errorf("try %d: watis %v; but want %v", i+1, t1.Sub(t0), exp.Sub(t0))
			}
		}
	}
}

func TestWaitDeadline(t *testing.T) {
	t0 := time.Now()
	exp := t0.Add(interval / 2)

	ctx, cancel := context.WithDeadline(context.Background(), exp)
	w := New(ctx)
	err := w.Wait()
	if err == nil {
		t.Errorf("Wait() = nil; but want an error")
	}
	t1 := time.Now()
	if !testAfterAccuracy(t1, exp, tAccuracy) {
		t.Errorf("watis %v; but want %v", t1.Sub(t0), exp.Sub(t0))
	}
	cancel()
}

func testAfterAccuracy(actual time.Time, expect time.Time, d time.Duration) bool {
	return actual.After(expect) && actual.Sub(expect) <= d
}

type tWaiter struct {
	c chan error
}

func (w *tWaiter) Wait() error {
	return <-w.c
}

func TestBroadcastCancel(t *testing.T) {
	tab := []int{0, 1, 2, 10}
	for _, v := range tab {
		var w tWaiter
		b := NewBroadcast(&w)
		c := make(chan error, v)
		t.Run(fmt.Sprintf("N=%d", v), func(t *testing.T) {
			var wg sync.WaitGroup
			for i := 0; i < v; i++ {
				wg.Add(1)
				go func() {
					c <- b.Wait()
					wg.Done()
				}()
			}
			go func() {
				wg.Wait()
				close(c)
			}()
			<-time.After(1 * time.Millisecond) // HACK
			b.Cancel()

			var n int
			for err := range c {
				if err != nil {
					n++
				}
			}
			if n != v {
				t.Errorf("number of respond error = %d; want %d", n, v)
			}
		})
	}
}
