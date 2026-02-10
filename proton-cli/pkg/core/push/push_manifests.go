package push

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"
)

const _DeployInstllerManifestsMagic = "deploy-compose-manifest"

type DeployInstallerManifest struct {
	*deployInstaller
	url string
}

type compseManifestMeta struct {
	Name    string `json:"mname"`
	Version string `json:"mversion"`
}

func NewDeployInstallerManifest(installer *deployInstaller, k kubernetes.Interface) deployInstallerUploader {
	return &DeployInstallerManifest{
		deployInstaller: installer,
		url:             newInstallerResourceUrl(installer, k, "manifests"),
	}
}

func (c *DeployInstallerManifest) CheckFile(fpath string) error {
	_, err := c.checkFile(fpath)
	return err
}

func (c *DeployInstallerManifest) checkFile(fpath string) ([]byte, error) {
	tarFilePath := fpath

	// 打开 tar 文件
	file, err := os.Open(tarFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	stats, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if stats.Size() > 1024*1024*3 {
		return nil, errors.Join(ErrUnknowFile, fmt.Errorf("文件过大,非manifests"))
	}

	bs, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	bsdecode, err := hex.DecodeString(string(bs))
	if err != nil {
		return nil, errors.Join(ErrUnknowFile, fmt.Errorf("文件内容无法解码"), err)
	}
	content := string(bsdecode)
	lenFile := len(content)
	lenMagic := len(_DeployInstllerManifestsMagic)

	if lenFile <= 2*lenMagic {
		return nil, ErrUnknowFile
	}

	if _DeployInstllerManifestsMagic != content[:lenMagic] ||
		_DeployInstllerManifestsMagic != content[lenFile-lenMagic:] {
		return nil, errors.Join(ErrUnknowFile, fmt.Errorf("文件内容未包含magic"))
	}
	return bsdecode[lenMagic : lenFile-lenMagic], nil
}

func (c *DeployInstallerManifest) Upload(ctx context.Context, log *logrus.Logger, fpath string) (string, error) {
	data, err := c.checkFile(fpath)
	if err != nil {
		return "", err
	}
	meta := &compseManifestMeta{}
	if err := json.Unmarshal(data, meta); err != nil {
		return "", errors.Join(ErrUnknowFile, fmt.Errorf("文件内容格式错误,无法解码"))
	}
	sucessMsg := fmt.Sprintf("包名: '%s', 包版本: '%s'", meta.Name, meta.Version)

	return sucessMsg, deployInstallerUpload(ctx, log, c.khttpCli, http.MethodPost, c.url, data, fpath)
}
