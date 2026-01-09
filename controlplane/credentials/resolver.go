package credentials

import "context"

type Resolver interface {
	Resolve(ctx context.Context, ref string) ([]byte, error)
}
