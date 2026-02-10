package push

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/mholt/archiver/v3"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
)

var log = logger.NewLogger()

type ImagePushOpts struct {
	Registry       string
	Username       string // used for set Authenticate config of docker
	Password       string // used for set Authenticate config of docker
	OCIPackagePath string
	PrePullImages  bool // used for set pre pull if need
}

func NewCr(opts ImagePushOpts) (*cr.Cr, error) {
	var clusterConf *configuration.ClusterConfig
	// If the registry is specified, it is treated as an external K8S + external cr processing,
	// else get infomation of cr from cluster config, otherwise, error will be return.
	if opts.Registry != "" {
		clusterConf = &configuration.ClusterConfig{
			Cs: &configuration.Cs{Provisioner: configuration.KubernetesProvisionerExternal},
			Cr: &configuration.Cr{
				External: &configuration.ExternalCR{
					Registry: &configuration.Registry{
						Host:     opts.Registry,
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
			return nil, err
		}
		clusterConf = c
	} else {
		log.Errorf("unable load old cluster conf: %v", client.ErrKubernetesClientSetNil)
		return nil, client.ErrKubernetesClientSetNil
	}
	cr := &cr.Cr{
		Logger:        log,
		ClusterConf:   clusterConf,
		PrePullImages: opts.PrePullImages && clusterConf.Cs.Provisioner == configuration.KubernetesProvisionerLocal,
	}
	return cr, nil
}

func PushImagesWithCr(crCli *cr.Cr, ociPkgPath string, workDir string) error {
	if fi, err := os.Stat(ociPkgPath); err != nil {
		return err
	} else if !fi.IsDir() {
		if workDir == "" {
			workDir = os.TempDir()
		} else {
			if err := os.MkdirAll(workDir, 0700); err != nil {
				return err
			}
		}

		// Decompress the compressed archive file
		dir, err := os.MkdirTemp(workDir, "images")
		log.Debugf("images oci dir: %s", dir)
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)
		if err = archiver.Unarchive(ociPkgPath, dir); err != nil {
			return errors.Join(ErrUnknowFile, fmt.Errorf("decompress %s failed: %v", ociPkgPath, err))
		}
		ociPkgPath = dir
	}
	log.Debugf("oci package: %v", ociPkgPath)
	if err := crCli.PushImages(ociPkgPath); err != nil {
		if errors.Is(err, cr.ErrSkopeoImagesFile) {
			return errors.Join(ErrUnknowFile, err)
		}
		return err
	}
	fmt.Printf("\033[1;37;42m%s\033[0m\n", "Push images success")
	return nil
}

func PushImages(opts ImagePushOpts, workDir string) error {
	cr, err := NewCr(opts)
	if err != nil {
		return err
	}

	return PushImagesWithCr(cr, opts.OCIPackagePath, workDir)
}
