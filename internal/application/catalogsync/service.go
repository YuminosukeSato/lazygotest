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

// BuildSnapshotFromModules is a helper for quickly turning module inputs into Snapshot.
func BuildSnapshotFromModules(mods []catalog.ModuleInput) (catalog.Snapshot, error) {
	return catalog.BuildSnapshot(mods)
}

func (s Service) Sync(ctx context.Context) (catalog.Snapshot, error) {
	modules, err := s.source.Modules(ctx)
	if err != nil {
		return catalog.Snapshot{}, err
	}
	return catalog.BuildSnapshot(modules)
}
