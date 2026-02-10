package test

import (
	// load test file
	"context"
	"embed"
	"fmt"
	"io/fs"

	"taskrunner/pkg/component"
	"taskrunner/trait"
)

// MemoryHelmRepoMock use as a memory repo mock
type MemoryHelmRepoMock struct {
	RepoName string
	embed.FS
}

// Fetch repo mock
func (r *MemoryHelmRepoMock) Fetch(ctx context.Context, c *component.HelmComponent) ([]byte, *trait.Error) {
	path := fmt.Sprintf("testdata/charts/%s-%s.tgz", c.Name, c.Version)
	bs, err := r.FS.ReadFile(path)
	if _, ok := err.(*fs.PathError); ok {
		return nil, &trait.Error{Internal: trait.ErrHelmChartNoFound, Err: err, Detail: fmt.Sprintf("chart %s:%s not found", c.Name, c.Version)}
	}
	return bs, nil
}

// Name helm repo mock
func (r *MemoryHelmRepoMock) Name() string {
	return r.RepoName
}

// Store helm repo mock
func (r *MemoryHelmRepoMock) Store(ctx context.Context, chart *component.HelmComponent, data []byte) *trait.Error {
	return nil
}
