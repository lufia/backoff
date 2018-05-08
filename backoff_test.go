package backoff

import (
	"context"
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
