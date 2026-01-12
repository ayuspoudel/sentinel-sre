package heartbeat

import (
	"context"
	"time"
)

func Start(ctx context.Context, interval time.Duration, send func(context.Context) error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = send(ctx)
		}
	}
}
