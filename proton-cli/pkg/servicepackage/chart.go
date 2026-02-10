package servicepackage

import (
	"strings"

	"golang.org/x/mod/semver"
	"helm.sh/helm/v3/pkg/chart"
)

// Chart 代表 service-package 包含的 chart。
type Chart struct {
	// chart 文件、目录相对于 service-package 的路径。
	Path string
	// Chart 的元数据。
	Metadata chart.Metadata
}

type Charts []Chart

// Get 返回指定名称和版本的 Chart。
// 如果未指定版本则返回最新的版本。
// 如果不存在则返回 nil。
// 因为 Charts 是以 Name 升序、以 Version 降序排列，所以每个 Name 的第一个即为最新版本。
func (l Charts) Get(name, version string) *Chart {
	for i := range l {
		if l[i].Metadata.Name == name && (l[i].Metadata.Version == version || version == "") {
			return &l[i]
		}
	}
	return nil
}

// ByNameAndVersion 以 Name 增序，再以 Version 降序
// 从遍历时方便获取某个 Name 的最新版本
type ByNameAndVersion Charts

func (a ByNameAndVersion) Len() int      { return len(a) }
func (a ByNameAndVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByNameAndVersion) Less(i, j int) bool {
	if a[i].Metadata.Name == a[j].Metadata.Name {
		return semver.Compare("v"+a[i].Metadata.Version, "v"+a[j].Metadata.Version) == 1
	}
	return strings.Compare(a[i].Metadata.Name, a[j].Metadata.Name) == -1
}
