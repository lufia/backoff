# backoff

Backoff implements exponental backoff algorithm in Go.

[![GoDoc](https://godoc.org/github.com/lufia/backoff?status.svg)](https://godoc.org/github.com/lufia/backoff)
[![Build Status](https://travis-ci.org/lufia/backoff.svg?branch=master)](https://travis-ci.org/lufia/backoff)

## DESCRIPTION

## EXAMPLE

This is simple example.
`w.Wait` blocks until next period that are:

1. 1±0.5s
2. 2±1s
3. 4±2s
4. 8±4s

It increases exponentally forever.

```go
import (
	"context"
	"log"

	"github.com/lufia/backoff"
)

func main() {
	var w backoff.Backoff
	for {
		if err := w.Wait(context.Background()); err != nil {
			log.Fatalln(err)
		}
	}
}
```

Next example has limits for increase.
`w.Wait` blocks like above example but it is limited by `Initial`, `Peak`, and `Limit`.
Therefore periods are:

1. 0.5±0.25s
2. 1±0.5s
3. 2±1s
4. 2±1s
5. 2±1s

`w.Wait` ends up with an error.

```go
import (
	"context"
	"log"
	"time"

	"github.com/lufia/backoff"
)

func main() {
	w := backoff.Backoff{
		Initial: 500 * time.Millisecond,
		Peak:    2 * time.Second,
		Limit:   5,
	}
	for {
		if err := w.Wait(context.Background()); err != nil {
			log.Fatalln(err)
		}
	}
}
```
