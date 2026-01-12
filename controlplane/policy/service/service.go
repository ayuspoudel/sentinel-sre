package service

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/spec"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/status"
)

type PolicyService interface {
	ApplyPolicy(ctx context.Context, p *spec.PolicySpec) (*status.PolicyStatus, error)
	GetPolicy(ctx context.Context, name string) (*spec.PolicySpec, error)
	ListPolicies(ctx context.Context) ([]*spec.PolicySpec, error)
	DeletePolicy(ctx context.Context, name string) error
	GetStatus(ctx context.Context, name string) (*status.PolicyStatus, error)
}
