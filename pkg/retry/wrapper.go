package retry

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	errs "github.com/deimossy/order-processing-system/pkg/errors"
)

func Do(ctx context.Context, maxRetries int, backoff, maxBackoff time.Duration, fn func() error) error {
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if attempt == maxRetries-1 {
			break
		}

		delay := min(backoff*(1<<attempt), maxBackoff) // exp backoff

		jitter := time.Duration(rand.Int63n(int64(delay)))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(jitter): // full jitter
		}
	}

	return fmt.Errorf("%w: last error: %v", errs.ErrMaxRetriesAttemptsExceeded, err)
}
