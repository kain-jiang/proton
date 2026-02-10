package utils

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"taskrunner/trait"

	"github.com/ghodss/yaml"
	cv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// NewKubeCliInCluster create a k8s client in cluster
// TODO sink k8s client operation into abstract struct
func NewKubeCliInCluster() (kubernetes.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

// NewKubeCliOutCluster create a k8s client out cluster
func NewKubeCliOutCluster() (kubernetes.Interface, error) {
	cfgPath := path.Join(homedir.HomeDir(), ".kube", "config")
	cfg, err := clientcmd.BuildConfigFromFlags("", cfgPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

func NewKubeHTTPClient() (*http.Client, *trait.Error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		cfgPath := path.Join(homedir.HomeDir(), ".kube", "config")
		cfg, err = clientcmd.BuildConfigFromFlags("", cfgPath)
		if err != nil {
			return nil, &trait.Error{
				Internal: trait.ECK8sUnknow,
				Err:      err,
				Detail:   "init k8s cluster client",
			}
		}
	}
	cli, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECK8sUnknow,
			Err:      err,
			Detail:   "init k8s cluster client",
		}
	}
	return cli, nil
}

// NewKubeclient create a k8s client
func NewKubeclient() (kubernetes.Interface, *trait.Error) {
	kcli, err := NewKubeCliInCluster()
	if err != nil {
		kcli, err = NewKubeCliOutCluster()
		if err != nil {
			return nil, &trait.Error{
				Internal: trait.ECK8sUnknow,
				Err:      err,
				Detail:   "init k8s cluster client",
			}
		}
	}
	return kcli, nil
}

type SecretRW struct {
	Namespace string
	ConfName  string
	Confkey   string
	Kcli      kubernetes.Interface
}

// NewProtonCli auto load k8s config and create a proton client
func NewSecretRW(namespace, confName, confkey string) (pcli *SecretRW, err *trait.Error) {
	if namespace == "" || confName == "" || confkey == "" {
		return nil, &trait.Error{
			Internal: trait.ErrParam,
			Detail: fmt.Sprintf(
				"must not empty. ns: %s, confName: %s, confKey: %s",
				namespace, confName, confkey),
			Err: nil,
		}
	}
	kcli, rerr := NewKubeclient()
	if rerr != nil {
		return nil, &trait.Error{
			Internal: trait.ECK8sUnknow,
			Detail:   "初始化k8s客户端",
			Err:      rerr,
		}
	}

	return &SecretRW{
		Namespace: namespace,
		ConfName:  confName,
		Confkey:   confkey,
		Kcli:      kcli,
	}, nil
}

func (c *SecretRW) SetContent(ctx context.Context, content any) *trait.Error {
	kcli := c.Kcli
	bs, rerr := yaml.Marshal(content)
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Internal: trait.ErrParam,
			Detail:   "content encode error",
		}
	}
	se := &cv1.Secret{}
	se.Name = c.ConfName
	se.Data = map[string][]byte{
		c.Confkey: bs,
	}
	if _, rerr := kcli.CoreV1().Secrets(c.Namespace).Update(ctx, se, v1.UpdateOptions{}); rerr != nil {
		if errors.IsNotFound(rerr) {
			_, rerr = kcli.CoreV1().Secrets(c.Namespace).Create(ctx, se, v1.CreateOptions{})
		}
		if rerr != nil {
			return &trait.Error{
				Err:      rerr,
				Internal: trait.ECK8sUnknow,
				Detail:   "update secret error",
			}
		}
	}
	return nil
}

// GetFullConf for proton-cli-config
func (c *SecretRW) GetFullConf(ctx context.Context, recv any) (err *trait.Error) {
	bs, err := c.getConfBytes(ctx)
	if err != nil {
		return
	}
	if rerr := yaml.Unmarshal(bs, recv); rerr != nil {
		err = &trait.Error{
			Err:      rerr,
			Internal: trait.ErrComponentDecodeError,
			Detail:   fmt.Sprintf("content '%s' yaml decode error", c.Confkey),
		}
		return
	}

	return nil
}

func (c *SecretRW) getConfBytes(ctx context.Context) ([]byte, *trait.Error) {
	kcli := c.Kcli
	sec, err := kcli.CoreV1().Secrets(c.Namespace).Get(ctx, c.ConfName, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, &trait.Error{
				Err:      err,
				Internal: trait.ErrComponentNotFound,
				Detail:   "proton cli conf not found",
			}
		}
		return nil, &trait.Error{
			Err:      err,
			Internal: trait.ECK8sUnknow,
			Detail:   "k8s secret request",
		}
	}
	scfg, ok := sec.Data[c.Confkey]
	if !ok {
		return nil, &trait.Error{
			Err:      fmt.Errorf("the proton conf miss key '%s', please check the conf version and this tool version", c.Confkey),
			Internal: trait.ErrComponentDecodeError,
			Detail:   fmt.Sprintf("proton cli conf missk key '%s'", c.Confkey),
		}
	}
	// bs, err := base64.StdEncoding.DecodeString(string(scfg))
	// if err != nil {
	// 	err = &trait.WrapperInternalError{
	// 		Err:      err,
	// 		Internal: trait.ErrComponentDecodeError,
	// 		Detail:   fmt.Sprintf("'%s' base64 decode error", c.Confkey),
	// 	}
	// 	return nil, err
	// }
	return scfg, nil
}
