package completion

import (
	"context"
	"fmt"

	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

type asInfoT struct {
	Mode       string `yaml:"mode"`
	Devicespec string `yaml:"devicespec.conf"`
}

func (a *asInfoT) DeviceSpec() string {
	cfg, err := ini.Load([]byte(a.Devicespec))
	if err != nil {
		// unreachable
		panic(err)
	}
	return cfg.Section("DeviceSpec").Key("HardwareType").String()
}

func GuessDeployConfig(ctx context.Context, kube *kubernetes.Clientset, namespace string) (*configuration.Deploy, error) {
	// If namespace is not specified, use the default from file
	if namespace == "" {
		namespace = configuration.GetProtonCliConfigNSFromFile()
	}

	// Log the namespace being used for debugging
	fmt.Printf("GuessDeployConfig using namespace: %s\n", namespace)

	secretName := fmt.Sprintf("%s-%s", SecretPrefix, "anyshare")
	secret, err := kube.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get secret failed from ns %s: %w", namespace, err)
	}
	var asInfo asInfoT
	err = yaml.Unmarshal(secret.Data["default.yaml"], &asInfo)
	if err != nil {
		return nil, fmt.Errorf("cannot analysis as info: %w", err)
	}
	return &configuration.Deploy{
		Mode:       asInfo.Mode,
		Devicespec: asInfo.DeviceSpec(),
	}, nil
}
