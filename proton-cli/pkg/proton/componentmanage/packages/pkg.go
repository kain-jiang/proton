package packages

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/push"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

type ComponentPackage struct {
	File    string
	Logger  *logrus.Logger
	WorkDir string
	Charts  servicepackage.Charts
	Images  []string
}

func NewPackage(file string, log *logrus.Logger, workdir string) *ComponentPackage {
	if workdir == "" {
		workdir = os.TempDir()
	}

	return &ComponentPackage{
		File:    file,
		Logger:  log,
		WorkDir: workdir,
	}
}

func (p *ComponentPackage) Push(cfg *configuration.ClusterConfig, pull bool) error {
	dir, err := os.MkdirTemp(p.WorkDir, "component-package")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	p.Logger.Infof("start unarchive %s to %s...", p.File, dir)
	if err := archiver.Unarchive(p.File, dir); err != nil {
		return fmt.Errorf("decompress %s failed: %v", p.File, err)
	}

	cr := &cr.Cr{
		Logger:        p.Logger,
		ClusterConf:   cfg,
		PrePullImages: pull,
	}
	return errors.Join(
		cr.PushCharts(filepath.Join(dir, "charts")),
		push.PushImagesWithCr(cr, filepath.Join(dir, "images"), filepath.Join(dir, "images-temp")),
		p.ResolveInfomations(dir),
	)
}

func (p *ComponentPackage) ResolveInfomations(rootDir string) (err error) {
	pkg := new(servicepackage.ServicePackage)
	err = pkg.Load(rootDir)
	if err != nil {
		return
	}
	p.Charts = pkg.Charts()
	p.Images, err = cr.GetImageTagsFromOCIPackage(filepath.Join(rootDir, "images"))
	return
}
