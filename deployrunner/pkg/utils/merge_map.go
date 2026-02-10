package utils

import (
	"os"

	"github.com/ghodss/yaml"
)

func mergeMap(a, b map[string]any, ignoreNil bool) map[string]any {
	out := make(map[string]any, len(a))
	for k, v := range a {
		out[k] = v
	}

	for k, v := range b {
		if v, ok := v.(map[string]any); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]any); ok {
					out[k] = mergeMap(bv, v, ignoreNil)
					continue
				}
			}
		}
		if v != nil || !ignoreNil {
			out[k] = v
		}
	}
	return out
}

// MergeMaps merge two maps
func MergeMaps(maps ...map[string]any) map[string]any {
	length := len(maps)
	output := map[string]any{}
	for i := 0; i < length; i++ {
		output = mergeMap(output, maps[i], false)
	}
	return output
}

// MergeMapsIgnoreNil merge maps, but ignore when next one is nil
func MergeMapsIgnoreNil(maps ...map[string]any) map[string]any {
	length := len(maps)
	output := map[string]any{}
	for i := 0; i < length; i++ {
		output = mergeMap(output, maps[i], true)
	}
	return output
}

func ReadYamlFromFile(fpath string, recv any) error {
	bs, err := os.ReadFile(fpath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bs, recv)
}
