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
	registry            string
	registryUsername    string
	registryPassword    string
	registryPlainHTTP   bool
	chartmuseumURL      string
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
		if opts.registry == "" {
			opts.registry = resolved.registry
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
		if opts.chartmuseumURL == "" {
			opts.chartmuseumURL = resolved.chartmuseumURL
		}
		if opts.chartmuseumUsername == "" {
			opts.chartmuseumUsername = resolved.chartmuseumUsername
		}
		if opts.chartmuseumPassword == "" {
			opts.chartmuseumPassword = resolved.chartmuseumPassword
		}
	}

	if strings.TrimSpace(opts.registry) == "" {
		return fmt.Errorf("target registry is required, set --registry or enable --auto")
	}
	if strings.TrimSpace(opts.chartmuseumURL) == "" {
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

	registryHost, registryUsername, registryPassword := global.ImageRepository(clusterCfg.Cr)
	if strings.TrimSpace(registryHost) == "" {
		return nil, fmt.Errorf("auto-detect import targets: current proton cluster has no image registry")
	}
	chartmuseumURL, chartmuseumUsername, chartmuseumPassword := global.Chartmuseum(clusterCfg.Cr)
	if strings.TrimSpace(chartmuseumURL) == "" {
		return nil, fmt.Errorf("auto-detect import targets: current proton cluster has no chartmuseum")
	}

	targets := &appImportAutoTargets{
		registry:            registryHost,
		registryUsername:    registryUsername,
		registryPassword:    registryPassword,
		chartmuseumURL:      chartmuseumURL,
		chartmuseumUsername: chartmuseumUsername,
		chartmuseumPassword: chartmuseumPassword,
	}
	if clusterCfg.Cr.External != nil && clusterCfg.Cr.External.ImageRepo == configuration.RepoOCI && clusterCfg.Cr.External.OCI != nil {
		targets.registryPlainHTTP = clusterCfg.Cr.External.OCI.PlainHTTP
	}

	return targets, nil
}
