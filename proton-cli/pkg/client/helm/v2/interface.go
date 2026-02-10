package v2

import "context"

type Interface interface {
	UpdateRepoCache(ctx context.Context, repo string) error

	// Install or update release to the latest version in the chart repository
	// using the specified chart and values.
	Reconcile(ctx context.Context, release string, chart string, values map[string]any) error
}
