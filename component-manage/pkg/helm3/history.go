package helm3

import (
	"errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type historyParams struct {
	max            int
	ignoreNotExist bool
}

type HistoryOption func(historyParams *historyParams)

func WithHistoryMax(max int) HistoryOption {
	return func(historyParams *historyParams) {
		historyParams.max = max
	}
}

func WithHistoryIgnoreNotExist(ignoreNotExist bool) HistoryOption {
	return func(historyParams *historyParams) {
		historyParams.ignoreNotExist = ignoreNotExist
	}
}

func (c *helmv3) HistoryRelease(name string, opts ...HistoryOption) ([]*release.Release, error) {
	param := &historyParams{}
	for _, opt := range opts {
		opt(param)
	}
	historier := action.NewHistory(c.actionConfig)
	historier.Max = param.max

	rlses, err := historier.Run(name)
	if param.ignoreNotExist && errors.Is(err, driver.ErrReleaseNotFound) {
		return nil, nil
	}
	return rlses, err
}
