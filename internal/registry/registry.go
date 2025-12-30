package registry

import "context"

type Registry interface {
	Load(ctx context.Context) error
	Guards() []Guard
}

type Guard struct {
	Name string
}
