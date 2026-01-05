package executor

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/internal/action"
)

/*
	@ayuspoudel
	Executor is a side effect thing. Executor will consume actions produced by engine
	and apply the effects to real time infra. It is idempotent and it does not influence
	decisions. It also never evaluates metrics or policy.
*/

type Executor interface {
	Name() string
	Apply(ctx context.Context, a action.Action) error
}
