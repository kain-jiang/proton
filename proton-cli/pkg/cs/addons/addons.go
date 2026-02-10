package addons

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"

	"github.com/go-test/deep"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/util/version"
	"sigs.k8s.io/yaml"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

const (
	// node-exporter 的 chart 的名字
	ChartNameNodeExporter = "proton-node-exporter"
	// kube-state-metrics 的 chart 的名字
	ChartNameStateMetrics = "proton-kube-state-metrics"

	// node-exporter 的 helm release 的名字
	ReleaseNameNodeExporter = "node-exporter"
	// kube-state-metrics 的 helm release 的名字
	ReleaseNameStateMetrics = "proton-kube-state-metrics"
)

var chartNameMap = map[configuration.CSAddonName]string{
	configuration.CSAddonNameNodeExporter: ChartNameNodeExporter,
	configuration.CSAddonNameStateMetrics: ChartNameStateMetrics,
}

var releaseNameMap = map[configuration.CSAddonName]string{
	configuration.CSAddonNameNodeExporter: ReleaseNameNodeExporter,
	configuration.CSAddonNameStateMetrics: ReleaseNameStateMetrics,
}

// Reconcile 处理 Proton CS 的插件，至期望的最终状态
//
//   - release 不存在，使用 chart 仓库中最新版本安装
//   - release 版本高于 chart 仓库中的最新版本，无操作
//   - release 版本等于 chart 仓库中的最新版本，如果 values 不符合期望则更新
//   - release 版本低于 chart 仓库中的最新版本，更新至 chart 仓库的最新版本
func Reconcile(ctx context.Context, lg logrus.FieldLogger, h helm3.Client, pkg *servicepackage.ServicePackage, registry string, name configuration.CSAddonName) error {
	var ok bool

	// addon 的 chart 名称
	var chart string
	if chart, ok = chartNameMap[name]; !ok {
		return fmt.Errorf("chart name of addon %s is not registered", name)
	}

	// addon 的 release 名称
	var release string
	if release, ok = releaseNameMap[name]; !ok {
		return fmt.Errorf("release name of addon %s is not registered", name)
	}

	chartInfo := pkg.Charts().Get(chart, "")
	latestVersion, err := version.ParseSemantic(chartInfo.Metadata.Version)
	if err != nil {
		return err
	}

	var releaseVersionString string
	// releaseVersionString, err := h.GetReleaseVersion(release)
	r, err := h.GetRelease(release)
	if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) && !errors.Is(err, driver.ErrNoDeployedReleases) {
		return err
	}
	if r != nil && r.Chart != nil && r.Chart.Metadata != nil {
		releaseVersionString = r.Chart.Metadata.Version
	} else {
		releaseVersionString = "0.0.0"
	}
	releaseVersion, err := version.ParseSemantic(releaseVersionString)
	if err != nil {
		return err
	}

	// release 的 chart 版本高于 chart 仓库的最新版本，不需要更新
	if latestVersion.LessThan(releaseVersion) {
		lg.Debugf("skip update helm release %v, because the latest chart version %v is less than the version of helm release %v, ", release, latestVersion, releaseVersion)
		return nil
	}

	// 期望的 release values
	var values = Values{
		Image: ValuesImage{
			Registry: registry,
		},
	}

	// 如果 release 的版本小于 chart 仓库的最新版本或 values 与期望不同则更新
	if !(releaseVersion.LessThan(latestVersion) || deep.Equal(r.Config, toMap(values)) != nil) {
		lg.Infof("skip update helm release %v, because helm values are satisfied", release)
		return nil
	}

	lg.Infof("install or upgrade helm release of addon %v: %v", name, release)
	lg.Debugf("update helm release %v, values:\n%v", release, toYAML(values))
	if err := h.Upgrade(
		release, &helm3.ChartRef{File: path.Join(pkg.BaseDir(), chartInfo.Path)},
		helm3.WithUpgradeValues(toMap(values)),
		helm3.WithUpgradeInstall(true),
	); err != nil {
		lg.Errorf("install or upgrade helm release fail: %v, release: %v, chart: %v, values:\n%v", err, release, chart, toYAML(values))
		return err
	}

	return nil
}

type Values struct {
	Image ValuesImage `json:"image,omitempty"`
}

type ValuesImage struct {
	Registry string `json:"registry,omitempty"`
}

// toYAML 以 yaml 格式序列化，便于在日志中显示 struct
func toYAML(v any) string {
	b, _ := yaml.Marshal(v)
	return string(b)
}

// toMap 将任意结构转为 map[string]any{}，便于比较 release 的 values
func toMap(v any) map[string]any {
	b, _ := json.Marshal(v)
	var r map[string]any
	_ = json.Unmarshal(b, &r)
	return r
}
