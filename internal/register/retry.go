package register

import (
	"context"
	"math/rand"
	"time"

	"github.com/monet88/chatgpt-creator/internal/util"
)

func waitWithContext(ctx context.Context, delay time.Duration) error {
	return util.WaitWithContext(ctx, delay)
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
