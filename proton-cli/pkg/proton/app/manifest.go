package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"sigs.k8s.io/yaml"
)

// LoadManifest 从文件路径加载 VersionSet manifest。
func LoadManifest(path string) (*VersionSet, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve manifest path %q: %w", path, err)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest %q: %w", absPath, err)
	}
	var vs VersionSet
	if err := yaml.Unmarshal(data, &vs); err != nil {
		return nil, fmt.Errorf("parse manifest %q: %w", absPath, err)
	}
	return &vs, nil
}

// BuildInstallPlan 根据 manifest 文件路径递归展开 dependencies，
// 并按 stage 构建有序安装计划。
//
// 安装顺序：
//  1. 先递归安装所有 dependencies（按声明顺序）
//  2. 同一 VersionSet 内，stage="pre" 的 release 作为 Step[0]，其余作为 Step[1]
//
// 依赖去重：同一 product+version 只安装一次（按首次出现顺序）。
func BuildInstallPlan(manifestPath string) (*InstallPlan, error) {
	seen := make(map[string]bool)
	var allSteps [][]InstallItem

	if err := buildPlanRecursive(manifestPath, seen, &allSteps); err != nil {
		return nil, err
	}

	return &InstallPlan{Steps: allSteps}, nil
}

// buildPlanRecursive 递归处理一个 manifest，将生成的步骤追加到 allSteps。
func buildPlanRecursive(manifestPath string, seen map[string]bool, allSteps *[][]InstallItem) error {
	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		return fmt.Errorf("resolve path %q: %w", manifestPath, err)
	}

	vs, err := LoadManifest(absPath)
	if err != nil {
		return err
	}

	dedupKey := vs.Product + "@" + vs.Version
	if seen[dedupKey] {
		return nil
	}
	seen[dedupKey] = true

	manifestDir := filepath.Dir(absPath)

	// 先递归处理 dependencies
	for _, dep := range vs.Dependencies {
		depManifestPath := dep.Manifest
		if depManifestPath == "" {
			continue
		}
		if !filepath.IsAbs(depManifestPath) {
			depManifestPath = filepath.Join(manifestDir, depManifestPath)
		}
		if err := buildPlanRecursive(depManifestPath, seen, allSteps); err != nil {
			if dep.Optional {
				fmt.Fprintf(os.Stderr, "warning: optional dependency %s skipped: %v\n", dep.Product, err)
				continue
			}
			return fmt.Errorf("dependency %s: %w", dep.Product, err)
		}
	}

	itemsByRelease := make(map[string]InstallItem, len(vs.Releases))
	releaseNames := make([]string, 0, len(vs.Releases))
	for releaseName, entry := range vs.Releases {
		chartName := entry.Chart
		if chartName == "" {
			chartName = releaseName
		}
		itemsByRelease[releaseName] = InstallItem{
			ReleaseName:  releaseName,
			ChartName:    chartName,
			ChartVersion: entry.Version,
			HelmRepoName: vs.Source.HelmRepoName,
			HelmRepoURL:  vs.Source.HelmRepoURL,
			Product:      vs.Product,
			Stage:        entry.Stage,
			DependsOn:    append([]string(nil), entry.DependsOn...),
			Values:       entry.Values, // 传递 per-release 自定义 values
		}
		releaseNames = append(releaseNames, releaseName)
	}
	sort.Strings(releaseNames)

	steps, err := buildManifestSteps(releaseNames, itemsByRelease)
	if err != nil {
		return fmt.Errorf("build release graph for product %s: %w", vs.Product, err)
	}
	if len(steps) > 0 {
		*allSteps = append(*allSteps, steps...)
	}

	return nil
}

func buildManifestSteps(releaseNames []string, itemsByRelease map[string]InstallItem) ([][]InstallItem, error) {
	indegree := make(map[string]int, len(releaseNames))
	adjacency := make(map[string]map[string]struct{}, len(releaseNames))
	preReleases := make([]string, 0)
	normalReleases := make([]string, 0)

	for _, releaseName := range releaseNames {
		indegree[releaseName] = 0
		adjacency[releaseName] = make(map[string]struct{})
		if itemsByRelease[releaseName].Stage == "pre" {
			preReleases = append(preReleases, releaseName)
		} else {
			normalReleases = append(normalReleases, releaseName)
		}
	}

	addEdge := func(from, to string) {
		if _, exists := adjacency[from][to]; exists {
			return
		}
		adjacency[from][to] = struct{}{}
		indegree[to]++
	}

	for _, releaseName := range releaseNames {
		item := itemsByRelease[releaseName]
		for _, dep := range item.DependsOn {
			if dep == releaseName {
				return nil, fmt.Errorf("release %q cannot depend on itself", releaseName)
			}
			if _, ok := itemsByRelease[dep]; !ok {
				return nil, fmt.Errorf("release %q has unknown dependency release %q", releaseName, dep)
			}
			dependencyItem := itemsByRelease[dep]
			dependencyItem.WaitForReady = true
			itemsByRelease[dep] = dependencyItem
			addEdge(dep, releaseName)
		}
	}

	for _, preRelease := range preReleases {
		preItem := itemsByRelease[preRelease]
		preItem.WaitForReady = true
		itemsByRelease[preRelease] = preItem
		for _, normalRelease := range normalReleases {
			addEdge(preRelease, normalRelease)
		}
	}

	var steps [][]InstallItem
	processed := 0
	ready := make([]string, 0)
	for _, releaseName := range releaseNames {
		if indegree[releaseName] == 0 {
			ready = append(ready, releaseName)
		}
	}

	for len(ready) > 0 {
		sort.Strings(ready)
		stepNames := append([]string(nil), ready...)
		step := make([]InstallItem, 0, len(stepNames))
		nextReady := make([]string, 0)
		nextReadySet := make(map[string]struct{})

		for _, releaseName := range stepNames {
			step = append(step, itemsByRelease[releaseName])
			processed++
			children := make([]string, 0, len(adjacency[releaseName]))
			for child := range adjacency[releaseName] {
				children = append(children, child)
			}
			sort.Strings(children)
			for _, child := range children {
				indegree[child]--
				if indegree[child] == 0 {
					if _, exists := nextReadySet[child]; !exists {
						nextReady = append(nextReady, child)
						nextReadySet[child] = struct{}{}
					}
				}
			}
		}

		steps = append(steps, step)
		ready = nextReady
	}

	if processed != len(releaseNames) {
		cycleNodes := make([]string, 0)
		for _, releaseName := range releaseNames {
			if indegree[releaseName] > 0 {
				cycleNodes = append(cycleNodes, releaseName)
			}
		}
		sort.Strings(cycleNodes)
		return nil, fmt.Errorf("dependency cycle detected among releases: %v", cycleNodes)
	}

	return steps, nil
}
