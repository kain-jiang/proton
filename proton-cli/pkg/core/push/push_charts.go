package push

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
)

type ChartPushOpts struct {
	HelmRepo  string
	Username  string // Account name on repo, used for Authentication when push charts
	Password  string // Account password on repo, used for Authentication when push charts
	ChartsDir string
}

func PushCharts(opts ChartPushOpts) error {
	var clusterConf *configuration.ClusterConfig
	// If the helm repo is specified, it is treated as an external K8S + external cr processing,
	// else get infomation of cr from cluster config, otherwise, error will be return.
	if opts.HelmRepo != "" {
		clusterConf = &configuration.ClusterConfig{
			Cs: &configuration.Cs{Provisioner: configuration.KubernetesProvisionerExternal},
			Cr: &configuration.Cr{
				External: &configuration.ExternalCR{
					ChartRepo: configuration.RepoChartmuseum,
					Chartmuseum: &configuration.Chartmuseum{
						Host:     opts.HelmRepo,
						Username: opts.Username,
						Password: opts.Password,
					},
				},
			},
		}
	} else if _, k := client.NewK8sClient(); k != nil {
		c, err := configuration.LoadFromKubernetes(context.Background(), k, "")
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			log.Errorf("unable load old cluster conf: %v", err)
			return err
		}
		clusterConf = c
	} else {
		log.Errorf("unable load old cluster conf: %v", client.ErrKubernetesClientSetNil)
		return client.ErrKubernetesClientSetNil
	}
	cr := &cr.Cr{
		Logger:      log,
		ClusterConf: clusterConf,
	}
	if err := cr.PushCharts(opts.ChartsDir); err != nil {
		return err
	}
	fmt.Printf("\033[1;37;42m%s\033[0m\n", "Push charts success")
	return nil
}
