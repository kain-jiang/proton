package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

// Manager 负责按 InstallPlan 驱动 helm install/upgrade。
type Manager struct {
	Helm      helm3.Client
	Namespace string
	Logger    logrus.FieldLogger
}

// NewManager 创建一个 Manager，helm client 使用指定 namespace。
func NewManager(namespace string, logger logrus.FieldLogger) (*Manager, error) {
	if logger == nil {
		l := logrus.New()
		l.SetLevel(logrus.InfoLevel)
		logger = l
	}
	entry, ok := logger.(*logrus.Entry)
	if !ok {
		entry = logrus.NewEntry(logrus.StandardLogger()).WithField("component", "helm")
	}
	helmCli, err := helm3.NewCli(namespace, entry)
	if err != nil {
		return nil, fmt.Errorf("init helm client: %w", err)
	}
	return &Manager{
		Helm:      helmCli,
		Namespace: namespace,
		Logger:    logger,
	}, nil
}

// Install 按 manifest 文件路径构建安装计划并执行。
func (m *Manager) Install(ctx context.Context, manifestPath string, cfg *configuration.ClusterConfig, opts InstallOptions) error {
	ns := opts.Namespace
	if ns == "" {
		ns = m.Namespace
	}

	values := BuildHelmValues(cfg, ns)
	plan, err := BuildInstallPlanWithValues(manifestPath, values)
	if err != nil {
		return fmt.Errorf("build install plan: %w", err)
	}

	timeout := 5 * time.Minute
	if opts.Timeout != "" {
		d, err := time.ParseDuration(opts.Timeout)
		if err == nil {
			timeout = d
		}
	}

	totalSteps := len(plan.Steps)

	if opts.DryRun {
		m.Logger.Infof("=== Install Plan (%d steps, dry-run) ===", totalSteps)
		for stepIdx, step := range plan.Steps {
			m.Logger.Infof("Step %d/%d (%d releases):", stepIdx+1, totalSteps, len(step))
			for _, item := range step {
				m.Logger.Infof("  %-40s chart=%-40s version=%s",
					item.ReleaseName, item.ChartName, item.ChartVersion)
			}
		}
		return nil
	}

	for stepIdx, step := range plan.Steps {
		m.Logger.Infof("Step %d/%d: installing %d release(s)", stepIdx+1, totalSteps, len(step))
		for _, item := range step {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if err := m.installRelease(ctx, item, values, ns, timeout, opts.CreateNamespace); err != nil {
				return fmt.Errorf("install release %s: %w", item.ReleaseName, err)
			}
		}
	}
	return nil
}

// InstallWithValues 按 manifest 文件路径构建安装计划，使用已构建好的 values map 执行安装。
func (m *Manager) InstallWithValues(ctx context.Context, manifestPath string, values map[string]interface{}, opts InstallOptions) error {
	plan, err := BuildInstallPlanWithValues(manifestPath, values)
	if err != nil {
		return fmt.Errorf("build install plan: %w", err)
	}

	ns := opts.Namespace
	if ns == "" {
		ns = m.Namespace
	}

	timeout := 5 * time.Minute
	if opts.Timeout != "" {
		d, err := time.ParseDuration(opts.Timeout)
		if err == nil {
			timeout = d
		}
	}

	m.Logger.Infof("Install plan: %d step(s)", len(plan.Steps))
	for stepIdx, step := range plan.Steps {
		for _, item := range step {
			m.Logger.Infof("  Step %d: %s (chart=%s version=%s product=%s)",
				stepIdx+1, item.ReleaseName, item.ChartName, item.ChartVersion, item.Product)
		}
	}

	totalSteps := len(plan.Steps)
	for stepIdx, step := range plan.Steps {
		m.Logger.Infof("Step %d/%d: installing %d release(s)", stepIdx+1, totalSteps, len(step))
		for _, item := range step {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// 如果 InstallOptions 中指定了 Helm 仓库，覆盖 manifest 中的配置
			// 用于离线部署时自动使用内置 ChartMuseum
			if opts.HelmRepoName != "" {
				item.HelmRepoName = opts.HelmRepoName
			}
			if opts.HelmRepoURL != "" {
				item.HelmRepoURL = opts.HelmRepoURL
			}

			if opts.DryRun {
				m.Logger.Infof("[dry-run] would install %s chart=%s version=%s repo=%s",
					item.ReleaseName, item.ChartName, item.ChartVersion, item.HelmRepoName)
				continue
			}

			if err := m.installRelease(ctx, item, values, ns, timeout, opts.CreateNamespace); err != nil {
				return fmt.Errorf("install release %s: %w", item.ReleaseName, err)
			}
		}
	}
	return nil
}

// installRelease 安装或升级单个 helm release，等价于：
//
//	helm upgrade --install <release> <repo>/<chart> \
//	  -f <values.yaml> --version <ver> --devel --wait --timeout=600s --namespace <ns>
func (m *Manager) installRelease(
	_ context.Context,
	item InstallItem,
	values map[string]interface{},
	namespace string,
	timeout time.Duration,
	createNamespace bool,
) error {
	log := m.Logger.WithField("release", item.ReleaseName).WithField("chart", item.ChartName)
	shouldWait := item.WaitForReady
	log = log.WithField("stage", item.Stage).WithField("wait", shouldWait).WithField("timeout", timeout)
	log.Infof("Submitting release install (version=%s namespace=%s)", item.ChartVersion, namespace)
	if shouldWait {
		log.Info("waiting for release readiness before continuing")
	} else {
		log.Info("Release submitted without readiness wait")
	}

	// 合并 base values 和 per-release values
	// per-release values 优先级更高，会覆盖 base values 中的同名字段
	mergedValues := deepMergeValues(values, item.Values)

	// ISF 不能使用统一的 depServices.rds.database，各 chart 自己指定具体库名。
	// 等同于 deploy/scripts/services/isf.sh 中的 `sed '/database:/d'`。
	if strings.EqualFold(item.Product, "isf") {
		mergedValues = StripISFRdsDatabase(mergedValues)
	}

	chartRef := &helm3.ChartRef{
		Name: repoChartName(item.HelmRepoName, item.ChartName),
		ChartPathOptions: action.ChartPathOptions{
			Version: item.ChartVersion,
		},
	}

	helmCli := m.Helm.NameSpace(namespace)
	err := helmCli.Upgrade(item.ReleaseName, chartRef,
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeValues(mergedValues),
		helm3.WithUpgradeCreateNamespace(createNamespace),
		helm3.WithUpgradeDevel(true),
		helm3.WithUpgradeWait(shouldWait, timeout),
	)
	if err != nil {
		log.WithError(err).Errorf("install release failed")
		return err
	}
	if shouldWait {
		log.Info("Release readiness confirmed")
	} else {
		log.Info("Release install request submitted successfully")
	}
	return nil
}

// deepMergeValues 深度合并两个 values map，override 中的值优先级更高。
// 用于合并 base values（从 ClusterConfig 生成）和 per-release values（从 manifest 定义）。
func deepMergeValues(base, override map[string]interface{}) map[string]interface{} {
	if base == nil {
		base = make(map[string]interface{})
	}
	if override == nil {
		return base
	}

	result := make(map[string]interface{})
	// 先复制 base 的所有字段
	for k, v := range base {
		result[k] = v
	}

	// 递归合并 override 的字段
	for k, overrideVal := range override {
		baseVal, exists := result[k]
		if !exists {
			// base 中不存在，直接使用 override 的值
			result[k] = overrideVal
			continue
		}

		// 如果两者都是 map，递归合并
		baseMap, baseIsMap := baseVal.(map[string]interface{})
		overrideMap, overrideIsMap := overrideVal.(map[string]interface{})
		if baseIsMap && overrideIsMap {
			result[k] = deepMergeValues(baseMap, overrideMap)
		} else {
			// 否则 override 覆盖 base
			result[k] = overrideVal
		}
	}

	return result
}

// DeepMergeValues exposes the existing install-time merge behavior for callers
// that need to combine generated values with CLI overrides before install.
func DeepMergeValues(base, override map[string]interface{}) map[string]interface{} {
	return deepMergeValues(base, override)
}

// Uninstall 按 manifest 文件路径构建卸载计划并执行卸载。
// 卸载顺序与安装顺序相反：先卸载当前产品，再卸载依赖。
func (m *Manager) Uninstall(ctx context.Context, manifestPath string, opts UninstallOptions) error {
	plan, err := BuildInstallPlan(manifestPath)
	if err != nil {
		return fmt.Errorf("build uninstall plan: %w", err)
	}

	ns := opts.Namespace
	if ns == "" {
		ns = m.Namespace
	}

	// 解析超时时间，默认 1 分钟。
	// 卸载卡住多半是 hook/finalizer 等问题，较短默认值能更快暴露异常。
	timeout := 1 * time.Minute
	if opts.Timeout != "" {
		d, err := time.ParseDuration(opts.Timeout)
		if err == nil {
			timeout = d
		}
	}

	// 按安装计划逆序卸载：后安装的先卸载。
	// 同一步内的顺序也逆序，尽量保持与安装顺序严格对偶。
	releases := make([]string, 0)
	for stepIdx := len(plan.Steps) - 1; stepIdx >= 0; stepIdx-- {
		step := plan.Steps[stepIdx]
		for itemIdx := len(step) - 1; itemIdx >= 0; itemIdx-- {
			releases = append(releases, step[itemIdx].ReleaseName)
		}
	}

	m.Logger.Infof("Uninstall plan: %d release(s) to uninstall", len(releases))
	for _, releaseName := range releases {
		m.Logger.Infof("  - %s", releaseName)
	}

	if opts.DryRun {
		m.Logger.Info("Dry-run mode: no releases will be uninstalled")
		return nil
	}

	// 卸载所有 releases
	helmCli := m.Helm.NameSpace(ns)
	for _, releaseName := range releases {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		log := m.Logger.WithField("release", releaseName)
		log.Infof("Uninstalling %s from namespace %s", releaseName, ns)

		// 使用 IgnoreNotFound 选项，避免 Helm 客户端打印 "not found" 错误日志
		err := helmCli.Uninstall(releaseName,
			helm3.WithUninstallIgnoreNotFound(true),
			helm3.WithUninstallTimeout(timeout))
		if err != nil {
			// 其他错误只打印 info，避免误导用户
			log.Infof("release %s uninstall skipped: %v", releaseName, err)
			continue
		}
		log.Infof("✓ %s uninstalled successfully", releaseName)
	}

	return nil
}

// repoChartName 拼接 repo/chart 名称，如 "kweaver/auth-service"。
func repoChartName(repoName, chartName string) string {
	if repoName == "" {
		return chartName
	}
	return repoName + "/" + chartName
}
