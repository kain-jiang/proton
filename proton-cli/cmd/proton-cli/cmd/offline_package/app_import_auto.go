package offline_package

import (
	"context"
	"fmt"
	"strings"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
)

type appImportAutoTargets struct {
	registries          []string
	registryUsername    string
	registryPassword    string
	registryPlainHTTP   bool
	chartmuseumURLs     []string
	chartmuseumUsername string
	chartmuseumPassword string
}

var loadAppImportAutoTargetsFunc = loadAppImportAutoTargets

func hydrateAppImportOptions(ctx context.Context, opts *appImportOptions) error {
	if opts.auto {
		resolved, err := loadAppImportAutoTargetsFunc(ctx)
		if err != nil {
			return err
		}
		if len(opts.registries) == 0 && opts.registry == "" && len(resolved.registries) > 0 {
			opts.registries = resolved.registries
		}
		if opts.registry == "" && len(resolved.registries) > 0 {
			opts.registry = resolved.registries[0]
		}
		if opts.registryUsername == "" {
			opts.registryUsername = resolved.registryUsername
		}
		if opts.registryPassword == "" {
			opts.registryPassword = resolved.registryPassword
		}
		if !opts.registryPlainHTTP && resolved.registryPlainHTTP {
			opts.registryPlainHTTP = true
		}
		if len(opts.chartmuseumURLs) == 0 && opts.chartmuseumURL == "" && len(resolved.chartmuseumURLs) > 0 {
			opts.chartmuseumURLs = resolved.chartmuseumURLs
		}
		if opts.chartmuseumURL == "" && len(resolved.chartmuseumURLs) > 0 {
			opts.chartmuseumURL = resolved.chartmuseumURLs[0]
		}
		if opts.chartmuseumUsername == "" {
			opts.chartmuseumUsername = resolved.chartmuseumUsername
		}
		if opts.chartmuseumPassword == "" {
			opts.chartmuseumPassword = resolved.chartmuseumPassword
		}
	}

	if len(opts.registries) == 0 && strings.TrimSpace(opts.registry) == "" {
		return fmt.Errorf("target registry is required, set --registry or enable --auto")
	}
	if len(opts.chartmuseumURLs) == 0 && strings.TrimSpace(opts.chartmuseumURL) == "" {
		return fmt.Errorf("target chartmuseum is required, set --chartmuseum-url or enable --auto")
	}
	return nil
}

func loadAppImportAutoTargets(ctx context.Context) (*appImportAutoTargets, error) {
	_, k := client.NewK8sClient()
	if k == nil {
		return nil, fmt.Errorf("auto-detect import targets: kubernetes client is unavailable")
	}

	clusterCfg, err := configuration.LoadFromKubernetes(ctx, k)
	if err != nil {
		return nil, fmt.Errorf("auto-detect import targets: load current proton cluster config: %w", err)
	}
	return appImportAutoTargetsFromConfig(clusterCfg)
}

func appImportAutoTargetsFromConfig(clusterCfg *configuration.ClusterConfig) (*appImportAutoTargets, error) {
	if clusterCfg == nil || clusterCfg.Cr == nil {
		return nil, fmt.Errorf("auto-detect import targets: current proton cluster config has no cr settings")
	}
	if !clusterCfg.Cr.UseChartmuseum() {
		return nil, fmt.Errorf("auto-detect import targets: current proton cluster chart repository is not chartmuseum")
	}

	registryHosts, registryUsername, registryPassword := global.ImageRepositoryMulti(clusterCfg.Cr)
	if len(registryHosts) == 0 {
		return nil, fmt.Errorf("auto-detect import targets: current proton cluster has no image registry")
	}
	chartmuseumURLs, chartmuseumUsername, chartmuseumPassword := global.ChartmuseumMulti(clusterCfg.Cr)
	if len(chartmuseumURLs) == 0 {
		return nil, fmt.Errorf("auto-detect import targets: current proton cluster has no chartmuseum")
	}

	targets := &appImportAutoTargets{
		registries:          registryHosts,
		registryUsername:    registryUsername,
		registryPassword:    registryPassword,
		chartmuseumURLs:     chartmuseumURLs,
		chartmuseumUsername: chartmuseumUsername,
		chartmuseumPassword: chartmuseumPassword,
	}
	if clusterCfg.Cr.External != nil && clusterCfg.Cr.External.ImageRepo == configuration.RepoOCI && clusterCfg.Cr.External.OCI != nil {
		targets.registryPlainHTTP = clusterCfg.Cr.External.OCI.PlainHTTP
	}

	return targets, nil
}
