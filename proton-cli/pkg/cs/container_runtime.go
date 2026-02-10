package cs

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	k "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cs/kubernetes"
)

// isSpecifiedContainerRuntimeSource 返回是否指定了容器运行时
func isSpecifiedContainerRuntimeSource(s *configuration.ContainerRuntimeSource) bool {
	switch {
	case s.Containerd != nil:
		return true
	case s.Docker != nil:
		return true
	default:
		return false
	}
}

// 用于标识节点已拥有的容器运行时
type nodeContainerRuntime string

const (
	// containerd
	nodeContainerRuntimeContainerd nodeContainerRuntime = "containerd"
	// docker
	nodeContainerRuntimeDocker nodeContainerRuntime = "docker"
)

// 容器运行时的 rpm 包名
const (
	rpmPackageNameContainerd string = "containerd"
	rpmPackageNameDockerCE   string = "docker-ce"
)

// detectNodeCommonContainerRuntime 探查所有节点共有的容器运行时。如果存在多个，按以下顺序返回：
//  1. docker
//  2. containerd
func detectNodeCommonContainerRuntime(kc *k.KubernetesCluster) (nodeContainerRuntime, error) {
	// found container runtimes
	found := sets.New[nodeContainerRuntime]()

	for _, n := range append(kc.Workers, kc.Masters...) {
		runtimes, err := detectNodeContainerRuntimes(&n)
		if err != nil {
			return "", err
		}
		found.Insert(runtimes...)
	}

	for _, r := range []nodeContainerRuntime{
		nodeContainerRuntimeDocker,
		nodeContainerRuntimeContainerd,
	} {
		if found.Has(r) {
			return r, nil
		}
	}

	return "", errors.New("container runtime not found")
}

func detectNodeContainerRuntimes(n *k.Node) (runtimes []nodeContainerRuntime, err error) {
	for _, item := range []struct {
		// container runtime
		r nodeContainerRuntime
		// container runtime rpm package name
		n string
	}{
		// containerd
		{
			r: nodeContainerRuntimeContainerd,
			n: rpmPackageNameContainerd,
		},
		// docker
		{
			r: nodeContainerRuntimeDocker,
			n: rpmPackageNameDockerCE,
		},
	} {
		if _, err := n.Query(item.n); err != nil {
			continue
		}
		runtimes = append(runtimes, item.r)
	}
	return
}

func generateContainerRuntimeSourceInto(r nodeContainerRuntime, target *configuration.ContainerRuntimeSource, localCR *configuration.LocalCR, bip string, dockerDataDir string) {
	switch r {
	case nodeContainerRuntimeContainerd:
		target.Containerd = generateContainerdContainerRuntimeSource(localCR)
	case nodeContainerRuntimeDocker:
		target.Docker = generateDockerContainerRuntimeSource(localCR, bip, dockerDataDir)
	default:
		return
	}
}

func generateContainerdContainerRuntimeSource(localCR *configuration.LocalCR) *configuration.ContainerdContainerRuntimeSource {
	s := &configuration.ContainerdContainerRuntimeSource{
		Root: "/sysvol/proton_data/cs_containerd_data",
		// TODO: generate structurally
		SandboxImage: fmt.Sprintf("%s/public/pause:3.6", net.JoinHostPort(global.RegistryDomain, strconv.Itoa(localCR.Ha_ports.Registry))),
	}

	var hosts []string
	hosts = append(hosts, net.JoinHostPort(global.RegistryDomain, strconv.Itoa(localCR.Ha_ports.Registry)))
	for _, h := range localCR.Hosts {
		hosts = append(hosts, net.JoinHostPort(h, strconv.Itoa(localCR.Ports.Registry)))
	}

	for _, h := range hosts {
		s.Registries = append(s.Registries, generateContainerdRegistryHostConfig(h))
	}

	return s
}

func generateContainerdRegistryHostConfig(host string) configuration.RegistryHostConfig {
	s := &url.URL{
		// registry.aishu.cn:15000 和各个 node:5000 的 registry 使用 http 协议
		Scheme: "http",
		Host:   host,
	}
	return configuration.RegistryHostConfig{
		Host:   host,
		Server: s.String(),
		HostConfigs: map[string]configuration.RegistryHostFileConfig{
			host: {
				SkipVerify: ptr.To(true),
			},
		},
	}
}

func generateDockerContainerRuntimeSource(localCR *configuration.LocalCR, bip string, dockerDataDir string) *configuration.DockerContainerRuntimeSource {
	var registry []string
	registry = append(registry, net.JoinHostPort(global.RegistryDomain, strconv.Itoa(localCR.Ha_ports.Registry)))
	for _, h := range localCR.Hosts {
		registry = append(registry, net.JoinHostPort(h, strconv.Itoa(localCR.Ports.Registry)))
	}

	return &configuration.DockerContainerRuntimeSource{
		DataDir:            dockerDataDir,
		BIP:                bip,
		InsecureRegistries: registry,
	}
}
