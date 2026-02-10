package prometheus

import (
	"errors"
	"fmt"

	"component-manage/internal/global"
	"component-manage/pkg/certutil"
	"component-manage/pkg/models/types"
)

const commonName = "proton-prometheus-observability"

type etcdCertsForPrometheus struct {
	ProtonEtcdCert etcdCertForPrometheus
	K8sEtcdCert    etcdCertForPrometheus
}

type etcdCertForPrometheus struct {
	SecretName       string
	CaCertFieldName  string
	CertFieldName    string
	CertKeyFieldName string
}

func prepareEtcdCertForPrometheus(name string, param *types.PrometheusComponentParams) (etcdCertsForPrometheus, error) {
	certs := etcdCertsForPrometheus{
		ProtonEtcdCert: etcdCertForPrometheus{
			SecretName:       fmt.Sprintf("%s-for-%s", param.CAInfo.ProtonEtcd.CertSecretName, name),
			CaCertFieldName:  "ca-protonetcd.crt",
			CertFieldName:    "prometheus-metrics-protonetcd.crt",
			CertKeyFieldName: "prometheus-metrics-protonetcd.key",
		},
		K8sEtcdCert: etcdCertForPrometheus{
			SecretName:       fmt.Sprintf("%s-for-%s", param.CAInfo.K8sEtcd.SecretName, name),
			CaCertFieldName:  "ca-k8setcd.crt",
			CertFieldName:    "prometheus-metrics-k8setcd.crt",
			CertKeyFieldName: "prometheus-metrics-k8setcd.key",
		},
	}

	if err := prepareK8sEtcdCertForPrometheus(&certs.K8sEtcdCert, param); err != nil {
		return certs, err
	}
	if err := prepareProtonEtcdCertForPrometheus(&certs.ProtonEtcdCert, param); err != nil {
		return certs, err
	}

	return certs, nil
}

func prepareProtonEtcdCertForPrometheus(cert *etcdCertForPrometheus, param *types.PrometheusComponentParams) error {
	existSecret, err := global.K8sCli.SecretExist(cert.SecretName, param.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get secret %s: %w", cert.SecretName, err)
	}
	if existSecret {
		global.Logger.Debugf("skip prepare proton etcd cert")
		return nil
	}

	if exist, _ := global.K8sCli.SecretExist(param.CAInfo.ProtonEtcd.KeySecretName, param.CAInfo.ProtonEtcd.Namespace); !exist {
		return fmt.Errorf("failed to get ca secret %s", param.CAInfo.ProtonEtcd.KeySecretName)
	}
	if exist, _ := global.K8sCli.SecretExist(param.CAInfo.ProtonEtcd.CertSecretName, param.CAInfo.ProtonEtcd.Namespace); !exist {
		return fmt.Errorf("failed to get ca secret %s", param.CAInfo.ProtonEtcd.CertSecretName)
	}
	caKey, err1 := global.K8sCli.SecretGet(param.CAInfo.ProtonEtcd.KeySecretName, param.CAInfo.ProtonEtcd.Namespace)
	caCert, err2 := global.K8sCli.SecretGet(param.CAInfo.ProtonEtcd.CertSecretName, param.CAInfo.ProtonEtcd.Namespace)
	if err1 != nil || err2 != nil {
		return fmt.Errorf(
			"failed to get ca secret %s/%s: %w",
			param.CAInfo.ProtonEtcd.KeySecretName, param.CAInfo.ProtonEtcd.CertSecretName,
			errors.Join(err1, err2),
		)
	}

	caCertData := caCert[param.CAInfo.ProtonEtcd.CertSecretKey]
	caKeyData := caKey[param.CAInfo.ProtonEtcd.KeySecretKey]

	if caCertData == nil || caKeyData == nil {
		return fmt.Errorf("failed to get ca cert or key from secret %s/%s", param.CAInfo.ProtonEtcd.CertSecretName,
			param.CAInfo.ProtonEtcd.KeySecretName,
		)
	}

	_ca, _cert, _key, err := certutil.GenerateSomeCert(commonName, caCertData, caKeyData)
	if err != nil {
		return fmt.Errorf("failed to generate cert for promethues: %w", err)
	}

	return global.K8sCli.SecretSet(cert.SecretName, param.Namespace, map[string][]byte{
		cert.CaCertFieldName:  _ca,
		cert.CertFieldName:    _cert,
		cert.CertKeyFieldName: _key,
	})
}

func prepareK8sEtcdCertForPrometheus(cert *etcdCertForPrometheus, param *types.PrometheusComponentParams) error {
	existSecret, err := global.K8sCli.SecretExist(cert.SecretName, param.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get secret %s: %w", cert.SecretName, err)
	}
	if existSecret {
		global.Logger.Debugf("skip prepare k8s etcd cert")
		return nil
	}

	if exist, _ := global.K8sCli.SecretExist(param.CAInfo.K8sEtcd.SecretName, param.CAInfo.K8sEtcd.Namespace); !exist {
		return fmt.Errorf("failed to get ca secret %s", param.CAInfo.K8sEtcd.SecretName)
	}
	caData, err := global.K8sCli.SecretGet(param.CAInfo.K8sEtcd.SecretName, param.CAInfo.K8sEtcd.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get ca secret %s: %w", param.CAInfo.K8sEtcd.SecretName, err)
	}

	caCertData := caData[param.CAInfo.K8sEtcd.CertKeyname]
	caKeyData := caData[param.CAInfo.K8sEtcd.KeyKeyname]

	if caCertData == nil || caKeyData == nil {
		return fmt.Errorf("failed to get ca cert or key from secret %s", param.CAInfo.K8sEtcd.SecretName)
	}

	_ca, _cert, _key, err := certutil.GenerateSomeCert(commonName, caCertData, caKeyData)
	if err != nil {
		return fmt.Errorf("failed to generate cert for promethues: %w", err)
	}

	return global.K8sCli.SecretSet(cert.SecretName, param.Namespace, map[string][]byte{
		cert.CaCertFieldName:  _ca,
		cert.CertFieldName:    _cert,
		cert.CertKeyFieldName: _key,
	})
}
