package completion

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"strings"
	"time"

	"k8s.io/utils/clock"

	node "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

func CompleteExtVIP(spec *configuration.ECeph, nodeInterface []node.Interface) {
	flagExtVIPExists := false
	for _, n := range nodeInterface {
		ipInterface, err := n.NetworkInterfaces()
		if err != nil {
			logger.NewLogger().WithError(err).Error("cannot get ip information from ip address command")
		}
		for _, ipAddrSet := range ipInterface {
			for _, a := range ipAddrSet.Addresses {
				if strings.Contains(a.String(), spec.Keepalived.External) {
					if flagExtVIPExists {
						logger.NewLogger().Warning("specified External VIP occurred more than once in ip address command output")
					}
					spec.Keepalived.External = a.String()
					flagExtVIPExists = true
				}
			}
		}
	}
	if !flagExtVIPExists {
		logger.NewLogger().Warning("specified External VIP does not contain subnet mask and cannot complete it from ip address command output")
	}
}

// Complete completes ECeph deployment configuration.
func CompletePost(spec *configuration.ECeph, nodes []configuration.Node, nodeInterface []node.Interface) {
	if spec == nil {
		return
	}

	// queue external vip in ip address command for subnet mask if subnet mask does not exist for some reason
	if spec.Keepalived != nil {
		if !strings.Contains(spec.Keepalived.External, "/") &&
			(strings.Contains(spec.Keepalived.External, ".") || strings.Contains(spec.Keepalived.External, ":")) &&
			len(spec.Keepalived.External) > 1 {
			if net.ParseIP(strings.Split(spec.Keepalived.External, "/")[0]) != nil {
				CompleteExtVIP(spec, nodeInterface)
			}
		}
	}

	// The IP used to issue the certificate. Use the node IP if there is only
	// one node unless use the external network virtual IP.
	var ip string
	if len(nodes) > 0 {
		ip = nodes[0].IP()
	} else if spec.Keepalived != nil {
		// TODO: complete() should return error
		t, _, err := net.ParseCIDR(spec.Keepalived.External)
		if err != nil {
			logger.NewLogger().WithError(err).Error("parse spec.keepalived.external fail")
		}
		ip = t.String()
	}

	completeTLS(&spec.TLS, ip)
}

// completeTLS completes ECeph server's TLS.
//
// TODO: generate tls certificate.
func completeTLS(tls *configuration.ECephTLS, ip string) {
	tls.Secret = completeTLSName(tls.Secret)
	if len(tls.CertificateData) == 0 && len(tls.KeyData) == 0 {
		// TODO: complete() should return error
		var err error
		if tls.CertificateData, tls.KeyData, err = generateSelfSignedCertKey(ip); err != nil {
			logger.NewLogger().WithError(err).Error("generate self signed certificate and key fail")
		}
	}
}

func generateSelfSignedCertKey(commonName string) ([]byte, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	now := clk.Now()

	tmpl := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"Eisoo"},
		},
		NotBefore: now,
		NotAfter:  now.Add(time.Hour * 24 * 365 * 10),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	certDERBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, key.Public(), key)
	if err != nil {
		return nil, nil, err
	}

	var keyBuf bytes.Buffer
	if err := pem.Encode(&keyBuf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		return nil, nil, err
	}

	var certBuf bytes.Buffer
	if err := pem.Encode(&certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: certDERBytes}); err != nil {
		return nil, nil, err
	}

	return certBuf.Bytes(), keyBuf.Bytes(), nil
}

// clk is the interface of clock, which is convenient for testing.
var clk clock.Clock = new(clock.RealClock)

// completeTLSName return the secret name as "eceph-%Y-%m-%d-%H-%M-%S" in
// timezone UTC if the given name is empty.
func completeTLSName(name string) string {
	if name != "" {
		return name
	}
	return clk.Now().Format("eceph-2006-01-02-15-04-05")
}
