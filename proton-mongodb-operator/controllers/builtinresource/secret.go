package baseresource

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

	mongodbv1 "proton-mongodb-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewMongoTLSSecret(instance *mongodbv1.MongodbOperator) (*corev1.Secret, error) {
	ca, serverCertKey, err := generateMongoTLS(instance)
	if err != nil {
		return nil, err
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongo-tls-secret",
			Namespace: instance.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"ca":              ca,
			"server-cert-key": serverCertKey,
		},
	}, nil
}

func generateMongoTLS(instance *mongodbv1.MongodbOperator) ([]byte, []byte, error) {
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	caSerial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	caTmpl := &x509.Certificate{
		SerialNumber:          caSerial,
		Subject:               pkix.Name{CommonName: "mongo-tls-ca"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.AddDate(10, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})

	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	serverSerial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	svcName := fmt.Sprintf("%s-%s", instance.GetName(), "mongodb")
	serviceFQDN := fmt.Sprintf("%s.%s.svc.cluster.local", svcName, instance.Namespace)
	wildcardFQDN := fmt.Sprintf("*.%s.%s.svc.cluster.local", svcName, instance.Namespace)

	serverTmpl := &x509.Certificate{
		SerialNumber: serverSerial,
		Subject:      pkix.Name{CommonName: serviceFQDN},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames: []string{
			"localhost",
			svcName,
			serviceFQDN,
			wildcardFQDN,
		},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, serverTmpl, caTmpl, &serverKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})
	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverDER})
	serverCertKeyPEM := append(serverKeyPEM, serverCertPEM...)

	return caPEM, serverCertKeyPEM, nil
}
