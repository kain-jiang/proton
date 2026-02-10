package detectors

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	etcdRootCAPath     = "/etc/kubernetes/pki/etcd/ca.crt"
	etcdClientCertPath = "/etc/kubernetes/pki/etcd/peer.crt"
	etcdClientKeyPath  = "/etc/kubernetes/pki/etcd/peer.key"
)

func getETCDStatus(nodes []corev1.Node) ([]string, map[string]int64, map[string]uint64, error) {
	caCert, err := os.ReadFile(etcdRootCAPath)
	if err != nil {
		return nil, nil, nil, err
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, nil, nil, fmt.Errorf("failed to append CA certs")
	}
	clientCert, err := tls.LoadX509KeyPair(etcdClientCertPath, etcdClientKeyPath)
	if err != nil {
		return nil, nil, nil, err
	}

	offlineMember := []string{}
	memberDBSize := map[string]int64{}
	raftIndex := map[string]uint64{}

	for _, node := range nodes {
		// etcdEndpoints = append(etcdEndpoints, fmt.Sprintf("%s:2379", node.Status.Addresses[0].Address))
		ep := net.JoinHostPort(node.Status.Addresses[0].Address, "2379")
		etcdCli, err := clientv3.New(clientv3.Config{
			Endpoints:          []string{ep},
			MaxCallSendMsgSize: 100 * 1024 * 1024,
			DialTimeout:        3 * time.Second,
			TLS: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{clientCert},
			},
		})
		if err != nil {
			offlineMember = append(offlineMember, ep)
			continue
		}
		defer etcdCli.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		resp, err := etcdCli.Status(ctx, ep)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				offlineMember = append(offlineMember, ep)
				continue
			}
		}
		memberDBSize[ep] = resp.DbSize
		raftIndex[ep] = resp.RaftIndex
	}

	return offlineMember, memberDBSize, raftIndex, nil
}
