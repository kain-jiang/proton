package cr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"

	resty "github.com/go-resty/resty/v2"
	imagespec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/yaml"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	exec_v1alpha1 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"

	"github.com/containers/image/v5/pkg/docker/config"
	"github.com/containers/image/v5/types"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/provenance"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilsexec "k8s.io/utils/exec"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr/chart"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr/chartmuseum"
)

const ExternalChartRepositoryName = global.HelmRepo

const (
	HelmChartRepositoryLocal  = "local"
	HelmChartRepositoryStable = "stable"
)

var (
	ErrSkopeoImagesFile = fmt.Errorf("input images file is not skopeo oci images file or dir")
)

type Cr struct {
	Logger        *logrus.Logger
	ClusterConf   *configuration.ClusterConfig
	PrePullImages bool
	exec          utilsexec.Interface
}

func (c *Cr) Apply() error {
	c.Logger.Debug("cr setting")

	if c.ClusterConf.Cr.Local != nil {
		b, err := os.ReadFile(global.CrConfPath)
		if err != nil {
			return err
		}
		var crConf configuration.CrConf
		if err := yaml.Unmarshal(b, &crConf); err != nil {
			return err
		}

		crConf.Port = configuration.Port{
			Chartmuseum: c.ClusterConf.Cr.Local.Ports.Chartmuseum,
			Registry:    c.ClusterConf.Cr.Local.Ports.Registry,
			Rpm:         c.ClusterConf.Cr.Local.Ports.Rpm,
			Crmanager:   c.ClusterConf.Cr.Local.Ports.Cr_manager,
		}
		crConf.Storage = c.ClusterConf.Cr.Local.Storage
		hosts := c.ClusterConf.Nodes
		var wg sync.WaitGroup
		var errList []error
		for i := 0; i < len(hosts); i++ {
			var host string
			if hosts[i].IP4 != "" {
				host = hosts[i].IP4
			} else {
				host = hosts[i].IP6
			}
			sshConf := client.RemoteClientConf{
				Host:     host,
				HostName: hosts[i].Name,
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := c.setCr(sshConf, crConf); err != nil {
					errList = append(errList, fmt.Errorf("%s: %w", host, err))
				}
			}()
		}

		wg.Wait()
		if errList != nil {
			return utilerrors.NewAggregate(errList)
		}
	}
	if err := c.PushImages(filepath.Join(global.ServicePackage, "images")); err != nil {
		return fmt.Errorf("push docker images fail: %w", err)
	}
	if c.ClusterConf.ECeph != nil && len(c.ClusterConf.ECeph.Hosts) > 0 && !c.ClusterConf.ECeph.SkipECephUpdate {
		if err := c.PushImages(filepath.Join(global.ServicePackageECeph, "images")); err != nil {
			return fmt.Errorf("push ECeph related images fail: %w", err)
		}
		if err := c.PushCharts(filepath.Join(global.ServicePackageECeph, "charts")); err != nil {
			return fmt.Errorf("push ECeph related helm charts fail: %w", err)
		}
	}

	if err := c.PushCharts(filepath.Join(global.ServicePackage, "charts")); err != nil {
		return fmt.Errorf("push helm charts fail: %w", err)
	}

	return nil
}
func (c *Cr) PushImages(ociPkgPath string) error {
	// 设置镜像仓库的验证信息, skopeo 在 `${XDG_RUNTIME_DIR}/containers/auth.json` 不存在时会使用 docker cli 的配置文件 `${HOME}/.docker/config.json`。
	host, username, password := global.ImageRepository(c.ClusterConf.Cr)
	if username != "" && password != "" {
		if err := config.SetAuthentication(&types.SystemContext{AuthFilePath: DockerCLIConfigPath()}, host, username, password); err != nil {
			return fmt.Errorf("unable to set docker cli authentication: %w", err)
		}
	}

	// 如果未配置 exec 则使用真实的 exec
	if c.exec == nil {
		c.exec = utilsexec.New()
	}

	// 在日志中记录 skopeo 版本
	sv, err := GetSkopeoVersion(c.exec)
	if err != nil {
		return fmt.Errorf("unable to get skopeo version: %w", err)
	}
	c.Logger.Debugf("skopeo version: %s", sv)
	imageTags, err := GetImageTagsFromOCIPackage(ociPkgPath)
	if err != nil {
		return err
	}

	// 查找 registry endpoint, 'external' use the host, 'local' use nodeName:registryPort
	var registryHosts []string
	if c.ClusterConf.Cr.External != nil {
		registryHosts = append(registryHosts, host)
	} else {
		for _, nodeName := range c.ClusterConf.Cr.Local.Hosts {
			registryHosts = append(registryHosts, fmt.Sprintf("%s:%d", nodeName, c.ClusterConf.Cr.Local.Ports.Registry))

		}
	}

	// 使用skopeo上传镜像，二进制在/usr/bin下
	if len(registryHosts) > 0 {
		_, err := exec.LookPath("skopeo")
		if err != nil {
			c.Logger.Error("skopeo command not found, we need skopeo to push images")
		}
		c.Logger.Infof("push images to %v begin", registryHosts)
		for _, registry := range registryHosts {
			for _, imageTag := range imageTags {
				c.Logger.Infof("push image[%s] to %s", imageTag, registry)
				src := fmt.Sprintf("oci:%s:%s", ociPkgPath, imageTag)
				dest := fmt.Sprintf("docker://%s", filepath.Join(registry, strings.SplitN(imageTag, "/", 2)[1]))
				if err := RunSkopeoCopy(c.exec, src, dest, SkopeoCopyOptions{
					InsecurePolicy:              true,
					DisableDestinationTLSVerify: true,
					RetryTimes:                  3,
				}); err != nil {
					return fmt.Errorf("unable to copy contaienr images: %w", err)
				}
			}
		}
		c.Logger.Infof("push %d images to %v end", len(imageTags), registryHosts)
	} else {
		c.Logger.Info("cr node is empty, skip push images")
	}

	// pull images for all nodes only if pre-pull is true
	if !c.PrePullImages {
		return nil
	}
	return c.pullImagesForAllNodes(host, imageTags)
}

func (c *Cr) pullImagesForAllNodes(registry string, imageTags []string) error {
	var images []string
	// registry: registry.aishu.cn:15000
	// imageTags: []string{acr.aishu.cn/public/pause:3.6,...}
	// replace imageTags's "acr.aishu.cn" to "registry.aishu.cn:15000"
	for _, t := range imageTags {
		// 如果未找到 / 说明 t 的格式有问题
		i := strings.Index(t, "/")
		if i < 0 {
			c.Logger.WithField("tag", t).Warn("invalid image tag")
			continue
		}
		images = append(images, registry+t[i:])
	}

	// 所有节点同时拉取镜像
	g := new(errgroup.Group)
	for i := range c.ClusterConf.Nodes {
		n := &c.ClusterConf.Nodes[i]
		var executor = exec_v1alpha1.NewECMSExecutorForHost(v1alpha1.NewForHost(n.IP()).Exec())
		g.Go(func() error {
			gg := new(errgroup.Group)
			// 每个节点同时拉取 8 个镜像。因为 sshd 的 MaxSessions 的默认配置为 10。
			gg.SetLimit(8)
			for _, image := range images {
				gg.Go(func() error {
					c.Logger.WithFields(logrus.Fields{"node": n.Name, "image": image}).Info("pull image on node")
					if err := c.pullImageOnNode(executor, image); err != nil {
						return fmt.Errorf("pull image %s on node %s fail: %w", image, n.Name, err)
					}
					return nil
				})
			}

			return gg.Wait()
		})
	}

	return g.Wait()
}

func (c *Cr) PushCharts(dirPath string) error {

	files := []string{}
	dir, err := os.ReadDir(dirPath)
	if err != nil {
		c.Logger.Error(err)
		return err
	}

	suffix := strings.ToLower(".tgz") //匹配后缀

	for _, file := range dir {
		if file.IsDir() {
			continue //忽略目录
		}
		if filepath.Ext(strings.ToLower(file.Name())) != suffix {
			// 跳过不匹配后缀的文件
			continue
		}
		files = append(files, path.Join(dirPath, file.Name()))
		// TODO: use library instead of command
		args := []string{"inspect", "chart", filepath.Join(dirPath, file.Name())}
		out, err := RunHelm3Command(c.exec, args...)
		if err != nil {
			return fmt.Errorf("cannot execute helm3 inspect chart: %v", err)
		}
		var chartInfo configuration.ChartInfo
		if err := yaml.Unmarshal(out, &chartInfo); err != nil {
			return fmt.Errorf("cannot parse helm3 inspect chart output: %w", err)
		}
		global.ChartInfoList = append(global.ChartInfoList, chartInfo)
	}

	if c.ClusterConf.Cr.External != nil {
		// 外置CR仓库
		switch c.ClusterConf.Cr.External.ChartRepo {
		case configuration.RepoChartmuseum:
			return c.PushCharts2Chartmuseum(files)
		case configuration.RepoOCI:
			return c.PushCharts2OCI(files)
		case configuration.RepoDefault:
			return c.PushCharts2Chartmuseum(files)
		default:
			return fmt.Errorf("cannot analysis chart repository type: %s", c.ClusterConf.Cr.External.ChartRepo)
		}
	}

	var crNodes []string
	var postChartUrl = "http://%s:%d/api/charts"
	// 查找 cr 节点
	for i := 0; i < len(c.ClusterConf.Cr.Local.Hosts); i++ {
		crNodes = append(crNodes, c.getNodeIPbyNodeName(c.ClusterConf.Cr.Local.Hosts[i]))
	}

	// 上传 chart 到 cr 仓库
	chartClient := resty.New()
	var stdout bytes.Buffer
	for _, v := range files {
		// TODO: use library instead of command
		args := []string{"inspect", "chart", v}
		inspectOut, err := RunHelm3Command(c.exec, args...)
		if err != nil {
			return fmt.Errorf("cannot execute helm3 inspect chart: %v", err)
		}
		stdout.Write(inspectOut)
		var chartInfo configuration.ChartInfo
		if err := yaml.Unmarshal(stdout.Bytes(), &chartInfo); err != nil {
			return err
		}
		global.ChartInfoList = append(global.ChartInfoList, chartInfo)
		for _, crNodeIp := range crNodes {
			resp, err := chartClient.R().
				SetFile("chart", v).
				SetContentLength(true).
				Post(fmt.Sprintf(postChartUrl, crNodeIp, c.ClusterConf.Cr.Local.Ports.Chartmuseum))
			if err != nil {
				c.Logger.Error(err)
				return err
			}
			if !resp.IsSuccess() {
				if resp.StatusCode() == 409 {
					_, err = chartClient.R().
						Delete(fmt.Sprintf(postChartUrl, crNodeIp, c.ClusterConf.Cr.Local.Ports.Chartmuseum) + "/" + chartInfo.Name + "/" + chartInfo.Version)
					if err != nil {
						return err
					}
					resp, err = chartClient.R().
						SetFile("chart", v).
						SetContentLength(true).
						Post(fmt.Sprintf(postChartUrl, crNodeIp, c.ClusterConf.Cr.Local.Ports.Chartmuseum))
					if err != nil {
						return err
					}
					if resp.IsSuccess() {
						c.Logger.Debugf("push chart %s on %s success", v, crNodeIp)
					} else {
						c.Logger.Error(string(resp.Body()))
						return fmt.Errorf("push chart failed: %s", string(resp.Body()))
					}
				} else {
					c.Logger.Error(string(resp.Body()))
					return fmt.Errorf("push chart failed: %s", string(resp.Body()))
				}
			} else {
				c.Logger.Debugf("push chart %s on %s success", v, crNodeIp)
			}
		}
	}

	c.Logger.Infof("%d charts have been pushed", len(files))
	return nil
}

func (c *Cr) PushCharts2Chartmuseum(files []string) error {
	chartmuseumCli, err := chartmuseum.NewClient(global.Chartmuseum(c.ClusterConf.Cr))
	if err != nil {
		return fmt.Errorf("unable to create chartmusem client: %w", err)
	}
	for _, file := range files {
		// 获取本地 chart 文件的 digest
		digest, err := provenance.DigestFile(file)
		if err != nil {
			return fmt.Errorf("unable to caculate chart file digest: %w", err)
		}
		// 获取本地 chart 文件
		local, err := loader.Load(file)
		if err != nil {
			return fmt.Errorf("unable to load chart file: %w", err)
		}
		// 获取 chartmuseum 中的元数据，容忍 not found 错误
		index, err := chartmuseumCli.IndexFile()
		if err != nil {
			return fmt.Errorf("unable to get chartmuseum index: %w", err)
		}
		remote, err := index.Get(local.Metadata.Name, local.Metadata.Version)
		if err != nil && !chart.IsNotFound(err) {
			return fmt.Errorf("unable to get chart metadata from repository index: %w", err)
		}
		// 如果 chartmuseum 中已经存在且相同则跳过
		if remote != nil && remote.Digest == digest {
			c.Logger.Debugf("skip %s: already exists", file)
			continue
		}
		// 如果 chartmuseum 中已经存在则覆盖
		c.Logger.Debugf("push chart %s", file)
		if err := chartmuseumCli.PushChartFile(file, chartmuseum.PushOptions{Force: remote != nil}); err != nil {
			return fmt.Errorf("unable to push chart %q to chartmuseum: %w", file, err)
		}
	}
	c.Logger.Infof("%d charts have been pushed", len(files))
	return nil
}

func (c *Cr) PushCharts2OCI(files []string) error {
	helm3cli, err := helm3.NewCli("resource", c.Logger.WithField("helm", "v3"))
	if err != nil {
		return fmt.Errorf("create helm3 client failed: %w", err)
	}
	oci := c.ClusterConf.Cr.External.OCI
	for _, file := range files {
		err := helm3cli.PushChart(file, &helm3.OCIRegistryConfig{
			PlainHTTP: oci.PlainHTTP,
			Registry:  oci.Registry,
			Username:  oci.Username,
			Password:  oci.Password,
		})
		if err != nil {
			return fmt.Errorf("push chart failed: %s, err: %w", file, err)
		}
	}
	return nil
}

func (c *Cr) getNodeIPbyNodeName(name string) string {

	for i := 0; i < len(c.ClusterConf.Nodes); i++ {
		if c.ClusterConf.Nodes[i].Name == name {
			if c.ClusterConf.Nodes[i].IP4 != "" {
				return c.ClusterConf.Nodes[i].IP4
			} else {
				return c.ClusterConf.Nodes[i].IP6
			}
		}
	}
	return ""
}

// RunHelm3Command 调用 helm3 命令并返回输出
// 首先尝试从全局路径执行 helm3，如果失败则尝试从当前路径执行
func RunHelm3Command(execer utilsexec.Interface, args ...string) ([]byte, error) {
	// 首先尝试从系统路径执行 helm3
	out, err := execer.Command("helm3", args...).Output()
	if err == nil {
		return out, nil
	}

	// 如果系统路径上的 helm3 执行失败，尝试从当前路径执行
	out, err = execer.Command("./helm3", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("cannot execute helm3 command: %v, tried both system path and current directory", err)
	}
	return out, nil
}

func (c *Cr) Reset() error {
	c.Logger.Debug("skip reset cr")
	return nil
}
func (c *Cr) setCr(sshConf client.RemoteClientConf, crConf configuration.CrConf) error {
	var ctx = context.TODO()
	var executor = exec_v1alpha1.NewECMSExecutorForHost(v1alpha1.NewForHost(sshConf.Host).Exec())
	// ecms/v1alpha1/files.Interface
	var f = v1alpha1.NewForHost(sshConf.Host).Files()
	if err := f.Create(ctx, crConf.Storage, true, nil); err != nil {
		return err
	}
	out, err := yaml.Marshal(crConf)
	if err != nil {
		return err
	}
	if err := f.Create(ctx, global.CrConfPath, false, out); err != nil {
		return err
	}
	c.Logger.Infof("restart proton-cr begin on node: %s", sshConf.Host)
	if err := executor.Command("systemctl", "enable", "proton-cr").Run(); err != nil {
		return err
	}
	if err := executor.Command("systemctl", "restart", "proton-cr").Run(); err != nil {
		return err
	}
	if err := executor.Command("systemctl", "enable", "docker").Run(); err != nil {
		return err
	}
	if err := executor.Command("systemctl", "start", "docker").Run(); err != nil {
		return err
	}

	time.Sleep(10 * time.Second)

	if c.ClusterConf.Cr.UseChartmuseum() {
		repositoryUrl, repositoryUsername, repositoryPassword := global.Chartmuseum(c.ClusterConf.Cr)
		if err := executor.Command("helm3", "repo", "add", global.HelmRepo, repositoryUrl, "--username", repositoryUsername, "--password", repositoryPassword).Run(); err != nil {
			return err
		}
	}
	c.Logger.Infof("restart proton-cr end on node: %s", sshConf.Host)

	return nil
}

// read index.json which stored in ociPkgPath, return the tags of images those recorded in index.json
func GetImageTagsFromOCIPackage(ociPkgPath string) ([]string, error) {
	var imageTags []string
	fileData, err := os.ReadFile(filepath.Join(ociPkgPath, "index.json"))
	if err != nil {
		return nil, errors.Join(ErrSkopeoImagesFile, err)
	}
	ociIndex := imagespec.Index{}
	if err := json.Unmarshal(fileData, &ociIndex); err != nil {
		return nil, errors.Join(ErrSkopeoImagesFile, fmt.Errorf("Unmarshal bytes in index.json failed: %v", err))
	}
	for _, descriptor := range ociIndex.Manifests {
		if refName, ok := descriptor.Annotations[imagespec.AnnotationRefName]; ok {
			imageTags = append(imageTags, refName)
		}
	}
	return imageTags, nil
}

// pullImagesOnNode 在指定节点拉取镜像
func (cr *Cr) pullImageOnNode(e exec_v1alpha1.Executor, image string) error {
	err := e.Command("crictl", "pull").Run()
	if err == nil {
		return nil
	}

	entry := cr.Logger.WithError(err)
	if ee := new(exec_v1alpha1.ErrExitError); errors.As(err, &ee) {
		cr.Logger.WithError(err).WithField("out", string(ee.Stderr)).Error("remote pull image fail")
		entry = entry.WithField("out", string(ee.Stderr))
	}
	entry.Error("remote pull image fail")
	return err
}
