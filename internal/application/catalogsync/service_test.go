package catalogsync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/s21066/lazygotest/internal/application/catalogsync"
	"github.com/s21066/lazygotest/internal/domain/catalog"
)

type fakeSource struct {
	modules []catalog.ModuleInput
	err     error
	calls   int
}

func (f *fakeSource) Modules(ctx context.Context) ([]catalog.ModuleInput, error) {
	f.calls++
	return f.modules, f.err
}

func TestServiceSyncReturnsSnapshot(t *testing.T) {
	t.Parallel()
	modIn := catalog.ModuleInput{
		Name:    "mod",
		RootDir: "/repo",
		Packages: []catalog.PackageInput{
			{
				ImportPath: "github.com/me/a",
				Dir:        "/repo/a",
				Tests: []catalog.TestInput{
					{Name: "TestFoo", File: "foo_test.go", Line: 12, Kind: catalog.TestKindUnit},
				},
			},
		},
	}
	src := &fakeSource{modules: []catalog.ModuleInput{modIn}}
	svc := catalogsync.NewService(src)

	snap, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if src.calls != 1 {
		t.Fatalf("expected source to be called once, got %d", src.calls)
	}
	if len(snap.Modules()) != 1 {
		t.Fatalf("expected snapshot to contain single module, got %d", len(snap.Modules()))
	}
}

func TestServiceSyncPropagatesErrors(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("boom")
	src := &fakeSource{err: wantErr}
	svc := catalogsync.NewService(src)

	_, err := svc.Sync(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
