package etcd

import (
	"path/filepath"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	KubernetesDir                = "/etc/kubernetes"
	DefaultCertificateDir        = "pki"
	EtcdCACertName               = "etcd/ca.crt"
	EtcdCAKeyName                = "etcd/ca.key"
	APIServerEtcdClientCertName  = "apiserver-etcd-client.crt"
	APIServerEtcdClientKeyName   = "apiserver-etcd-client.key"
	KubeletRunDirectory          = "/var/lib/kubelet"
	KubeletConfigurationFileName = "config.yaml"
	KubeletEnvFileName           = "kubeadm-flags.env"
	KubeletCertDiretory          = "/var/lib/kubelet/pki"
)
const EtcdSnapshotFileName = "etcd-snapshot.db"

const EtcdEndpoint = "https://127.0.0.1:2379"

var (
	APIServerEtcdClientCertPath = filepath.Join(KubernetesDir, DefaultCertificateDir, APIServerEtcdClientCertName)
	APIServerEtcdClientKeyPath  = filepath.Join(KubernetesDir, DefaultCertificateDir, APIServerEtcdClientKeyName)
	EtcdCACertPath              = filepath.Join(KubernetesDir, DefaultCertificateDir, EtcdCACertName)
)

func EtcdClientConfig(endpoint, cert, key, ca string) (clientv3.Config, error) {
	info := &transport.TLSInfo{
		CertFile:      cert,
		KeyFile:       key,
		TrustedCAFile: ca,
	}
	tlsCfg, err := info.ClientConfig()
	if err != nil {
		return clientv3.Config{}, err
	}

	return clientv3.Config{
		Endpoints:            []string{endpoint},
		DialTimeout:          50 * time.Second,
		DialKeepAliveTimeout: 50 * time.Second,
		TLS:                  tlsCfg,
	}, nil
}
