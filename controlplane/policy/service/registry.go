package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/spec"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/status"
	"github.com/ayuspoudel/sentinel-sre/controlplane/policy/store"
)

type RegistryService struct {
	store store.PolicyStore

	clusters   ClusterRegistry
	controller ControllerReader
	prom       PrometheusClient
}

func NewRegistryService(store store.PolicyStore, clusters ClusterRegistry, controller ControllerReader, prom PrometheusClient) *RegistryService {
	return &RegistryService{
		store:      store,
		clusters:   clusters,
		controller: controller,
		prom:       prom,
	}
}

func (s *RegistryService) ApplyPolicy(
	ctx context.Context,
	p *spec.PolicySpec,
) (*status.PolicyStatus, error) {

	// Pure spec validation (hard fail)
	if err := spec.Validate(p); err != nil {
		return nil, err
	}

	// Cluster intent validation (hard fail)
	exists, err := s.clusters.ClusterExists(ctx, p.Target.Cluster)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf(
			"cluster %q is not registered in sentinel; add cluster before applying policy",
			p.Target.Cluster,
		)
	}

	// Status is only created after intent is valid
	st := &status.PolicyStatus{
		PolicyName:      p.Metadata.Name,
		ClusterExists:   true,
		LastValidatedAt: time.Now().UTC(),
	}

	// Controller-backed validation (soft)
	st.ClusterReachable, _ = s.controller.ClusterReachable(ctx, p.Target.Cluster)
	st.NamespaceExists, _ = s.controller.NamespaceExists(ctx, p.Target.Cluster, p.Target.Namespace)
	st.AgentInstalled, _ = s.controller.AgentInstalled(ctx, p.Target.Cluster)
	st.AgentHealthy, _ = s.controller.AgentHealthy(ctx, p.Target.Cluster)

	if !st.NamespaceExists {
		errMsg := "namespace does not exist"
		st.LastError = &errMsg
	}

	// Prometheus validation (soft)
	if err := s.prom.Query(ctx, p.Signals.Traffic.Query); err == nil {
		if err := s.prom.Query(ctx, p.Signals.Errors.Query); err == nil {
			st.PrometheusReachable = true
			st.QueriesValid = true
		}
	}

	// Persist spec + status (authoritative write)
	if err := s.store.UpsertPolicy(ctx, p); err != nil {
		return nil, err
	}
	if err := s.store.UpdateStatus(ctx, st); err != nil {
		return nil, err
	}

	return st, nil
}

func (s *RegistryService) GetPolicy(ctx context.Context, name string) (*spec.PolicySpec, error) {
	return s.store.GetPolicy(ctx, name)
}

func (s *RegistryService) ListPolicies(ctx context.Context) ([]*spec.PolicySpec, error) {
	return s.store.ListPolicies(ctx)
}

func (s *RegistryService) DeletePolicy(ctx context.Context, name string) error {
	return s.store.DeletePolicy(ctx, name)
}

func (s *RegistryService) GetStatus(ctx context.Context, name string) (*status.PolicyStatus, error) {
	return s.store.GetStatus(ctx, name)
}
