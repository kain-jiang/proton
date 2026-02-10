package etcd

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"component-manage/internal/global"
	"component-manage/internal/logic/base"
	"component-manage/pkg/models/types"
)

// 为适应多etcd副本多etcd secret需求，如果组件是默认名称就返回基本etcd secret名称，否则和组件名称拼接
func GetETCDName4MultiUser(baseName string, componentName string) string {
	if componentName == ETCDDefaultName {
		return baseName
	} else {
		return fmt.Sprintf("%s-%s", componentName, baseName)
	}
}

func generateCert(name string, params *types.ETCDComponentParams) error {
	kc := global.K8sCli

	for _, n := range []string{
		GetETCDName4MultiUser(ETCDSSLSecretKeyNameBase, name),
		GetETCDName4MultiUser(ETCDSSLSecretNameBase, name),
	} {
		if result, err := kc.SecretExist(n, params.Namespace); err == nil && result {
			global.Logger.Debugf("proton etcd secret %q already exists", name)
			return nil
		} else if err != nil {
			return fmt.Errorf("failed to determine if ETCD secret exists: %w", err)
		}
	}

	global.Logger.Infof("Generate etcd(name:%s) certs", name)

	// 生成ca证书
	// 生成私钥
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("unable to generate ca key error: %w", err)
	}

	caTemplate := newCertTemplate()
	caTemplate.IsCA = true
	caTemplate.KeyUsage = x509.KeyUsageCRLSign | x509.KeyUsageCertSign
	caTemplate.Subject.CommonName = "etcd-ca"
	caTemplate.Issuer = pkix.Name{CommonName: "etcd-ca"}
	// 生成自签证书(template=parent)
	caCertDer, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caPrivateKey.PublicKey, caPrivateKey) // DER 格式
	if err != nil {
		return fmt.Errorf("unable to create self sign ca error: %w", err)
	}
	caPrivateBytes, err := x509.MarshalPKCS8PrivateKey(caPrivateKey)
	if err != nil {
		return fmt.Errorf("unable to converts ca key to PKCS #8, ASN.1 DER form error: %w", err)
	}
	// 将私钥转为pem格式
	caKey := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caPrivateBytes})

	// 将证书转为pem格式
	caCrt := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDer})

	// server

	serverPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("unable to generate server key error: %w", err)
	}
	serverTemplate := newCertTemplate()

	etcdEndpointDNSNameBase := base.TemplateName(name, "proton-etcd")
	serverTemplate.IPAddresses = append(serverTemplate.IPAddresses, net.ParseIP("127.0.0.1"))
	serverTemplate.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	serverTemplate.IsCA = false
	serverTemplate.Subject.CommonName = fmt.Sprintf("%s-peer", etcdEndpointDNSNameBase)
	serverTemplate.DNSNames = append(serverTemplate.DNSNames, fmt.Sprintf("%s-*.%s-headless.%s.svc.cluster.local", etcdEndpointDNSNameBase, etcdEndpointDNSNameBase, params.Namespace))
	serverTemplate.DNSNames = append(serverTemplate.DNSNames, fmt.Sprintf("*.%s-headless.%s.svc.cluster.local", etcdEndpointDNSNameBase, params.Namespace))
	serverTemplate.DNSNames = append(serverTemplate.DNSNames, fmt.Sprintf("*.%s-headless.%s.svc.cluster", etcdEndpointDNSNameBase, params.Namespace))
	serverTemplate.DNSNames = append(serverTemplate.DNSNames, fmt.Sprintf("*.%s-headless.%s.svc", etcdEndpointDNSNameBase, params.Namespace))
	serverTemplate.DNSNames = append(serverTemplate.DNSNames, fmt.Sprintf("*.%s-headless.%s", etcdEndpointDNSNameBase, params.Namespace))
	serverTemplate.DNSNames = append(serverTemplate.DNSNames, fmt.Sprintf("*.%s-headless", etcdEndpointDNSNameBase))
	serverTemplate.DNSNames = append(serverTemplate.DNSNames, fmt.Sprintf("%s.%s.svc.cluster.local", etcdEndpointDNSNameBase, params.Namespace))
	serverTemplate.DNSNames = append(serverTemplate.DNSNames, fmt.Sprintf("%s.%s.svc.cluster", etcdEndpointDNSNameBase, params.Namespace))
	serverTemplate.DNSNames = append(serverTemplate.DNSNames, fmt.Sprintf("%s.%s.svc", etcdEndpointDNSNameBase, params.Namespace))
	serverTemplate.DNSNames = append(serverTemplate.DNSNames, fmt.Sprintf("%s.%s", etcdEndpointDNSNameBase, params.Namespace))
	serverTemplate.DNSNames = append(serverTemplate.DNSNames, etcdEndpointDNSNameBase)
	serverTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	// 使用ca签名
	serverCertDer, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverPrivateKey.PublicKey, caPrivateKey) // DER 格式
	if err != nil {
		return fmt.Errorf("unable to create server key error: %w", err)
	}

	serverPrivateBytes, err := x509.MarshalPKCS8PrivateKey(serverPrivateKey)
	if err != nil {
		return fmt.Errorf("unable to converts server key to PKCS #8, ASN.1 DER form error: %w", err)
	}

	// 将私钥转为pem格式
	serverKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: serverPrivateBytes})

	// 将证书转为pem格式
	serverCrt := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDer})

	global.Logger.Infof("Create kubernetes secret of etcd ca key: %s", GetETCDName4MultiUser(ETCDSSLSecretKeyNameBase, name))
	if err := kc.SecretSet(GetETCDName4MultiUser(ETCDSSLSecretKeyNameBase, name), params.Namespace, map[string][]byte{
		"ca.key": caKey,
	}); err != nil {
		return fmt.Errorf("unable to create %s secret in %s namespace error: %w", GetETCDName4MultiUser(ETCDSSLSecretKeyNameBase, name), params.Namespace, err)
	}

	sslSecretData := map[string][]byte{}
	sslSecretData["ca.crt"] = caCrt
	sslSecretData["peer.key"] = serverKey
	sslSecretData["peer.crt"] = serverCrt

	err = kc.SecretSet(GetETCDName4MultiUser(ETCDSSLSecretNameBase, name), params.Namespace, sslSecretData)
	if err != nil {
		return fmt.Errorf("unable to create %s secret in %s namespace error: %w", GetETCDName4MultiUser(ETCDSSLSecretNameBase, name), params.Namespace, err)
	} else {
		global.Logger.Info(fmt.Sprintf("create secret %s success", GetETCDName4MultiUser(ETCDSSLSecretNameBase, name)))
	}
	return nil
}

func newCertTemplate() *x509.Certificate {
	max := new(big.Int).Lsh(big.NewInt(1), 128)   // 把 1 左移 128 位，返回给 big.Int
	serialNumber, _ := rand.Int(rand.Reader, max) // 返回在 [0, max) 区间均匀随机分布的一个随机值

	template := &x509.Certificate{
		SerialNumber:          serialNumber, // SerialNumber 是 CA 颁布的唯一序列号，在此使用一个大随机数来代表它
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true, // 指示证书是不是ca证书
		BasicConstraintsValid: true, // 指示证书是不是ca证书
	}
	return template
}

func GetEtcdCAInfo(etcdObj *types.ComponentETCD) *types.ETCDCAInfo {
	return &types.ETCDCAInfo{
		Namespace:      etcdObj.Params.Namespace,
		CertSecretName: GetETCDName4MultiUser(ETCDSSLSecretNameBase, etcdObj.Name),
		CertSecretKey:  "ca.crt",
		KeySecretName:  GetETCDName4MultiUser(ETCDSSLSecretKeyNameBase, etcdObj.Name),
		KeySecretKey:   "ca.key",
	}
}
