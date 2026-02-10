package helm

import (
	"context"
	"errors"
	"fmt"

	"taskrunner/pkg/component"
	"taskrunner/trait"
)

// Repo store and load chart to helm repo
type Repo interface {
	Store(ctx context.Context, chart *component.HelmComponent, data []byte) *trait.Error
	Fetch(ctx context.Context, chart *component.HelmComponent) ([]byte, *trait.Error)
	Name() string
}

// Repos repo manager
type Repos struct {
	repos map[string]Repo
}

// NewHelmIndexRepo create a helm repos manager from repo
func NewHelmIndexRepo(repoConfig ...Repo) *Repos {
	rs := Repos{
		repos: make(map[string]Repo),
	}

	for _, repo := range repoConfig {
		rs.repos[repo.Name()] = repo
	}

	return &rs
}

// Store chart package
func (h *Repos) Store(ctx context.Context, chart *component.HelmComponent, data []byte) *trait.Error {
	r, ok := h.repos[chart.Repository]
	if !ok {
		return &trait.Error{
			Internal: trait.ErrHelmRepoNoFound,
			Err:      errors.New("repo not founc"),
			Detail:   fmt.Sprintf("helm repo %s not found when store chart", chart.Repository),
		}
	}
	return r.Store(ctx, chart, data)
}

// Fetch down load the chart
func (h *Repos) Fetch(ctx context.Context, c *component.HelmComponent) ([]byte, *trait.Error) {
	r, ok := h.repos[c.Repository]
	if !ok {
		return nil, &trait.Error{
			Internal: trait.ErrHelmRepoNoFound,
			Err:      errors.New("repo not founc"),
			Detail:   fmt.Sprintf("helm repo %s not found when fetch chart", c.Repository),
		}
	}
	return r.Fetch(ctx, c)
}

// Name imply helm repo
func (h *Repos) Name() string {
	return ""
}
