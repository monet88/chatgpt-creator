package util

import (
	"context"
	"time"
)

// WaitWithContext sleeps for the given delay but returns early if the context is cancelled.
func WaitWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
