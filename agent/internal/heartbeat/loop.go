package heartbeat

import (
	"context"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/logging"
)

func Start(ctx context.Context, interval time.Duration, send func(context.Context) error) {
	log := logging.From(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("heartbeat loop stopping")
			return

		case <-ticker.C:
			reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			if err := send(reqCtx); err != nil {
				log.Error("heartbeat send failed", "error", err)
			}

			cancel()
		}
	}
}
