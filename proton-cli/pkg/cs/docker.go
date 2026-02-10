package cs

import (
	"encoding/json"
	"fmt"

	"golang.org/x/exp/slices"
)

const InsecureRegistriesKeyName = "insecure-registries"

func addDockerConfigInsecureHost(crhosts []string, port string, b []byte) ([]byte, error) {
	conf := make(map[string]interface{})
	if err := json.Unmarshal(b, &conf); err != nil {
		return nil, fmt.Errorf("failed to Unmarshal docker daemon.json: %w", err)
	}

	switch t := conf[InsecureRegistriesKeyName].(type) {
	case []interface{}:
		var hosts []string
		for _, v := range t {
			hosts = append(hosts, fmt.Sprintf("%v", v))
		}
		for _, crhost := range crhosts {
			registry := fmt.Sprintf("%s:%s", crhost, port)
			if !slices.Contains(hosts, registry) {
				t = append(t, registry)
				conf[InsecureRegistriesKeyName] = t
			}
		}
	default:
		return nil, fmt.Errorf("failed to detect type %T", t)
	}

	if j, err := json.MarshalIndent(&conf, "", "    "); err != nil {
		return nil, fmt.Errorf("failed to MarshalIndent json content: %w", err)
	} else {
		return j, nil
	}
}
