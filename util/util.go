package util

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

const LimitIgnoreLabels = "status/DNM,status/WIP,S: DNM,S: WIP"

// RetryOnError defines a action with retry when "fn" returns error,
// we can specify the number and interval of retries
// code snippet from https://github.com/pingcap/schrodinger
func RetryOnError(ctx context.Context, retryCount int, fn func() error) error {
	var err error
	for i := 0; i < retryCount; i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		err = fn()
		if err == nil {
			break
		}

		Error(err)
		Sleep(ctx, 2*time.Second)
	}

	return errors.Wrap(err, "retry error")
}

// Sleep defines special `sleep` with context
func Sleep(ctx context.Context, sleepTime time.Duration) {
	ticker := time.NewTicker(sleepTime)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		return
	case <-ticker.C:
		return
	}
}
