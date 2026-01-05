package executor

import (
	"context"
	"log"

	"github.com/ayuspoudel/sentinel-sre/internal/action"
)

/*
	@ayuspoudel
	Logging executor is used for testing, dry runs, validating phase 3 logic end to end
*/

type LoggingExecutor struct{}

func newLoggingExecutor() *LoggingExecutor {
	return &LoggingExecutor{}

}

func (l *LoggingExecutor) Name() string {
	return "logging"
}
func (l *LoggingExecutor) Apply(ctx context.Context, a action.Action) error {
	log.Printf(
		"[executor=%s] guard=%s action=%s phase=%s reason=%s",
		l.Name(),
		a.GuardName,
		a.Type,
		a.Phase,
		a.Reason,
	)
	return nil
}
