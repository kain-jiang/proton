package helm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"taskrunner/pkg/component"
	"taskrunner/trait"
)

type Local struct {
	Path string `json:"path"`
}

func (l *Local) Name() string {
	return "local"
}

func (l *Local) Store(ctx context.Context, chart *component.HelmComponent, data []byte) *trait.Error {
	panic("unsupport store")
}

func (l *Local) Fetch(ctx context.Context, c *component.HelmComponent) ([]byte, *trait.Error) {
	cname := fmt.Sprintf("%s-%s.tgz", c.Name, c.Version)
	chartPath := filepath.Join(l.Path, cname)
	data, err := os.ReadFile(chartPath)
	if err != nil {
		return nil, &trait.Error{
			Err:      err,
			Internal: trait.ErrHelmRepoUnknow,
			Detail:   fmt.Errorf("read chart %s error", chartPath),
		}
	}
	return data, nil
}
