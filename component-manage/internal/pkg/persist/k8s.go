package persist

import (
	"fmt"
	"sync"

	"component-manage/pkg/models/types"
	"component-manage/pkg/util"

	"component-manage/pkg/k8s"

	"github.com/sirupsen/logrus"
)

type Persist interface {
	//GetPlugin(name string) (map[string]any, error)
	//SetPlugin(name string, data map[string]any) error
	//
	//GetComponent(name string) (map[string]any, error)
	//SetComponent(name string, data map[string]any) error
	DelComponent(name string) error

	SetPluginObject(name string, obj *types.PluginObject) error
	GetPluginObject(name string) (*types.PluginObject, error)

	GetComponentObject(name string) (*types.ComponentObject, error)
	GetAllComponentObject() (map[string]*types.ComponentObject, error)
	SetComponentObject(name string, obj *types.ComponentObject) error
}

type k8sPersist struct {
	client    k8s.Client
	namespace string

	lock            sync.Mutex
	componentSecret string
	pluginSecret    string
	logger          *logrus.Entry
}

func NewK8sPersist(client k8s.Client, namespace, componentSecret, pluginSecret string, logger *logrus.Entry) Persist {
	return &k8sPersist{
		client:          client,
		namespace:       namespace,
		componentSecret: componentSecret,
		pluginSecret:    pluginSecret,
		logger:          logger,
	}
}

func (k *k8sPersist) GetPlugin(name string) (map[string]any, error) {
	return k.getMapFromSecret(pluginSecret, name)
}

func (k *k8sPersist) SetPlugin(name string, data map[string]any) error {
	return k.setMapToSecret(pluginSecret, name, data)
}

func (k *k8sPersist) DelComponent(name string) error {
	return k.delItemFromSecret(componentSecret, name)
}

func (k *k8sPersist) GetComponent(name string) (map[string]any, error) {
	return k.getMapFromSecret(componentSecret, name)
}

func (k *k8sPersist) SetComponent(name string, data map[string]any) error {
	return k.setMapToSecret(componentSecret, name, data)
}

///

func (k *k8sPersist) SetPluginObject(name string, obj *types.PluginObject) error {
	m, err := util.ToMap(obj)
	if err != nil {
		return fmt.Errorf("failed to convert plugin object to map: %w", err)
	}
	return k.SetPlugin(name, m)
}

func (k *k8sPersist) GetPluginObject(name string) (*types.PluginObject, error) {
	m, err := k.GetPlugin(name)
	if err != nil {
		return nil, err
	}
	obj, err := util.FromMap[types.PluginObject](m)
	if err != nil {
		return nil, fmt.Errorf("failed to convert plugin object from map: %w", err)
	}
	return obj, nil
}

func (k *k8sPersist) GetComponentObject(name string) (*types.ComponentObject, error) {
	m, err := k.GetComponent(name)
	if err != nil {
		return nil, err
	}

	obj, err := util.FromMap[types.ComponentObject](m)
	if err != nil {
		return nil, fmt.Errorf("failed to convert component object from map: %w", err)
	}
	return obj, nil
}

func (k *k8sPersist) SetComponentObject(name string, obj *types.ComponentObject) error {
	m, err := util.ToMap(obj)
	if err != nil {
		return fmt.Errorf("failed to convert component object to map: %w", err)
	}
	return k.SetComponent(name, m)
}

func (k *k8sPersist) GetAllComponentObject() (map[string]*types.ComponentObject, error) {
	datas, err := k.getAllFromSecret(componentSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get all component object: %w", err)
	}
	results := make(map[string]*types.ComponentObject)
	for name, value := range datas {
		obj, err := util.FromMap[types.ComponentObject](value)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal component object: %w", err)
		}
		results[name] = obj
	}
	return results, nil
}
