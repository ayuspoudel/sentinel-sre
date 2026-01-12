package store

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/spec"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/status"
)

/*
Author: @ayuspoudel
This defines what policy service's storage layer will do. It defines all the methods
supported independed of db type.
*/
type PolicyStore interface {
	UpsertPolicy(ctx context.Context, policy *spec.PolicySpec) error
	GetPolicy(ctx context.Context, name string) (*spec.PolicySpec, error)
	ListPolicies(ctx context.Context) ([]*spec.PolicySpec, error)
	DeletePolicy(ctx context.Context, name string) error

	GetStatus(ctx context.Context, policyName string) (*status.PolicyStatus, error)
	UpdateStatus(ctx context.Context, st *status.PolicyStatus) error
}
