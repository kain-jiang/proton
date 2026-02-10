package cr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
	moby_image "github.com/moby/moby/api/types/image"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	ecms "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

type Image struct {
	Logger *logrus.Logger
}

type LocalImageInfo struct {
	Image   string
	ID      string
	Created int64
}

type RegisterImageInfo struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// 定义函数获取私有仓库镜像列表
func (i *Image) getImageList(url string) ([]string, error) {
	// 最多返回仓库中的10000个镜像，默认值为100，当前镜像数量已经超过100
	resp, err := http.Get(url + "/v2/_catalog?n=10000")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Repositories []string `json:"repositories"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Repositories, nil
}

// 定义函数获取镜像标签
func (i *Image) getImageTags(url, name string) ([]string, error) {
	resp, err := http.Get(url + "/v2/" + name + "/tags/list")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Tags, nil
}

// GetPrivateRegister 查询私有仓库images >> registerImage.json
func (i *Image) GetPrivateRegister(registry string) ([]RegisterImageInfo, error) {
	url := fmt.Sprintf("http://%s", registry)
	imageList, err := i.getImageList(url)
	if err != nil {
		return nil, fmt.Errorf("Error getting image list: %w", err)
	}

	var images []RegisterImageInfo
	for _, image := range imageList {
		tags, err := i.getImageTags(url, image)
		if err != nil {
			i.Logger.Errorf("Error getting tags for image %s: %v\n", image, err)
			continue
		}
		images = append(images, RegisterImageInfo{Name: image, Tags: tags})
	}

	return images, nil
}

// GetRunningContainer 查询docker ps >> runningImages.json
func (i *Image) GetRunningContainer() ([]string, error) {
	var runningContainers []string
	_, k := client.NewK8sClient()
	pods, err := k.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	// 遍历每个Pod
	for _, pod := range pods.Items {
		// 获取Pod中的容器
		containers := pod.Spec.Containers
		initContainers := pod.Spec.InitContainers
		// 遍历每个容器
		for _, container := range containers {
			// 追加容器使用的镜像
			runningContainers = append(runningContainers, container.Image)
		}
		for _, container := range initContainers {
			// 追加初始化容器使用的镜像
			runningContainers = append(runningContainers, container.Image)
		}
	}
	return runningContainers, nil
}

// ReleaseDockerSpace 空间清理
func (i *Image) ReleaseDockerSpace(clusterConf *configuration.ClusterConfig, retainNum int) error {
	// 如果是外置k8s，则跳过
	if clusterConf.Cs.Provisioner == configuration.KubernetesProvisionerExternal && clusterConf.Cr.External != nil {
		return fmt.Errorf("External kubernetes cleaning is not supported\n")
	}
	// 如果是外置仓库，则跳过
	if clusterConf.Cr.External != nil {
		return fmt.Errorf("External regisery cleaning is not supported\n")
	}

	// 获取正在运行的镜像
	runningContainers, err := i.GetRunningContainer()
	if err != nil {
		return fmt.Errorf("unable to get running containers: %w", err)
	}
	if retainNum == 0 {
		return fmt.Errorf("not support delete all images,image retain num must greater than 0\n")
	} else if retainNum < 0 {
		return fmt.Errorf("invalid image retain num\n")
	}

	// 对registry进行清理
	unsupportImagesNodeMap := map[string][]string{}
	for _, nodeName := range clusterConf.Cr.Local.Hosts {
		registry := fmt.Sprintf("%s:%d", nodeName, clusterConf.Cr.Local.Ports.Registry)
		// 获取私有仓库的镜像
		registerImages, err := i.GetPrivateRegister(registry)
		if err != nil {
			return fmt.Errorf("unable to get private registry images: %w", err)
		}
		retainImages := map[string][]string{}
		unsupportImages := []string{}
		// find retain registry images
		for _, image := range registerImages {
			i.Logger.Debug("registerImages get : ", image.Name, ":  ", image.Tags)
			vs := []*semver.Version{}
			for _, tag := range image.Tags {
				v, err := semver.NewVersion(tag)
				if err != nil {
					i.Logger.Warnf("Image [%s],Error parsing version: %s", image.Name, err)
					i.Logger.Warnf("not support to delete registry image: [%s:%s],tag is not Semantic Versioning", image.Name, tag)
					unsupportImages = append(unsupportImages, image.Name+":"+tag)
					continue
				}
				vs = append(vs, v)
			}
			sort.Sort(semver.Collection(vs))
			unsupportImagesNodeMap[nodeName] = unsupportImages
			for _, v := range vs {
				if v == nil {
					break
				}
				retainImages[image.Name] = append(retainImages[image.Name], v.String())
			}
			if len(retainImages[image.Name]) > retainNum {
				retainImages[image.Name] = retainImages[image.Name][len(retainImages[image.Name])-retainNum:]
			}

			i.Logger.Debugf("retainImages: [%s]  %v", image.Name, retainImages[image.Name])
		}
		// 清理私有仓库中未使用的镜像
		i.Logger.Infof("%s: clean registry image begin", nodeName)
		err = i.cleanOldImages(registry, runningContainers, registerImages, retainImages, unsupportImages)
		if err != nil {
			return fmt.Errorf("failed to clean old images: %w", err)
		}
		i.Logger.Infof("%s: clean registry image end", nodeName)
	}

	var errList []error
	for _, node := range clusterConf.Nodes {
		i.Logger.Infof("%s: clean local image begin", node.IP())

		localImages, err := i.GetLocalImages(node)
		if err != nil {
			return err
		}

		var unusedImages []LocalImageInfo
		for _, localImage := range localImages {
			found := false
			for _, runningImage := range runningContainers {
				if localImage.Image == runningImage {
					found = true
					break
				}
			}
			if !found {
				unusedImages = append(unusedImages, localImage)
			}
		}
		// find retain local images
		retainImageMap := make(map[string]map[string]int64)
		retainImages := make(map[string][]LocalImageInfo)
		var repoRetainImages []LocalImageInfo
		for _, image := range localImages {
			lastIndex := strings.LastIndex(image.Image, ":")
			repoName := image.Image[:lastIndex]
			repoTag := image.Image[lastIndex+1:]
			if retainImageMap[repoName] == nil {
				retainImageMap[repoName] = make(map[string]int64)
			}
			retainImageMap[repoName][repoTag] = image.Created
		}
		for name, tags := range retainImageMap {
			repoRetainImages = []LocalImageInfo{}
			for tag, created := range tags {
				repoRetainImages = append(repoRetainImages, LocalImageInfo{Image: name + ":" + tag, Created: created})
			}
			repoRetainImages = FindNumByPartSort(repoRetainImages, retainNum)
			retainImages[name] = repoRetainImages
			i.Logger.Debugf("localName: [%s]      localRetainImages: <<%v>>", name, repoRetainImages)
		}
		var wg sync.WaitGroup
		for _, image := range unusedImages {
			var retain, exist = false, true
			if strings.Contains(image.Image, "registry") || strings.Contains(image.Image, "acr") {
				for repoName, repoRetainImage := range retainImages {
					imageSlice := strings.Split(image.Image, ":")
					imageNameSlice := strings.Split(imageSlice[len(imageSlice)-2], "/")
					imageName := imageNameSlice[len(imageNameSlice)-1]
					rNameSlice := strings.Split(repoName, "/")
					rName := rNameSlice[len(rNameSlice)-1]
					if imageName == rName {
						for i, retainImage := range repoRetainImage {
							if image.Image == retainImage.Image {
								retain = true
								break
							}
							if i == len(repoRetainImage)-1 {
								exist = false
							}
						}
					}
					if retain || !exist {
						break
					}
				}
				if retain {
					continue
				}
				nodeIp := node.IP()
				sshConf := client.RemoteClientConf{
					Host:     nodeIp,
					HostName: node.Name,
				}
				wg.Add(1)
				go func(sshConf client.RemoteClientConf, image LocalImageInfo) {
					defer wg.Done()
					if err := i.removeLocalImage(sshConf, []string{image.Image}); err != nil {
						errList = append(errList, fmt.Errorf("%s: %w", nodeIp, err))
					}
				}(sshConf, image)
			}
			wg.Wait()
		}
		// concurrent delete will trigger the ssh rate limiting，currently support delete one by one
		// wg.Wait()
		i.Logger.Infof("%s: clean local image end", node.IP())
	}
	for node, images := range unsupportImagesNodeMap {
		i.Logger.Warnf("ImagesNum: [%d],  Node: [%s],  Not Support Delete ImageRepository Images:  %v", len(images), node, images)
	}

	return utilerrors.NewAggregate(errList)
}

// ReleaseContainerdSpace 空间清理
func (i *Image) ReleaseContainerdSpace(clusterConf *configuration.ClusterConfig) error {
	// 如果是外置k8s，则跳过
	if clusterConf.Cs.Provisioner == configuration.KubernetesProvisionerExternal && clusterConf.Cr.External != nil {
		return fmt.Errorf("External kubernetes cleaning is not supported\n")
	}
	// 如果是外置仓库，则跳过
	if clusterConf.Cr.External != nil {
		return fmt.Errorf("External regisery cleaning is not supported\n")
	}

	return utilerrors.NewAggregate(nil)
}

// ReleaseSpace 空间清理
func (i *Image) ReleaseSpace(clusterConf *configuration.ClusterConfig) error {
	// 如果是外置k8s，则跳过
	if clusterConf.Cs.Provisioner == configuration.KubernetesProvisionerExternal && clusterConf.Cr.External != nil {
		return fmt.Errorf("External kubernetes cleaning is not supported\n")
	}
	// 如果是外置仓库，则跳过
	if clusterConf.Cr.External != nil {
		return fmt.Errorf("External regisery cleaning is not supported\n")
	}

	// 获取正在运行的镜像
	runningContainers, err := i.GetRunningContainer()
	if err != nil {
		return fmt.Errorf("unable to get running containers: %w", err)
	}

	// 对registry进行清理
	for _, nodeName := range clusterConf.Cr.Local.Hosts {
		registry := fmt.Sprintf("%s:%d", nodeName, clusterConf.Cr.Local.Ports.Registry)
		// 获取私有仓库的镜像
		registerImages, err := i.GetPrivateRegister(registry)
		if err != nil {
			return fmt.Errorf("unable to get private registry images: %w", err)
		}

		// 清理私有仓库中未使用的镜像
		err = i.cleanUnusedRegisterImages(registry, runningContainers, registerImages)
		if err != nil {
			return fmt.Errorf("failed to clean unused register images: %w", err)
		}
	}

	var errList []error
	for _, node := range clusterConf.Nodes {
		i.Logger.Infof("%s: clean local image begin", node.IP())

		localImages, err := i.GetLocalImages(node)
		if err != nil {
			return err
		}

		var unusedImages []LocalImageInfo
		for _, localImage := range localImages {
			found := false
			for _, runningImage := range runningContainers {
				if localImage.Image == runningImage {
					found = true
					break
				}
			}
			if !found {
				unusedImages = append(unusedImages, localImage)
			}
		}

		var wg sync.WaitGroup
		for _, image := range unusedImages {
			if strings.Contains(image.Image, "registry") || strings.Contains(image.Image, "acr") {
				nodeIp := node.IP()
				sshConf := client.RemoteClientConf{
					Host:     nodeIp,
					HostName: node.Name,
				}
				wg.Add(1)
				go func(sshConf client.RemoteClientConf) {
					defer wg.Done()
					if err := i.releaseLocalImage(sshConf, image.ID); err != nil {
						errList = append(errList, fmt.Errorf("%s: %w", nodeIp, err))
					}
				}(sshConf)
			}
			wg.Wait()
		}
		i.Logger.Infof("%s: clean local image end", node.IP())
	}

	return utilerrors.NewAggregate(errList)
}

func (i *Image) GetLocalImages(node configuration.Node) ([]LocalImageInfo, error) {
	var ecms = ecms.NewForHost(node.IP())
	var executor = exec.NewECMSExecutorForHost(ecms.Exec())

	var args []string
	var out []byte
	var err error

	// get images from remote docker daemon
	args = []string{"images", "--all", "--no-trunc", "--quiet"}
	if out, err = executor.Command("docker", args...).Output(); err != nil {
		return nil, fmt.Errorf("unable to get image ids from remote daemon: %w", err)
	}
	imageIDs := strings.Fields(string(out))

	// inspect remote image
	args = nil
	args = append(args, "inspect")
	args = append(args, imageIDs...)
	if out, err = executor.Command("docker", args...).Output(); err != nil {
		return nil, fmt.Errorf("unable to inspect remote image: %w", err)
	}

	// unmarshal inspect image response
	var responses []moby_image.InspectResponse
	if err := json.Unmarshal(out, &responses); err != nil {
		return nil, fmt.Errorf("unable to unmarshal inspect image response: %w", err)
	}

	var images []LocalImageInfo
	for _, resp := range responses {
		created, err := time.Parse(time.RFC3339, resp.Created)
		if err != nil {
			return nil, fmt.Errorf("unable to parse image %s created time: %w", resp.ID, err)
		}
		for _, t := range resp.RepoTags {
			images = append(images, LocalImageInfo{
				Image:   t,
				ID:      resp.ID,
				Created: created.Unix(),
			})
		}
	}

	return images, nil
}

func (i *Image) releaseLocalImage(sshConf client.RemoteClientConf, ID string) error {
	var ecms = ecms.NewForHost(sshConf.Host)
	var executor = exec.NewECMSExecutorForHost(ecms.Exec())
	if err := executor.Command("docker", "rmi", ID).Run(); err != nil {
		i.Logger.
			WithField("host", sshConf.Host).
			WithField("image", ID).
			Warn("unable to remove image from docker daemon")
	}
	return nil
}

func (i *Image) removeLocalImage(sshConf client.RemoteClientConf, IDs []string) error {
	var ecms = ecms.NewForHost(sshConf.Host)
	var executor = exec.NewECMSExecutorForHost(ecms.Exec())

	for _, id := range IDs {
		// remove local image gracefully
		if err := executor.Command("docker", "rmi", id).Run(); err != nil {
			i.Logger.Errorf("image remove error: %s", err)
			return err
		}
		i.Logger.Info("Local image [", id, "] deleted succeed")
	}
	return nil
}

func (i *Image) cleanUnusedRegisterImages(registry string, runningContainers []string, registerInfoList []RegisterImageInfo) error {
	// 获取到私有仓库未使用的镜像
	prefix := "registry.aishu.cn:15000"
	host := strings.Split(registry, ":")[0]
	var unusedRegisterInfoList []RegisterImageInfo

	var ecms = ecms.NewForHost(host)
	var executor = exec.NewECMSExecutorForHost(ecms.Exec())

	for _, registerImage := range registerInfoList {
		for _, tag := range registerImage.Tags {
			used := false
			for _, container := range runningContainers {
				if container == fmt.Sprintf("%s/%s:%s", prefix, registerImage.Name, tag) {
					used = true
					break
				}
			}
			if !used {
				unusedRegisterInfoList = append(unusedRegisterInfoList, RegisterImageInfo{
					Name: registerImage.Name,
					Tags: []string{tag},
				})
			}
		}

	}

	// 通过api进行删除
	for _, unusedRegisterImage := range unusedRegisterInfoList {
		for _, tag := range unusedRegisterImage.Tags {
			url := fmt.Sprintf("http://%s/v2/%s/manifests/%s", registry, unusedRegisterImage.Name, tag)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			// extract the digest from the response header
			digest := res.Header.Get("Docker-Content-Digest")

			// delete the image using the digest
			if digest != "" {
				url = fmt.Sprintf("http://%s/v2/%s/manifests/%s", registry, unusedRegisterImage.Name, digest)
				req, err = http.NewRequest("DELETE", url, nil)
				if err != nil {
					return err
				}
				res, err = http.DefaultClient.Do(req)
				if err != nil {
					return err
				}
				defer res.Body.Close()
				if res.StatusCode != http.StatusAccepted {
					// 可能会出现 tag 能查到，但是删除它的的manifest时返回 unknown ，暂时未找到原因
					i.Logger.Errorf("failed to delete image %s:%s (status code %d)，deleted or env problem,please check", unusedRegisterImage.Name, tag, res.StatusCode)
				}
			}

		}
	}

	// 垃圾回收
	sshConf := client.RemoteClientConf{
		Host:     host,
		HostName: registry,
	}

	i.Logger.Infof("%s: registry garbage begin", sshConf.Host)

	if err := executor.Command("/bin/registry", "garbage-collect", "/etc/docker/registry/config.yml").Run(); err != nil {
		i.Logger.Errorf("%s: registry garbage error", sshConf.Host)
		return err
	}
	i.Logger.Infof("%s: registry garbage succeed", sshConf.Host)

	return nil
}

func (i *Image) cleanOldImages(registry string, runningContainers []string, registerInfoList []RegisterImageInfo, retainImages map[string][]string, unsupportImages []string) error {
	// 获取到私有仓库未使用的镜像
	prefix := "registry.aishu.cn:15000"
	host := strings.Split(registry, ":")[0]
	var unusedRegisterInfoList []RegisterImageInfo

	var ecms = ecms.NewForHost(host)
	var executor = exec.NewECMSExecutorForHost(ecms.Exec())

	for _, registerImage := range registerInfoList {
		for _, tag := range registerImage.Tags {
			used := false
			for _, container := range runningContainers {
				if container == fmt.Sprintf("%s/%s:%s", prefix, registerImage.Name, tag) {
					used = true
					break
				}
			}
			for _, unsupportImage := range unsupportImages {
				if fmt.Sprintf("%s/%s", prefix, unsupportImage) == fmt.Sprintf("%s/%s:%s", prefix, registerImage.Name, tag) {
					i.Logger.Warnf("%s/%s    unsupport", prefix, unsupportImage)
					used = true
					break
				}
			}
			for _, retainTag := range retainImages[registerImage.Name] {
				if fmt.Sprintf("%s/%s:%s", prefix, registerImage.Name, retainTag) == fmt.Sprintf("%s/%s:%s", prefix, registerImage.Name, tag) {
					i.Logger.Debugf("%s/%s:%s    retain", prefix, registerImage.Name, retainTag)
					used = true
					break
				}
			}
			if !used {
				unusedRegisterInfoList = append(unusedRegisterInfoList, RegisterImageInfo{
					Name: registerImage.Name,
					Tags: []string{tag},
				})
			}
		}

	}

	// 通过api进行删除
	for _, unusedRegisterImage := range unusedRegisterInfoList {
		for _, tag := range unusedRegisterImage.Tags {
			err := func() error {
				url := fmt.Sprintf("http://%s/v2/%s/manifests/%s", registry, unusedRegisterImage.Name, tag)
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					return err
				}
				req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
				// compatible for skopeo oci iamge
				req.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json")
				res, err := http.DefaultClient.Do(req)
				if err != nil {
					return err
				}
				defer res.Body.Close()
				// extract the digest from the response header
				digest := res.Header.Get("Docker-Content-Digest")
				// delete the image using the digest
				if digest != "" {
					url = fmt.Sprintf("http://%s/v2/%s/manifests/%s", registry, unusedRegisterImage.Name, digest)
					req, err = http.NewRequest("DELETE", url, nil)
					if err != nil {
						return err
					}
					res, err = http.DefaultClient.Do(req)
					if err != nil {
						return err
					}
					defer res.Body.Close()
					if res.StatusCode != http.StatusAccepted {
						i.Logger.Errorf("failed to delete image %s:%s (status code %d)", unusedRegisterImage.Name, tag, res.StatusCode)
					} else {
						i.Logger.Infof("Image [%s:%s] delete succeed", unusedRegisterImage.Name, tag)
					}
				}
				return nil
			}()
			if err != nil {
				return err
			}
		}
	}

	// 垃圾回收
	sshConf := client.RemoteClientConf{
		Host:     host,
		HostName: registry,
	}

	i.Logger.Infof("%s: registry garbage begin", sshConf.Host)

	if err := executor.Command("/bin/registry", "garbage-collect", "/etc/docker/registry/config.yml").Run(); err != nil {
		i.Logger.Errorf("%s: registry garbage error", sshConf.Host)
		return err
	}
	i.Logger.Infof("%s: registry garbage succeed", sshConf.Host)

	return nil
}

// 利用部分排序寻找最大的k个数
func FindNumByPartSort(data []LocalImageInfo, k int) (result []LocalImageInfo) {
	// if data length less than or equal k ,return data
	if len(data) < k+1 {
		return data
	}
	// default first max k
	result = data[0:k]

	for i := k; i < len(data); i++ {
		// find min index
		min := select_sort(result)
		// compare
		if result[min].Created < data[i].Created {
			// swap
			result[min], data[i] = data[i], result[min]
		}
	}

	return result
}

func select_sort(data []LocalImageInfo) (min int) {
	min = 0
	for i := 1; i < len(data); i++ {
		if data[i].Created < data[min].Created {
			min = i
		}
	}
	return min
}
