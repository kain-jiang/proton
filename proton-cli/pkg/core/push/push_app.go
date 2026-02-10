package push

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/kubernetes"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
)

var (
	ErrAPPConflict                 = errors.New("application has been upload")
	ErrUnknowFile                  = errors.New("unknow file")
	ErrDeployInstallerNotInstalled = errors.New("deploy installer not installed")
)

type deployInstallerUploader interface {
	Upload(ctx context.Context, log *logrus.Logger, fpath string) (string, error)
	CheckFile(fpath string) error
}

type DeployInstaller struct {
	uploader []deployInstallerUploader
}

func (c *DeployInstaller) Upload(ctx context.Context, log *logrus.Logger, fpath string) (string, error) {
	var err error
	var msg string
	for _, i := range c.uploader {
		if msg, err = i.Upload(ctx, log, fpath); err != nil {
			if !errors.Is(err, ErrUnknowFile) {
				return msg, err
			}
		} else {
			return msg, nil
		}
	}
	return msg, err
}

func (c *DeployInstaller) CheckFile(fpath string) error {
	var err error
	for _, i := range c.uploader {
		if err = i.CheckFile(fpath); err != nil {
			if !errors.Is(err, ErrUnknowFile) {
				return err
			}
		} else {
			return nil
		}
	}
	return err
}

func newDeployInstaller(ctx context.Context, namespace string) (*deployInstaller, kubernetes.Interface, error) {
	khcli, err := client.NewK8sHTTPClient()
	if err != nil {
		return nil, nil, err
	}
	_, k := client.NewK8sClientInterface()
	if k == nil {
		return nil, nil, ErrDeployInstallerNotInstalled
	}
	svc, err := k.CoreV1().Services(namespace).Get(ctx, "deploy-installer", v1.GetOptions{})
	client := &deployInstaller{
		Namespace: namespace,
		khttpCli:  khcli,
	}

	if kerrors.IsNotFound(err) {
		return nil, nil, ErrDeployInstallerNotInstalled
	} else if err != nil {
		return nil, nil, err
	}
	client.Service = svc.Spec.ClusterIP

	return client, k, nil
}

func NewDeployInstallerClient(ctx context.Context, namespace string, onlyCheck bool) (*DeployInstaller, error) {
	uploaders := &DeployInstaller{}
	if onlyCheck {
		uploaders.uploader = []deployInstallerUploader{
			&DeployInstallerManifest{},
			&DeployInstallerApp{},
		}
		return uploaders, nil
	}
	client, k, err := newDeployInstaller(ctx, namespace)
	if err != nil {
		return nil, err
	}
	buiders := []func(installer *deployInstaller, k kubernetes.Interface) deployInstallerUploader{
		NewDeployInstallerManifest,
		NewDeployInstallerApp,
	}

	for _, b := range buiders {
		uploaders.uploader = append(uploaders.uploader, b(client, k))
	}

	return uploaders, nil
}

func newInstallerResourceUrl(installer *deployInstaller, k kubernetes.Interface, obj string) string {
	return k.CoreV1().RESTClient().Get().
		Namespace(installer.Namespace).Resource("services").
		SubResource("proxy").
		Name(net.JoinSchemeNamePort("http", "deploy-installer", "9090")).
		Suffix(fmt.Sprintf("internal/api/deploy-installer/v1/%s", obj)).URL().String()
}

func deployInstallerUpload(ctx context.Context, log *logrus.Logger, cli *http.Client, method, url string, bs []byte, fpath string) error {
	res, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bs))
	if err != nil {
		return err
	}
	resp, err := cli.Do(res)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 409 {
		return ErrAPPConflict
	} else if resp.StatusCode == 400 {
		respStr, err0 := io.ReadAll(resp.Body)
		if err0 != nil {
			log.Errorf("upload application file error, http statusCode: %q, read response error: %q", resp.StatusCode, err0)
			return err0
		}
		return fmt.Errorf("upload application file error, http statusCode: %q, response %q mean client tool and deploy-installer server are incompatible. Please don't use this tool to upload %q. Please try with deploy-installer pod's command client", resp.StatusCode, string(respStr), fpath)
	} else if resp.StatusCode != 200 {
		respStr, err0 := io.ReadAll(resp.Body)
		if err0 != nil {
			log.Errorf("upload application file error, http statusCode: %q, read response error: %q", resp.StatusCode, err0)
			return err0
		}
		return fmt.Errorf("upload application file error, http statusCode: %q, read response : %q", resp.StatusCode, string(respStr))
	}

	return nil
}

// deployApplicationMeta deploy meta data
// this is a application overview
type deployApplicationMeta struct {
	// 应用定义类型与版本,为未来预留
	Type string `json:"appDefineType"`
	// 应用包ID
	AID int `json:"aid"`
	// 应用包版本
	Version string `json:"version"`
	// 应用包名称
	AName string `json:"name"`
}

type deployInstaller struct {
	Service   string
	Namespace string
	khttpCli  *http.Client
}

// DeployInstaller 上传应用定义包到deploy-installer
// 从集群内通过clusterIP,使用内部无鉴权HTTP接口
type DeployInstallerApp struct {
	deployInstaller
	url string
}

func (c *DeployInstallerApp) CheckFile(fpath string) error {
	_, err := c.checkFile(fpath)
	return err
}

func (c *DeployInstallerApp) checkFile(fpath string) (*deployApplicationMeta, error) {
	tarFilePath := fpath
	fileName := "appMeta.json"

	// 打开 tar 文件
	file, err := os.Open(tarFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size() > 1024*1024*3 {
		return nil, errors.Join(ErrUnknowFile, fmt.Errorf("文件过大,非manifests"))
	}

	ziper, err := gzip.NewReader(file)
	if err != nil {
		return nil, errors.Join(ErrUnknowFile, err)
	}
	defer ziper.Close()
	// 创建 tar Reader
	tarReader := tar.NewReader(ziper)

	// 遍历 tar 文件中的文件
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Join(ErrUnknowFile, err)
		}

		// 检查文件名是否匹配
		if header.Name == fileName {
			// 读取文件内容
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, err
			}
			deployMeta := &deployApplicationMeta{}

			if err := json.Unmarshal(data, deployMeta); err != nil {
				return nil, errors.Join(ErrUnknowFile, fmt.Errorf("unmarshal bytes in %q failed: %w", fileName, err))
			}
			if deployMeta.AName == "" || deployMeta.Type != "app/v1betav1" {
				return nil, errors.Join(ErrUnknowFile, fmt.Errorf("deploy application meta error with name %q or appDefineType %q", deployMeta.AName, deployMeta.Type))
			}
			return deployMeta, nil
		}
	}
	return nil, errors.Join(ErrUnknowFile, fmt.Errorf("could not find %q in input file %q", fileName, fpath))
}

func (c *DeployInstallerApp) Upload(ctx context.Context, log *logrus.Logger, fpath string) (string, error) {
	meta, err := c.checkFile(fpath)
	if err != nil {
		return "", err
	}
	sucessMsg := fmt.Sprintf("包名: %s, 包版本: %s", meta.AName, meta.Version)

	fin, err := os.Open(fpath)
	if err != nil {
		return sucessMsg, err
	}
	defer fin.Close()
	bs, err := io.ReadAll(fin)
	if err != nil {
		return sucessMsg, err
	}
	return sucessMsg, deployInstallerUpload(ctx, log, c.khttpCli, http.MethodPost, c.url, bs, fpath)
}

func NewDeployInstallerApp(installer *deployInstaller, k kubernetes.Interface) deployInstallerUploader {
	return &DeployInstallerApp{
		deployInstaller: *installer,
		url:             newInstallerResourceUrl(installer, k, "application"),
	}
}
