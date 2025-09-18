package usecase

import (
	"context"

	"lazygotest/internal/domain"
	"lazygotest/internal/shared/eventbus"
	"lazygotest/pkg/logger"
)

// ListPackagesUseCase handles package discovery
type ListPackagesUseCase struct {
	repo      PackageRepository
	publisher EventPublisher
}

// NewListPackagesUseCase creates a new ListPackagesUseCase
func NewListPackagesUseCase(repo PackageRepository, publisher EventPublisher) *ListPackagesUseCase {
	return &ListPackagesUseCase{
		repo:      repo,
		publisher: publisher,
	}
}

// Execute discovers and returns all packages with tests
func (uc *ListPackagesUseCase) Execute(ctx context.Context) ([]*domain.Package, error) {
	logger.Debug("Executing ListPackagesUseCase")

	packages, err := uc.repo.ListPackages(ctx)
	if err != nil {
		logger.Error("Failed to list packages", "error", err)
		return nil, err
	}

	// Publish event for each discovered package
	for _, pkg := range packages {
		uc.publisher.PublishAsync(ctx, eventbus.TopicPackageFound, pkg)
	}

	logger.Info("Found packages", "count", len(packages))
	return packages, nil
}

// FilterPackages filters packages based on a search pattern
func (uc *ListPackagesUseCase) FilterPackages(packages []*domain.Package, pattern string) []*domain.Package {
	return uc.repo.FilterPackages(packages, pattern)
}
