package retry

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"math/rand"
	"time"

	errs "github.com/deimossy/order-processing-system/pkg/errors"
)

var (
	CodeAdminShutdownViolation      = "57P01"
	CodeCrashShutdownViolation      = "57P02"
	CodeCannotConnectViolation      = "57P03"
	CodeSerializationViolation      = "40001"
	CodeDeadlockViolation           = "40P01"
	CodeTooManyConnectionsViolation = "53300"
	CodeQueryCanceledViolation      = "57014"
)

func Do(
	ctx context.Context,
	maxRetries int,
	backoff, maxBackoff time.Duration,
	fn func() error,
	shouldRetry func(error) bool,
) error {
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !shouldRetry(err) {
			return err
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

func ShouldRetryTx(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case CodeSerializationViolation,
			CodeDeadlockViolation:
			return true
		}
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled) {
		return true
	}

	return false
}

func ShouldRetryQuery(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case CodeAdminShutdownViolation,
			CodeCrashShutdownViolation,
			CodeCannotConnectViolation,
			CodeTooManyConnectionsViolation,
			CodeQueryCanceledViolation:
			return true
		}
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, context.Canceled) {
		return true
	}

	return false
}
