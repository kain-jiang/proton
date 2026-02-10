package store

import (
	"context"
	"fmt"

	"taskrunner/pkg/component"
	"taskrunner/pkg/component/resources"
	"taskrunner/pkg/store/proton/configuration"
	"taskrunner/pkg/store/proton/deploy"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	cv1 "k8s.io/api/core/v1"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	_ConfKey         = "ClusterConfiguration"
	_DefaultConfName = "proton-cli-config"
)

type CoreConfig = deploy.CoreConfig

type deployConf struct {
	*ProtonClient
	core CoreConfig
}

func NewStore(namespace, confName, confkey string, core CoreConfig, s trait.Store) (*Store, *trait.Error) {
	pcli, err := NewProtonCli(namespace, confName, confkey)
	return &Store{
		Store: s,
		deployConf: &deployConf{
			ProtonClient: pcli,
			core:         core,
		},
	}, err
}

// ProtonClient read proton conf
type ProtonClient struct {
	Namespace string
	ConfName  string
	Confkey   string
	Kcli      kubernetes.Interface
}

// NewProtonCli auto load k8s config and create a proton client
func NewProtonCli(namespace, confName, confkey string) (pcli *ProtonClient, err *trait.Error) {
	if confkey == "" {
		confkey = _ConfKey
	}
	kcli, err := utils.NewKubeclient()
	return &ProtonClient{
		Namespace: namespace,
		ConfName:  confName,
		Confkey:   confkey,
		Kcli:      kcli,
	}, err
}

// SetFullConf set full config for proton-cli-config. The conf file hasn't patch api
func (c *ProtonClient) SetFullConf(ctx context.Context, conf *configuration.ClusterConfig) *trait.Error {
	kcli := c.Kcli
	bs, rerr := yaml.Marshal(conf)
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Internal: trait.ErrParam,
			Detail:   "proton cli config encode error",
		}
	}
	se := &cv1.Secret{}
	se.Name = c.ConfName
	se.Data = map[string][]byte{
		c.Confkey: bs,
	}
	if _, rerr := kcli.CoreV1().Secrets(c.Namespace).Update(ctx, se, v1.UpdateOptions{}); rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Internal: trait.ECK8sUnknow,
			Detail:   "update proton cli config secret error",
		}
	}
	return nil
}

// GetFullConf for proton-cli-config
func (c *ProtonClient) GetFullConf(ctx context.Context) (conf *configuration.ClusterConfig, err *trait.Error) {
	bs, err := c.getConfBytes(ctx)
	if err != nil {
		return
	}
	conf = &configuration.ClusterConfig{}
	if rerr := yaml.Unmarshal(bs, conf); rerr != nil {
		err = &trait.Error{
			Err:      rerr,
			Internal: trait.ErrComponentDecodeError,
			Detail:   fmt.Sprintf("proton full conf '%s' yaml decode error", c.Confkey),
		}
		return
	}

	if conf.ApiVersion != "v1" {
		rerr := fmt.Errorf("proton cli config only support v1, don't support %s", conf.ApiVersion)
		return nil, &trait.Error{
			Err:      rerr,
			Internal: trait.ErrComponentDecodeError,
			Detail:   "parse proton cli conf",
		}
	}
	return conf, nil
}

func (c *ProtonClient) getConfBytes(ctx context.Context) ([]byte, *trait.Error) {
	kcli := c.Kcli
	sec, err := kcli.CoreV1().Secrets(c.Namespace).Get(ctx, c.ConfName, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, &trait.Error{
				Err:      err,
				Internal: trait.ErrComponentNotFound,
				Detail: fmt.Sprintf(
					"proton cli conf not found, namespace %s, secret: %s",
					c.Namespace, c.ConfName),
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

// GetConf get conf
func (c *ProtonClient) GetConf(ctx context.Context) (*ProtonConf, *trait.Error) {
	bs, err := c.getConfBytes(ctx)
	if err != nil {
		return nil, err
	}

	pcfg := &ProtonConf{}
	rerr := yaml.Unmarshal(bs, pcfg)
	if rerr != nil {
		return nil, &trait.Error{
			Err:      rerr,
			Internal: trait.ErrComponentDecodeError,
			Detail:   fmt.Sprintf("proton conf '%s' yaml decode error", c.Confkey),
		}
	}
	if pcfg.Apiversion != "v1" {
		rerr = fmt.Errorf("proton cli config only support v1, don't support %s", pcfg.Apiversion)
		return nil, &trait.Error{
			Err:      rerr,
			Internal: trait.ErrComponentDecodeError,
			Detail:   "parse proton cli conf",
		}
	}

	return pcfg, nil
}

func replaceAttribute(ctx context.Context, cli *deployConf, cins *trait.ComponentInstance, conf *ProtonConf) (*trait.ComponentInstance, *trait.Error) {
	if cins.Component.ComponentDefineType != component.ComponentProtonResourceType {
		return cins, nil
	}
	var err *trait.Error
	if conf == nil {
		conf, err = cli.GetConf(ctx)
		if err != nil {
			return nil, err
		}
	}
	// TODO reflect to avoid code
	var attr, cfg map[string]interface{}
	switch cins.Component.Type {
	case resources.RDSType:
		rds, err0 := conf.GetRDSComponent()
		if err0 != nil {
			return cins, err0
		}
		attr, cfg, err = rds.ToMap()
	case resources.REDISType:
		redis := conf.Resources.Redis
		if redis == nil {
			return cins, &trait.Error{
				Err:      fmt.Errorf("%s not found, the instance not installed", cins.Component.Type),
				Internal: trait.ErrComponentNotFound,
				Detail:   cins.Component.Type,
			}
		}
		attr, err = redis.ToMap()

	case resources.MQType:
		mq := conf.Resources.MQ
		if mq == nil {
			return cins, &trait.Error{
				Err:      fmt.Errorf("%s not found, the instance not installed", cins.Component.Type),
				Internal: trait.ErrComponentNotFound,
				Detail:   cins.Component.Type,
			}
		}
		attr, err = mq.ToMap()
	case resources.OpensearchType:
		es := conf.Resources.Opensearch
		if es == nil {
			return cins, &trait.Error{
				Err:      fmt.Errorf("%s not found, the instance not installed", cins.Component.Type),
				Internal: trait.ErrComponentNotFound,
				Detail:   cins.Component.Type,
			}
		}
		attr, err = es.ToDepMap()
	case resources.MongodbType:
		obj := conf.Resources.Mongodb
		if obj == nil {
			return cins, &trait.Error{
				Err:      fmt.Errorf("%s not found, the instance num less then 1", cins.Component.Type),
				Internal: trait.ErrComponentNotFound,
				Detail:   cins.Component.Type,
			}
		}
		attr, err = obj.ToDepMap()

	case resources.EtcdType:
		// TODO COPY SECRET
		obj := conf.Resources.Etcd
		if obj == nil {
			return cins, &trait.Error{
				Err:      fmt.Errorf("%s not found, the instance num less then 1", cins.Component.Type),
				Internal: trait.ErrComponentNotFound,
				Detail:   cins.Component.Type,
			}
		}
		ns := conf.DeployConf.Namespace
		if ns == "" {
			ns = "resource"
		}
		attr, err = obj.ToDepMap(ns)
	case resources.POAType:
		obj := conf.Resources.POA
		if obj == nil {
			return cins, &trait.Error{
				Err:      fmt.Errorf("%s not found, the instance num less then 1", cins.Component.Type),
				Internal: trait.ErrComponentNotFound,
				Detail:   cins.Component.Type,
			}
		}
		if host, ok := obj["hosts"]; !ok {
			err = &trait.Error{
				Err:      fmt.Errorf("%s attribute error, the hosts attribute no set", cins.Component.Type),
				Internal: trait.ErrComponentDecodeError,
			}
		} else {
			obj["host"] = host
		}
		attr = obj
	case resources.GraphType:
		stype, ok := cins.Config["source_type"]
		// proton nebula
		if ok && stype == "Proton_Nebula" {
			objConf := conf.Nebula
			if conf.Nebula == nil {
				return cins, &trait.Error{
					Err:      fmt.Errorf("%s not found, the instance num less then 1", cins.Component.Type),
					Internal: trait.ErrComponentNotFound,
					Detail:   cins.Component.Type,
				}
			}

			// create readonly user
			userIns, ok0 := cins.Config["readonlyuser"]
			user, ok00 := userIns.(string)
			passIns, ok1 := cins.Config["readonlypassword"]
			pass, ok11 := passIns.(string)
			if !(ok0 && ok00 && ok1 && ok11) {
				return cins, &trait.Error{
					Err:      fmt.Errorf("%s config error, the readonly user and password must set string, current is user: %#v, pass: %#v", cins.Component.Type, user, pass),
					Internal: trait.ErrComponentDecodeError,
				}
			}

			obj := &resources.GraphDB{
				Type:         "nebulaGraph",
				Host:         "nebula-graphd-svc.resource.svc.cluster.local.",
				Port:         9669,
				User:         "root",
				Password:     objConf.Password,
				ReadonlyUser: user,
				ReadonlyPass: pass,
			}
			attr, err = obj.ToDepMap()

		} else {
			attr = cins.Config
		}
	case resources.DeployCoreType:
		attr = cli.core.ToMapValues()

	default:
		return cins, &trait.Error{
			Err:      fmt.Errorf("%s not found", cins.Component.Type),
			Internal: trait.ErrComponentNotFound,
			Detail:   cins.Component.Type,
		}
	}

	if err != nil {
		return cins, err
	}
	cins.Attribute = utils.MergeMaps(cins.Attribute, attr)
	cins.Config = utils.MergeMaps(cins.Config, cfg)
	return cins, err
}

func replaceAttributes(ctx context.Context, cli *deployConf, cs []*trait.ComponentInstance) *trait.Error {
	if len(cs) == 0 {
		return nil
	}
	conf, err := cli.GetConf(ctx)
	if err != nil {
		return err
	}
	for _, cins := range cs {
		_, err = replaceAttribute(ctx, cli, cins, conf)
		if err != nil {
			return err
		}
	}
	return nil
}

func replaceAttributesIgnoreNoInstalled(ctx context.Context, cli *deployConf, cs []*trait.ComponentInstance) *trait.Error {
	if len(cs) == 0 {
		return nil
	}
	conf, err := cli.GetConf(ctx)
	if err != nil {
		return err
	}
	for _, cins := range cs {
		_, err = replaceAttribute(ctx, cli, cins, conf)
		if err != nil && !trait.IsInternalError(err, trait.ErrComponentNotFound) {
			return err
		}
	}
	return nil
}
