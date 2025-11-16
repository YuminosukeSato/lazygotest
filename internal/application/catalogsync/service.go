package catalogsync

import (
	"context"

	"github.com/s21066/lazygotest/internal/domain/catalog"
)

type Source interface {
	Modules(ctx context.Context) ([]catalog.ModuleInput, error)
}

type Service struct {
	source Source
}

func NewService(source Source) Service {
	return Service{source: source}
}

func (s Service) Sync(ctx context.Context) (catalog.Snapshot, error) {
	modules, err := s.source.Modules(ctx)
	if err != nil {
		return catalog.Snapshot{}, err
	}
	return catalog.BuildSnapshot(modules)
}
