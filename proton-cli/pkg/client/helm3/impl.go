package helm3

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
)

func (c *helmv3) GetRelease(name string) (*release.Release, error) {
	getter := action.NewGet(c.actionConfig)
	return getter.Run(name)
}

type ChartRef struct {
	File             string // Chart文件
	Name             string
	ChartPathOptions action.ChartPathOptions
}

func ChartRefFromFile(file string) *ChartRef {
	return &ChartRef{File: file}
}
