package certutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"time"
)

func newCertTemplate() *x509.Certificate {
	max := new(big.Int).Lsh(big.NewInt(1), 128)   // 把 1 左移 128 位，返回给 big.Int
	serialNumber, _ := rand.Int(rand.Reader, max) // 返回在 [0, max) 区间均匀随机分布的一个随机值

	template := &x509.Certificate{
		SerialNumber:          serialNumber, // SerialNumber 是 CA 颁布的唯一序列号，在此使用一个大随机数来代表它
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  false, // 指示证书是不是ca证书
		BasicConstraintsValid: false, // 指示证书是不是ca证书
	}
	return template
}

func GenerateSomeCert(commonName string, caCert, caKey []byte) (ca []byte, cert []byte, key []byte, err error) {
	sPK, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}
	serverTemplate := newCertTemplate()
	serverTemplate.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	serverTemplate.IsCA = false
	serverTemplate.Subject.CommonName = commonName
	serverTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
	caCertPEM, _ := pem.Decode(caCert)
	caCertObj, err := x509.ParseCertificate(caCertPEM.Bytes)
	if err != nil {
		return nil, nil, nil, err
	}
	caKeyPEM, _ := pem.Decode(caKey)
	caPKeyObj, err := x509.ParsePKCS8PrivateKey(caKeyPEM.Bytes)
	if err != nil {
		return nil, nil, nil, err
	}
	sCertDer, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCertObj, &sPK.PublicKey, caPKeyObj)
	if err != nil {
		return nil, nil, nil, err
	}
	sPBytes, err := x509.MarshalPKCS8PrivateKey(sPK)
	if err != nil {
		return nil, nil, nil, err
	}
	serverKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: sPBytes})
	serverCrt := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: sCertDer})
	return caCert, serverCrt, serverKey, nil
}
