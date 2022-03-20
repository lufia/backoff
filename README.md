# backoff

Backoff implements exponential backoff algorithm in Go.

[![GoDev][godev-image]][godev-url]
[![Actions Status][actions-image]][actions-url]
[![Coverage Status][coveralls-image]][coveralls-url]

## DESCRIPTION

## EXAMPLE

This is simple example.
`w.Wait` blocks until next period that are:

1. 1±0.5s
2. 2±1s
3. 4±2s
4. 8±4s

It increases exponentially forever.

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


[godev-image]: https://pkg.go.dev/badge/github.com/lufia/backoff
[godev-url]: https://pkg.go.dev/github.com/lufia/backoff
[actions-image]: https://github.com/lufia/backoff/workflows/Test/badge.svg?branch=main
[actions-url]: https://github.com/lufia/backoff/actions?workflow=Test
[coveralls-image]: https://coveralls.io/repos/github/lufia/backoff/badge.svg
[coveralls-url]: https://coveralls.io/github/lufia/backoff
