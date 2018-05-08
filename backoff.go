package backoff

import (
	"context"
	"time"
)

const (
	multiplier = 2
	interval   = 1 * time.Second
)

type Backoff struct {
	Peak time.Duration

	ctx context.Context
	n   time.Duration
}

func New(ctx context.Context) *Backoff {
	return &Backoff{
		ctx: ctx,
		n:   interval,
	}
}

func (p *Backoff) Wait() (err error) {
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

}

		}
	}
	}
}

}
