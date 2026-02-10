package task

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"taskrunner/pkg/cluster"
	"taskrunner/pkg/component/resources"
	"taskrunner/trait"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HoleTask resource task
type HoleTask struct {
	System *cluster.SystemContext
	Base
}

// Install  impl task interface
// TODO
func (t *HoleTask) Install(ctx context.Context) *trait.Error {
	// TODO upgrade and install
	return nil
}

// Uninstall impl task interface
func (t *HoleTask) Uninstall(ctx context.Context) *trait.Error {
	return nil
}

// ProtonResourceTask do nothing, read config from proton conf
type ProtonResourceTask HoleTask

func newProtonResourceTask(ins *trait.ComponentInstance, s *cluster.SystemContext) *ProtonResourceTask {
	return &ProtonResourceTask{
		System: s,
		Base: Base{
			ComponentInsData: ins,
		},
	}
}

func (t *ProtonResourceTask) Install(ctx context.Context) *trait.Error {
	cins := t.Base.ComponentIns()
	timeout := 600
	if cins.Timeout != 0 {
		timeout = cins.Timeout
	}
	ctx0, cancel := trait.WithTimeoutCauseContext(ctx, time.Second*time.Duration(timeout), &trait.Error{
		Internal: trait.ECTimeout,
		Err:      fmt.Errorf("watch tcp connection timeout or cancel"),
		Detail:   protonResourceType,
	})
	interval := 3 * time.Second
	defer cancel()
	switch cins.Component.Type {
	case resources.RDSType:
		if source, ok := cins.Attribute["source_type"]; !ok || source != "internal" {
			return nil
		}
		host, ok := cins.Attribute["host"]
		if !ok {
			return &trait.Error{
				Internal: trait.ErrParam,
				Err:      fmt.Errorf("rds host field not found"),
			}
		}
		addr, ok := host.(string)
		if !ok {
			return &trait.Error{
				Internal: trait.ErrParam,
				Err:      fmt.Errorf("rds host field not string"),
				Detail:   host,
			}
		}
		port, ok := cins.Attribute["port"]
		if !ok {
			return &trait.Error{
				Internal: trait.ErrParam,
				Err:      fmt.Errorf("rds port field not found"),
			}
		}
		portNum, ok := port.(float64)
		if !ok {
			return &trait.Error{
				Internal: trait.ErrParam,
				Err:      fmt.Errorf("rds port field not number"),
				Detail:   port,
			}
		}

		mgmtHost := cins.Attribute["mgmt_host"].(string)
		mgmtPort := cins.Attribute["mgmt_port"].(int)

		if err := tcpWatch(ctx0, fmt.Sprintf("%s:%d", mgmtHost, mgmtPort), interval, t.Log.Tracef); err != nil {
			return err
		}

		return tcpWatch(ctx0, fmt.Sprintf("%s:%.0f", addr, portNum), interval, t.Log.Tracef)
	case resources.MongodbType:
		if source, ok := cins.Attribute["source_type"]; !ok || source != "internal" {
			return nil
		}
		host, ok := cins.Attribute["host"]
		if !ok {
			return &trait.Error{
				Internal: trait.ErrParam,
				Err:      fmt.Errorf("mongodb host field not found"),
			}
		}
		addr, ok := host.(string)
		if !ok {
			return &trait.Error{
				Internal: trait.ErrParam,
				Err:      fmt.Errorf("mongodb host field not string"),
				Detail:   host,
			}
		}
		hosts := strings.Split(addr, ",")
		if len(hosts) < 1 {
			return &trait.Error{
				Internal: trait.ErrParam,
				Err:      fmt.Errorf("mongodb host field is unnormal format"),
				Detail:   host,
			}
		}
		addr = hosts[0]
		port, ok := cins.Attribute["port"]
		if !ok {
			return &trait.Error{
				Internal: trait.ErrParam,
				Err:      fmt.Errorf("mongodb port field not found"),
			}
		}
		portNum, ok := port.(float64)
		if !ok {
			return &trait.Error{
				Internal: trait.ErrParam,
				Err:      fmt.Errorf("mongodb port field not number"),
				Detail:   port,
			}
		}
		mgmtHost := cins.Attribute["mgmt_host"].(string)
		mgmtPort := cins.Attribute["mgmt_port"].(int)

		if err := tcpWatch(ctx0, fmt.Sprintf("%s:%d", mgmtHost, mgmtPort), interval, t.Log.Tracef); err != nil {
			return err
		}

		return tcpWatch(ctx0, fmt.Sprintf("%s:%.0f", addr, portNum), interval, t.Log.Tracef)
	case resources.EtcdType:
		return t.copyProtonEtcdSecret(ctx, cins)
	}
	return nil
}

func (t *ProtonResourceTask) copyProtonEtcdSecret(ctx context.Context, cins *trait.ComponentInstance) *trait.Error {
	if source, ok := cins.Attribute["source_type"]; !ok || source != "internal" {
		return nil
	}
	ns, ok := cins.Attribute["namespace"]
	if !ok {
		return nil
	}
	if ns == t.System.NameSpace {
		return nil
	}
	secretName := cins.Attribute["secret"].(string)

	kcli := t.System.Kcli
	se, err := kcli.CoreV1().Secrets(ns.(string)).Get(ctx, secretName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return &trait.Error{
			Internal: trait.ErrComponentNotFound,
			Detail:   "无法获取自建etcd的secret配置",
			Err:      err,
		}
	}
	seCopy := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: se.Data,
	}
	if _, err := kcli.CoreV1().Secrets(t.System.NameSpace).Update(ctx, seCopy, metav1.UpdateOptions{}); err != nil {
		if errors.IsNotFound(err) {
			if _, err := kcli.CoreV1().Secrets(t.System.NameSpace).Create(ctx, seCopy, metav1.CreateOptions{}); err != nil {
				return &trait.Error{
					Internal: trait.ECK8sUnknow,
					Detail:   fmt.Sprintf("拷贝secret到%s命名空间失败", t.System.NameSpace),
					Err:      err,
				}
			}
		} else {
			return &trait.Error{
				Internal: trait.ECK8sUnknow,
				Detail:   fmt.Sprintf("拷贝secret到%s命名空间失败", t.System.NameSpace),
				Err:      err,
			}
		}
	}
	return nil
}

func tcpWatch(ctx context.Context, addr string, interval time.Duration, log func(string, ...any)) *trait.Error {
	timeout := 1 * time.Second
	for {
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			select {
			case <-ctx.Done():
				rerr := ctx.Err()
				if err, ok := rerr.(*trait.Error); ok {
					return err
				}
				return &trait.Error{
					Internal: trait.ECTimeout,
					Err:      fmt.Errorf("watch tcp connection timeout or cancel"),
					Detail:   addr,
				}
			default:
				log("try connect tcp %s fail, retry later", addr)
				time.Sleep(interval)
			}
			continue
		}
		defer conn.Close()
		return nil
	}
}
