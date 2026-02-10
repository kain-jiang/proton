package kubernetes

import (
	"bytes"
	"text/template"
)

type kubeadmInitConfig struct {
	Version         string
	NodeIP          string
	ETCDDataDir     string
	LoadBalancer    string
	LoadBalancerIP  string
	ImageRepository string
	IPv4PodCIDR     string
	IPv6PodCIDR     string
	IPv4ServiceCIDR string
	IPv6ServiceCIDR string
	TemplateYAML    []byte
	CRISocket       string
}

type kubeadmJoinConfig struct {
	NodeIP       string
	LoadBalancer string
	Token        string
	CertHash     string
	CertKey      string
	TemplateYAML []byte
	CRISocket    string
}

const kubeadmInitYamlTemplate = `kind: InitConfiguration
apiVersion: kubeadm.k8s.io/v1beta3
bootstrapTokens:
- token: 783bde.3f89s0fje9f38fhf
  description: kubeadm bootstrap token
nodeRegistration:
  {{- with .CRISocket }}
  criSocket: {{ . }}
  {{- end }}
  taints: []
  kubeletExtraArgs:
    allowed-unsafe-sysctls: net.core.somaxconn
    node-ip: {{ .NodeIP }}
localAPIEndpoint:
  advertiseAddress: {{ .NodeIP }}
certificateKey: e6a2eb8581237ab72a4f494f30285ec12a9694d750b9785706a83bfcbbbd2204
---
kind: ClusterConfiguration
apiVersion: kubeadm.k8s.io/v1beta3
{{ if .ETCDDataDir }}
etcd:
  local:
    dataDir: {{ .ETCDDataDir }}
    extraArgs:
      election-timeout: "5000"
      heartbeat-interval: "500"
{{ end }}
networking:
{{- if and .IPv6PodCIDR  .IPv4PodCIDR }}
  podSubnet: {{ .IPv4PodCIDR }},{{ .IPv6PodCIDR }}
  serviceSubnet: {{ .IPv4ServiceCIDR }},{{ .IPv6ServiceCIDR }}
{{- else if .IPv6PodCIDR }}
  podSubnet: {{ .IPv6PodCIDR }}
  serviceSubnet: {{ .IPv6ServiceCIDR }}
{{- else if .IPv4PodCIDR }}
  podSubnet: {{ .IPv4PodCIDR }}
  serviceSubnet: {{ .IPv4ServiceCIDR }}
{{- end }}
kubernetesVersion: {{ .Version }}
controlPlaneEndpoint: {{ .LoadBalancer }}
apiServer:
  certSANs:
  - {{ .LoadBalancerIP }}
  - {{ .NodeIP }}
  extraArgs:
    enable-admission-plugins: NodeRestriction,DefaultStorageClass
imageRepository: {{ .ImageRepository }}
---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
evictionHard:
  imagefs.available: 0%
  nodefs.available: 5%
imageGCHighThresholdPercent: 100
imageGCLowThresholdPercent: 99
maxPods: 300
nodeStatusUpdateFrequency: 3s
`

const kubeadmJoinYamlTemplate = `apiVersion: kubeadm.k8s.io/v1beta3
kind: JoinConfiguration
discovery:
  bootstrapToken:
    apiServerEndpoint: {{ .LoadBalancer }}
    token: {{ .Token }}
    caCertHashes:
    - {{ .CertHash }}
{{- if .CertKey }}
controlPlane:
  certificateKey: {{ .CertKey }}
  localAPIEndpoint:
    advertiseAddress: {{ .NodeIP }}
{{- end }}
nodeRegistration:
  criSocket: {{ .CRISocket }}
  kubeletExtraArgs:
    node-ip: {{ .NodeIP }}
`

func (k *kubeadmInitConfig) renderTemplate(t string) error {
	tmpl, err := template.New("").Parse(t)
	if err != nil {
		return err
	}

	var renderTemplate bytes.Buffer
	if err := tmpl.Execute(&renderTemplate, k); err != nil {
		return err
	}
	k.TemplateYAML = renderTemplate.Bytes()
	return nil
}

func (k *kubeadmJoinConfig) renderTemplate(t string) error {
	tmpl, err := template.New("").Parse(t)
	if err != nil {
		return err
	}

	var renderTemplate bytes.Buffer
	if err := tmpl.Execute(&renderTemplate, k); err != nil {
		return err
	}
	k.TemplateYAML = renderTemplate.Bytes()
	return nil
}
