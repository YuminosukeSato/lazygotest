package usecase

import (
	"context"

	"github.com/YuminosukeSato/lazygotest/internal/adapter/secondary/runner"
	"github.com/YuminosukeSato/lazygotest/internal/domain"
)

// PackageRepository defines package discovery operations
type PackageRepository interface {
	ListPackages(ctx context.Context) ([]*domain.Package, error)
	GetPackage(ctx context.Context, pkgPath string) (*domain.Package, error)
	FilterPackages(packages []*domain.Package, pattern string) []*domain.Package
}

// TestRunner defines test execution operations
type TestRunner interface {
	Run(ctx context.Context, opts runner.RunOptions) (<-chan domain.TestEvent, <-chan error)
	ListTests(ctx context.Context, pkg string) ([]string, error)
}

// EventPublisher defines event publishing operations
type EventPublisher interface {
	Publish(ctx context.Context, topic string, event interface{})
	PublishAsync(ctx context.Context, topic string, event interface{})
}
