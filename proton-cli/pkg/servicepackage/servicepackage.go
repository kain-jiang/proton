package servicepackage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"helm.sh/helm/v3/pkg/chart/loader"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
)

const (
	// relativePathChart 是 charts 目录相对 service-package 的路径
	relativePathChart  = "charts"
	relativePathImages = "images"
)

// ServicePackage 是 ProtonPackage 提供的 service-package 目录的抽象。
type ServicePackage struct {
	// charts 以相对路径为 key 记录 service-package 中的 chart 的 metadata
	charts   []Chart
	images   []string
	basePath string
}

// Load 返回指定路径的 ServicePackage 对象。
func (p *ServicePackage) Load(path string) error {
	entries, err := os.ReadDir(filepath.Join(path, relativePathChart))
	if err != nil {
		return fmt.Errorf("unable to list charts directory: %w", err)
	}
	images, err := cr.GetImageTagsFromOCIPackage(filepath.Join(path, relativePathImages))
	if err != nil {
		return fmt.Errorf("unable to list images directory: %w", err)
	}
	p.basePath = path
	p.images = images
	p.charts = make(Charts, 0)
	for _, e := range entries {
		pp := filepath.Join(relativePathChart, e.Name())
		chart, err := loader.Load(filepath.Join(path, pp))
		if err != nil {
			continue
		}
		p.charts = append(p.charts, Chart{Path: pp, Metadata: *chart.Metadata})
	}
	// 排序，chart 名称升序，版本号降序
	sort.Sort(ByNameAndVersion(p.charts))
	return nil
}

// Charts 返回 service-package 中的 chart 对象列表
func (p *ServicePackage) Charts() Charts {
	return p.charts
}
func (p *ServicePackage) BaseDir() string {
	return p.basePath
}

func (p *ServicePackage) Images() []string {
	return p.images
}
