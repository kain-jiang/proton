package universal

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	ecms "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1/files"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

var log = logger.NewLogger()

// 仅用于清理服务数据目录。将数据目录(datapath)里的数据移动到备份数据目录下（datapath.bak）
func ClearDataDir(host, dataPath string) error {
	var ctx = context.TODO()

	var ecms = ecms.NewForHost(host)
	// ecms/v1alpha1 files interface
	var ecmsV1Alpha1Files files.Interface = ecms.Files()

	if info, err := ecmsV1Alpha1Files.Stat(ctx, dataPath); err != nil || !info.IsDir() {
		err = fmt.Errorf("host[%s] %s does not exist or is not a directory", host, dataPath)
		log.Warn(err)
		return err
	}

	log.Infof("host[%s] clear directory %s", host, dataPath)
	clearBySelfTip := fmt.Sprintf("You may need to clear %s manually.", dataPath)
	bakDir := strings.TrimSuffix(dataPath, "/") + ".bak"
	if err := ecmsV1Alpha1Files.Create(ctx, bakDir, true, nil); err != nil {
		err = fmt.Errorf("host[%s] create backup dir faild: %v. %s", host, err, clearBySelfTip)
		log.Error(err)
		return err
	}
	// move files to backup dir
	files, err := ecmsV1Alpha1Files.ListDirectory(ctx, dataPath)
	if err != nil {
		err = fmt.Errorf("host[%s] read the data path [%s] faild: %v. %s", host, dataPath, err, clearBySelfTip)
		log.Error(err)
		return err
	}
	for _, f := range files {
		// PosixRename will return SSH_FX_FAILURE if f is a dir and not empty.
		if f.IsDir() {
			if err := ecmsV1Alpha1Files.Delete(ctx, filepath.Join(bakDir, f.Name())); err != nil {
				log.Warnf("host[%s] clear remote dir[%s] failed, error: %v", host, filepath.Join(bakDir, f.Name()), err)
			}
		}
		if err := ecmsV1Alpha1Files.Rename(ctx, filepath.Join(dataPath, f.Name()), filepath.Join(bakDir, f.Name())); err != nil {
			err = fmt.Errorf("host[%s] move data to backup dir faild: %v. %s", host, err, clearBySelfTip)
			log.Error(err)
			return err
		}
	}
	return nil
}

// ClearDataDirViaNodeV1Alpha1 is the same implement of ClearDataDir via node/v1alpha1
//
// TODO: node/v1alpha1.Interface -> ecms/v1alpha1/files.Interface
func ClearDataDirViaNodeV1Alpha1(node v1alpha1.Interface, path string, logger logrus.FieldLogger) error {
	var ctx = context.TODO()
	logger.WithField("path", path).Debug("clear data directory")
	if info, err := node.ECMS().Files().Stat(ctx, path); errors.Is(err, fs.ErrNotExist) {
		logger.WithField("path", path).Info("data directory doesn't exist")
		return nil
	} else if err != nil {
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("%v is not a directory", path)
	}

	parent, base := filepath.Split(path)
	bakDir := filepath.Join(parent, base+".bak")
	if info, err := node.ECMS().Files().Stat(ctx, bakDir); err == nil {
		logger.WithField("path", bakDir).WithField("modTime", info.ModTime().Format(time.RFC3339)).Info("remove old backup directory")
		if err := node.ECMS().Files().Delete(ctx, bakDir); err != nil {
			return err
		}
	}
	logger.WithField("path", bakDir).Info("create backup directory")
	if err := node.ECMS().Files().Create(ctx, bakDir, true, nil); err != nil {
		return err
	}

	infos, err := node.ECMS().Files().ListDirectory(ctx, path)
	if err != nil {
		return err
	}
	for _, info := range infos {
		logger.WithField("path", filepath.Join(path, info.Name())).Info("backup file from data directory")
		if err := node.ECMS().Files().Rename(ctx, filepath.Join(path, info.Name()), filepath.Join(bakDir, info.Name())); err != nil {
			return err
		}
	}
	return nil
}

// ReconcileDataDirectory create data directory on the node if the given path
// isn't empty and doesn't exist.
func ReconcileDataDirectory(node v1alpha1.Interface, path string, logger logrus.FieldLogger) error {
	if path == "" {
		return nil
	}

	var ctx = context.TODO()

	logger.WithField("path", path).Info("reconcile data directory")
	if _, err := node.ECMS().Files().Stat(ctx, path); err == nil {
		logger.WithField("path", path).Debug("data directory already exists")
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	logger.WithField("path", path).Info("create data directory")
	return node.ECMS().Files().Create(ctx, path, true, nil)
}
