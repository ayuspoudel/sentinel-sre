package service

import (
	"context"

	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/model"
	"github.com/ayuspoudel/sentinel-sre/controlplane/cluster/store"
)

/*
	Author: @ayuspoudel
	This service represents cluster intent management.
	It does not reconcile anything and does not talk to agents.
	It only validates and persists cluster registration intent.
*/

type Service struct {
	store store.Store
}

func NewService(store store.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Register(ctx context.Context, c *model.Cluster) error {
	return s.store.Create(ctx, c)
}

func (s *Service) Get(ctx context.Context, name string) (*model.Cluster, error) {
	return s.store.Get(ctx, name)
}

func (s *Service) List(ctx context.Context) ([]*model.Cluster, error) {
	return s.store.List(ctx)
}

func (s *Service) Delete(ctx context.Context, name string) error {
	return s.store.Delete(ctx, name)
}
