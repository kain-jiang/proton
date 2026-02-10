package configuration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
)

const (
	ProtonClusterDefaultNamespace   = "resource"
	ProtonCliConfigDefaultNamespace = "proton"
)

// Load 优先从 Kubernetes ConfigMap载入 Cluster Config，如果载入失败再从文件系统载入
func Load(ctx context.Context, k kubernetes.Interface, path string, namespace ...string) (*ClusterConfig, error) {
	if cfg, err := LoadFromKubernetes(ctx, k, namespace...); err == nil {
		return cfg, nil
	}
	return LoadFromFile(path)
}

func LoadFromKubernetes(ctx context.Context, k kubernetes.Interface, namespace ...string) (*ClusterConfig, error) {
	// If namespace is specified, use it; otherwise, use the default from file
	nsToUse := GetProtonCliConfigNSFromFile()
	if len(namespace) > 0 && namespace[0] != "" {
		nsToUse = namespace[0]
	}
	// Log the namespace being used for debugging
	cm, err := k.CoreV1().Secrets(nsToUse).Get(ctx, ProtonCLIConfigSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return LoadFromBytes([]byte(cm.Data[ClusterConfigurationConfigMapKey]))
}

func LoadFromFile(path string) (*ClusterConfig, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadFromBytes(bytes)
}

func LoadFromBytes(bytes []byte) (*ClusterConfig, error) {
	cfg := new(ClusterConfig)
	if err := yaml.Unmarshal(bytes, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// 补全 $HOME/.proton-cli.yaml 命名空间的配置
// 如果不存在该文件，先判断是否存在 proton 命名空间，如果存在则是旧集群，直接沿用该配置并写到上面的文件里面
// 如果存在该文件，且命名空间也存在，则是新部署模式（单个命名空间）
func UpdateProtonCliEnvConfig(newNamespace string) error {
	fmt.Println("update deploy namespace")
	var existProtonCliConfigNs = true
	var k8sConnected = true
	if _, k := client.NewK8sClient(); k != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := k.CoreV1().Namespaces().Get(ctx, ProtonCliConfigDefaultNamespace, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				existProtonCliConfigNs = false
			} else {
				existProtonCliConfigNs = false
				k8sConnected = false
			}
		}
	} else {
		// 全新的环境，或k8s无法连上
		existProtonCliConfigNs = false
		k8sConnected = false
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("not found home dir %v", err)
	}
	filePath := filepath.Join(homeDir, ".proton-cli.yaml")
	var config = ProtonCliEnvConfig{}
	config.ResourceNamespace = ProtonClusterDefaultNamespace
	config.ProtonCliConfigNamespace = ProtonCliConfigDefaultNamespace
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if !existProtonCliConfigNs && newNamespace != "" {
			// 不存在配置文件，也不存在 proton 命名空间，且指定了命名空间部署，表示是新部署模式
			config.ResourceNamespace = newNamespace
			config.ProtonCliConfigNamespace = newNamespace
		}
		if err := writeConfig(filePath, &config); err != nil {
			return fmt.Errorf("update proton cli config namespace failed: %v", err)
		}
	} else {
		// 存在配置文件，也存在 proton 命名空间，表示旧模式，升级过配置文件，跳过
		// 存在配置文件，不存在 proton 命名空间，表示新模式，升级过配置文件，跳过
		// 存在配置文件，但新命名空间与旧命名空间不一致，报错，不允许修改命名空间
		// 存在配置文件，但k8s连不上，可以覆盖本地配置
		c, err := readConfig(filePath)
		if err != nil {
			return fmt.Errorf("update proton cli config %s failed: %v", filePath, err)
		}
		// If a new namespace is specified, update the configuration
		if newNamespace != "" {
			// If K8s is not connected or the namespace is different, update the configuration
			if !k8sConnected || c.ResourceNamespace != newNamespace {
				config.ResourceNamespace = newNamespace
				config.ProtonCliConfigNamespace = newNamespace
				if err := writeConfig(filePath, &config); err != nil {
					return fmt.Errorf("update proton cli config namespace failed: %v", err)
				}
				return nil
			}
		}
		// If K8s is connected and trying to change to a different namespace, don't allow it
		if k8sConnected && c.ResourceNamespace != newNamespace && newNamespace != "" {
			return fmt.Errorf("exist resource namespace %s cannot change to new resource name %s", config.ResourceNamespace, newNamespace)
		}
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("not found proton-cli.yaml")
	}
	return nil
}

// 读取proton-cli.yaml配置文件
func GetProtonResourceNSFromFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ProtonClusterDefaultNamespace
	}
	filePath := filepath.Join(homeDir, ".proton-cli.yaml")
	var config = ProtonCliEnvConfig{}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return ProtonClusterDefaultNamespace
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return ProtonClusterDefaultNamespace
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return ProtonClusterDefaultNamespace
	}

	return config.ResourceNamespace
}

func GetProtonCliConfigNSFromFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ProtonCliConfigDefaultNamespace
	}
	filePath := filepath.Join(homeDir, ".proton-cli.yaml")
	var config = ProtonCliEnvConfig{}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return ProtonCliConfigDefaultNamespace
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return ProtonCliConfigDefaultNamespace
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return ProtonCliConfigDefaultNamespace
	}

	return config.ProtonCliConfigNamespace
}

// 读取proton-cli.yaml配置文件
func readConfig(filePath string) (*ProtonCliEnvConfig, error) {
	var config = ProtonCliEnvConfig{}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &config, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %v", filePath, err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal %s: %v", filePath, err)
	}

	return &config, nil
}

// 写入proton-cli.yaml配置文件
func writeConfig(filePath string, config *ProtonCliEnvConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("yaml marshal failed: %v", err)
	}
	fmt.Println("write proton config")
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("write %s failed: %v", filePath, err)
	}

	return nil
}
