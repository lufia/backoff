package backoff

import (
	"context"
	"fmt"
	"testing"
	"time"
)

const (
	tAccuracy = 10 * time.Millisecond
	tInterval = 10 * time.Millisecond
)

type tRange struct {
	Begin time.Duration
	End   time.Duration
}

func (r *tRange) String() string {
	return fmt.Sprintf("%v..%v", r.Begin, r.End)
}

func (r *tRange) In(d time.Duration) bool {
	return d >= r.Begin && d <= r.End
}

func testWait(t *testing.T, c context.Context, w *Backoff, v time.Duration) {
	t.Helper()
	t0 := time.Now()
	if err := w.Wait(c); err != nil {
		t.Fatalf("failed to Wait(): %v", err)
	}
	d := time.Since(t0)
	r := &tRange{Begin: v, End: v + tAccuracy}
	if !r.In(d) {
		t.Errorf("Wait(): ellapsed %v; want %v", d, r)
	}
}

func TestWaitDefaultInterval(t *testing.T) {
	var w Backoff
	testWait(t, context.Background(), &w, defaultInterval)
}

func TestWait(t *testing.T) {
	tab := [][]time.Duration{
		{tInterval},
		{tInterval, tInterval * multiplier},
		{tInterval, tInterval * multiplier, tInterval * multiplier * 2},
	}
	c := context.Background()
	for _, a := range tab {
		var w Backoff
		w.Initial = tInterval
		for _, d := range a {
			testWait(t, c, &w, d)
		}
	}
}

func TestWaitWithPeak(t *testing.T) {
	const peak = tInterval + tInterval/2
	tab := [][]time.Duration{
		{tInterval},
		{tInterval, peak},
		{tInterval, peak, peak},
	}
	c := context.Background()
	for _, a := range tab {
		var w Backoff
		w.Initial = tInterval
		w.Peak = peak
		for _, d := range a {
			testWait(t, c, &w, d)
		}
	}
}

func TestWaitNext(t *testing.T) {
	var w Backoff
	w.Initial = 10 * time.Millisecond
	c := context.Background()

	tab := []struct {
		Next time.Duration
		Want time.Duration
	}{
		{Want: w.Initial},
		{Want: 5 * time.Millisecond, Next: 5 * time.Millisecond},
		{Want: w.Initial * 4},
	}
	for _, v := range tab {
		if v.Next > 0 {
			w.Next = v.Next
		}
		testWait(t, c, &w, v.Want)
	}
}

func TestWaitDeadline(t *testing.T) {
	t0 := time.Now()
	timeout := tInterval / 2

	var w Backoff
	w.Initial = tInterval
	ctx, cancel := context.WithDeadline(context.Background(), t0.Add(timeout))
	defer cancel()

	err := w.Wait(ctx)
	if err == nil {
		t.Errorf("Wait(ctx) must return an error that is deadline reached")
	}
	d := time.Since(t0)
	r := &tRange{Begin: timeout, End: timeout + tAccuracy}
	if !r.In(d) {
		t.Errorf("Wait(%v): ellapsed %v; want %v", timeout, d, r)
	}
}
