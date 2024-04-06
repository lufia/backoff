package backoff

import (
	"context"
	"errors"
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

func (r *tRange) weighted() (time.Duration, time.Duration) {
	w := time.Duration(float64(r.Begin) * jitter)
	return r.Begin - w, r.End + w
}

func (r *tRange) String() string {
	bp, ep := r.weighted()
	return fmt.Sprintf("%v..%v", bp, ep)
}

func (r *tRange) In(d time.Duration) bool {
	bp, ep := r.weighted()
	return d >= bp && d < ep
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
	tests := [][]time.Duration{
		{tInterval},
		{tInterval, tInterval * multiplier},
		{tInterval, tInterval * multiplier, tInterval * multiplier * 2},
	}
	ctx := context.Background()
	for _, tt := range tests {
		var w Backoff
		w.Initial = tInterval
		for _, d := range tt {
			testWait(t, ctx, &w, d)
		}
	}
}

func TestWaitWithPeak(t *testing.T) {
	const peak = tInterval + tInterval/2
	tests := [][]time.Duration{
		{tInterval},
		{tInterval, peak},
		{tInterval, peak, peak},
	}
	ctx := context.Background()
	for _, tt := range tests {
		var w Backoff
		w.Initial = tInterval
		w.Peak = peak
		for _, d := range tt {
			testWait(t, ctx, &w, d)
		}
	}
}

func TestWaitNext(t *testing.T) {
	var w Backoff
	w.Initial = 10 * time.Millisecond
	ctx := context.Background()

	tests := []struct {
		Next time.Duration
		Want time.Duration
	}{
		{Want: w.Initial},
		{Want: 5 * time.Millisecond, Next: 5 * time.Millisecond},
		{Want: w.Initial * 4},
	}
	for _, v := range tests {
		if v.Next > 0 {
			w.SetNext(v.Next)
		}
		testWait(t, ctx, &w, v.Want)
	}
}

func TestWaitDeadline(t *testing.T) {
	t0 := time.Now()
	timeout := tInterval / 4

	var w Backoff
	w.Initial = tInterval
	ctx, cancel := context.WithDeadline(context.Background(), t0.Add(timeout))
	defer cancel()

	err := w.Wait(ctx)
	if err == nil {
		t.Errorf("Wait(ctx) must return an error that mean deadline reached")
	}
	d := time.Since(t0)
	r := &tRange{Begin: timeout, End: timeout + tAccuracy}
	if !r.In(d) {
		t.Errorf("Wait(%v): ellapsed %v; want %v", timeout, d, r)
	}
}

func TestAdvanceMaxAge(t *testing.T) {
	w := Backoff{
		Initial: tInterval,
		MaxAge:  tInterval * 3,
	}
	time.Sleep(tInterval)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := w.Wait(ctx); err != nil {
			return
		}
	}
	t.Errorf("Age = %v, MaxAge = %v; want %v", w.age(), w.MaxAge, ErrExpired)
}

func TestStartMaxAge(t *testing.T) {
	w := Backoff{
		Initial: tInterval,
		MaxAge:  tInterval * 3,
	}
	w.Start()
	time.Sleep(tInterval)
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if err := w.Wait(ctx); err != nil {
			return
		}
		time.Sleep(tInterval / 10) // keep this for safety
	}
	t.Errorf("Age = %v, MaxAge = %v; want %v", w.age(), w.MaxAge, ErrExpired)
}

func Example() {
	// retryable function
	f := func(i int) error {
		if i < 2 {
			return errors.New("fail")
		}
		return nil
	}

	// w is backoff [1s, 2s, 4s, 8s, 16s, ...]
	var w Backoff
	var i int
	for err := f(i); err != nil; err = f(i) {
		if err := w.Wait(context.Background()); err != nil {
			fmt.Println(err)
		}
		i++
	}
	// Output:
}

func Example_limited() {
	// retryable function
	f := func() error {
		return errors.New("fail")
	}

	// w is backoff [100ms, 200ms, 200ms, 200ms, 200ms]
	w := Backoff{
		Initial: 100 * time.Millisecond,
		Peak:    200 * time.Millisecond,
		Limit:   5,
		MaxAge:  2 * time.Second,
	}
	for err := f(); err != nil; err = f() {
		if err := w.Wait(context.Background()); err != nil {
			fmt.Println(err)
			break
		}
	}
	// Output:
	// retry limit reached
}
