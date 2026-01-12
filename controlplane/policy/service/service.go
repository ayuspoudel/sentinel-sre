package service

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/spec"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/status"
)

/*
Author: @ayuspoudel
It defines what our policy registry can do. These services are all internal mechanisms that we can perform
It will be later exposed to http endpoints so other services can call these. But, in sum, these are simply
what we are doing with policy registry.
1. We are applying policy using ApplyPolicy. It will also validate, check reality, store it and compute status.
2. We are getting policy using GetPolicy.
3. We are listing all policies using ListPolicies.
4. We are deleting policy using DeletePolicy.
*/
type PolicyService interface {
	ApplyPolicy(ctx context.Context, p *spec.PolicySpec) (*status.PolicyStatus, error)
	GetPolicy(ctx context.Context, name string) (*spec.PolicySpec, error)
	ListPolicies(ctx context.Context) ([]*spec.PolicySpec, error)
	DeletePolicy(ctx context.Context, name string) error
	GetStatus(ctx context.Context, name string) (*status.PolicyStatus, error)
}
