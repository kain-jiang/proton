package persist

import (
	"fmt"

	"component-manage/pkg/util"
)

type secretType string

const (
	componentSecret = "component"
	pluginSecret    = "plugin"
)

func (k *k8sPersist) getMapFromSecret(what secretType, name string) (map[string]any, error) {
	secretName := ""
	switch what {
	case componentSecret:
		secretName = k.componentSecret
	case pluginSecret:
		secretName = k.pluginSecret
	default:
		return nil, fmt.Errorf("unknown secret type: %s", what)
	}

	k.lock.Lock()
	defer k.lock.Unlock()

	secretData, err := k.client.SecretGet(secretName, k.namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	yamlData, ok := secretData[name]
	if !ok {
		return nil, nil
	}

	rel, err := util.FromYamlBytes[map[string]any](yamlData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	return rel, nil
}

func (k *k8sPersist) setMapToSecret(what secretType, name string, data map[string]any) error {
	secretName := ""
	switch what {
	case componentSecret:
		secretName = k.componentSecret
	case pluginSecret:
		secretName = k.pluginSecret
	default:
		return fmt.Errorf("unknown secret type: %s", what)
	}

	k.lock.Lock()
	defer k.lock.Unlock()

	secretData, err := k.client.SecretGet(secretName, k.namespace)
	if err != nil {
		return fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	yamlData, ok := secretData[name]
	if ok {
		m, _ := util.FromYamlBytes[map[string]any](yamlData)
		k.logger.WithField("data", m).Infof("%s in type %s exist", name, what)
	}

	yamBytes, err := util.ToYamlBytes(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	secretData[name] = yamBytes

	err = k.client.SecretSet(secretName, k.namespace, secretData)
	if err != nil {
		return fmt.Errorf("failed to set secret %s: %w", secretName, err)
	}

	return nil
}

func (k *k8sPersist) delItemFromSecret(what secretType, name string) error {
	secretName := ""
	switch what {
	case componentSecret:
		secretName = k.componentSecret
	case pluginSecret:
		secretName = k.pluginSecret
	default:
		return fmt.Errorf("unknown secret type: %s", what)
	}

	k.lock.Lock()
	defer k.lock.Unlock()

	secretData, err := k.client.SecretGet(secretName, k.namespace)
	if err != nil {
		return fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	yamlData, ok := secretData[name]
	if ok {
		m, _ := util.FromYamlBytes[map[string]any](yamlData)
		k.logger.WithField("data", m).Infof("%s in type %s exist", name, what)
	}

	delete(secretData, name)

	err = k.client.SecretSet(secretName, k.namespace, secretData)
	if err != nil {
		return fmt.Errorf("failed to set secret %s: %w", secretName, err)
	}
	return nil
}

func (k *k8sPersist) getAllFromSecret(what secretType) (map[string]map[string]any, error) {
	secretName := ""
	switch what {
	case componentSecret:
		secretName = k.componentSecret
	case pluginSecret:
		secretName = k.pluginSecret
	default:
		return nil, fmt.Errorf("unknown secret type: %s", what)
	}

	k.lock.Lock()
	defer k.lock.Unlock()

	secretData, err := k.client.SecretGet(secretName, k.namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	result := make(map[string]map[string]any)
	for name, valBytes := range secretData {
		valMap, err := util.FromYamlBytes[map[string]any](valBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse secret %s.%s: %w", secretName, name, err)
		}
		result[name] = valMap
	}
	return result, nil
}
