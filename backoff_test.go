package backoff

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"
)

const (
	tInterval = 10 * time.Millisecond
)

func testWait(t *testing.T, c context.Context, w *Backoff, v time.Duration) {
	t.Helper()
	t0 := time.Now()
	if err := w.Wait(c); err != nil {
		t.Fatalf("failed to Wait(): %v", err)
	}
	bp := time.Duration(float64(v) * jitter)
	ep := bp + v
	if d := time.Since(t0); d < bp || d >= ep {
		t.Errorf("Wait(): ellapsed %v; want %v±%v", d, v, v/2)
	}
}

func TestWaitDefaultInterval(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var w Backoff
		testWait(t, t.Context(), &w, defaultInterval)
	})
}

func TestWait(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		tests := [][]time.Duration{
			{tInterval},
			{tInterval, tInterval * multiplier},
			{tInterval, tInterval * multiplier, tInterval * multiplier * 2},
		}
		for _, tt := range tests {
			var w Backoff
			w.Initial = tInterval
			for _, d := range tt {
				testWait(t, t.Context(), &w, d)
			}
		}
	})
}

func TestWaitWithPeak(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const peak = tInterval + tInterval/2
		tests := [][]time.Duration{
			{tInterval},
			{tInterval, peak},
			{tInterval, peak, peak},
		}
		for _, tt := range tests {
			var w Backoff
			w.Initial = tInterval
			w.Peak = peak
			for _, d := range tt {
				testWait(t, t.Context(), &w, d)
			}
		}
	})
}

func TestWaitNext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var w Backoff
		w.Initial = 10 * time.Millisecond

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
			testWait(t, t.Context(), &w, v.Want)
		}
	})
}

func TestWaitDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		t0 := time.Now()
		timeout := 100 * time.Millisecond

		var w Backoff
		w.Initial = 500 * time.Millisecond
		ctx, cancel := context.WithDeadline(t.Context(), t0.Add(timeout))
		t.Cleanup(cancel)

		if err := w.Wait(ctx); err == nil || !errors.Is(err, ctx.Err()) {
			t.Errorf("Wait(ctx) must return an error that mean deadline reached")
		}
		d := time.Since(t0)
		if d != timeout {
			t.Errorf("Wait() should cancel right now when ctx was cancelled: ellapsed %v; want %v", d, timeout)
		}
	})
}

func TestAdvanceMaxAge(t *testing.T) {
	// Wait should return ErrExpired when due to reached for MaxAge.
	synctest.Test(t, func(t *testing.T) {
		w := Backoff{
			Initial: tInterval,
			MaxAge:  tInterval * 3,
		}
		time.Sleep(tInterval)
		for i := 0; i < 3; i++ {
			if err := w.Wait(t.Context()); !errors.Is(err, ErrExpired) {
				return
			}
		}
		t.Errorf("Age = %v, MaxAge = %v; want %v", w.age(), w.MaxAge, ErrExpired)
	})
}

func TestStartMaxAge(t *testing.T) {
	// Wait should return ErrExpired when due to reached for MaxAge; it starts of Start().
	synctest.Test(t, func(t *testing.T) {
		w := Backoff{
			Initial: tInterval,
			MaxAge:  tInterval * 3,
		}
		w.Start()
		time.Sleep(tInterval)
		for i := 0; i < 2; i++ {
			if err := w.Wait(t.Context()); !errors.Is(err, ErrExpired) {
				return
			}
		}
		t.Errorf("Age = %v, MaxAge = %v; want %v", w.age(), w.MaxAge, ErrExpired)
	})
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
