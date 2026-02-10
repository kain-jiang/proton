package helm3

import (
	"errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type getParams struct {
	ignoreNotExist bool
	version        int
}

type GetOption func(getParams *getParams)

func WithGetIgnoreNotExist(ignoreNotExist bool) GetOption {
	return func(getParams *getParams) {
		getParams.ignoreNotExist = ignoreNotExist
	}
}

func WithGetVersion(version int) GetOption {
	return func(getParams *getParams) {
		getParams.version = version
	}
}

func (c *helmv3) GetRelease(name string, opts ...GetOption) (*release.Release, error) {
	param := &getParams{}
	for _, opt := range opts {
		opt(param)
	}
	getter := action.NewGet(c.actionConfig)
	getter.Version = param.version

	rls, err := getter.Run(name)
	if param.ignoreNotExist && errors.Is(err, driver.ErrReleaseNotFound) {
		return nil, nil
	}
	return rls, err
}

type ChartRef struct {
	File             string // Chart文件
	Name             string
	ChartPathOptions action.ChartPathOptions
}

func ChartRefFromFile(file string) *ChartRef {
	return &ChartRef{File: file}
}
