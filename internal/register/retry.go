package register

import (
	"context"
	"math/rand"
	"time"
)

func waitWithContext(ctx context.Context, delay time.Duration) error {
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

func backoffDelay(baseDelay time.Duration, attempt int) time.Duration {
	if baseDelay <= 0 {
		return 0
	}
	factor := 1 << (attempt - 1)
	if factor > 16 {
		factor = 16
	}
	delay := time.Duration(factor) * baseDelay
	jitter := time.Duration(rand.Int63n(int64(baseDelay) + 1))
	return delay + jitter
}
