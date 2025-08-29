package retry

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

var ErrMaxRetriesAttemptsExceeded = errors.New("max retries attempts exceeded")

func Do(ctx context.Context, maxRetries int, backoff time.Duration, fn func() error) error {
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = fn()
		if err == nil {
			return nil
		}

		if attempt == maxRetries-1 {
			return err
		}

		backoff = backoff * time.Duration(math.Pow(2, float64(attempt)))
		jitter := time.Duration(rand.Int63n(int64(backoff)))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff + jitter):
		}
	}

	return ErrMaxRetriesAttemptsExceeded
}
