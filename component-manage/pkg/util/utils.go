package util

import (
	"fmt"
	"net"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
)

func StringElse(str, el string) string {
	if str != "" {
		return str
	}
	return el
}

type numbers interface {
	int | int32 | int64 | float32 | float64
}

func Min[T numbers](a, b T) T {
	if a > b {
		return b
	} else {
		return a
	}
}

func Max[T numbers](a, b T) T {
	if a > b {
		return a
	} else {
		return b
	}
}

func InSlice[T comparable](s T, slc []T) bool {
	for _, item := range slc {
		if s == item {
			return true
		}
	}
	return false
}

func PackageVersion(pkg string) string {
	pkgName := filepath.Base(pkg)
	rel := regexp.MustCompile(`-(\d+(\.\d+){2,3})-`).FindString(pkgName)
	return strings.Trim(rel, "-")
}

func VersionOrdinal(version string) string {
	// ISO/IEC 14651:2011
	const maxByte = 1<<8 - 1
	vo := make([]byte, 0, len(version)+8)
	j := -1
	for i := 0; i < len(version); i++ {
		b := version[i]
		if '0' > b || b > '9' {
			vo = append(vo, b)
			j = -1
			continue
		}
		if j == -1 {
			vo = append(vo, 0x00)
			j = len(vo) - 1
		}
		if vo[j] == 1 && vo[j+1] == '0' {
			vo[j+1] = b
			continue
		}
		if vo[j]+1 > maxByte {
			panic("VersionOrdinal: invalid version")
		}
		vo = append(vo, b)
		vo[j]++
	}
	return string(vo)
}

func IsIPv6(ipAddr string) bool {
	ip := net.ParseIP(ipAddr)
	return ip != nil && strings.Contains(ipAddr, ":")
}

// ErrorString 获取Error的内容
func ErrorString(err error, dft string) string {
	if err == nil {
		return dft
	}
	return err.Error()
}

func Map[T1 any, T2 any](lst []T1, f func(i T1) T2) []T2 {
	r := make([]T2, 0)
	for _, i := range lst {
		r = append(r, f(i))
	}
	return r
}

// ToMap 将结构体转换为 map 需要拥有 yaml tag
func ToMap[T any](res T) (map[string]any, error) {
	var result map[string]any
	yamlBytes, err := yaml.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("marshal struct to yaml failed: %w", err)
	}

	err = yaml.Unmarshal(yamlBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal yaml to map failed: %w", err)
	}
	return result, nil
}

// FromMap 从map中解析出结构体 需要拥有 yaml tag
func FromMap[T any](res map[string]any) (*T, error) {
	if res == nil {
		return nil, nil
	}

	var result T
	yamlBytes, err := yaml.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("marshal map to yaml failed: %w", err)
	}

	err = yaml.Unmarshal(yamlBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal yaml to struct failed: %w", err)
	}
	return &result, nil
}

// ToYamlBytes 将结构体转换为 yaml
func ToYamlBytes[T any](res T) ([]byte, error) {
	yamlBytes, err := yaml.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("marshal map to yaml failed: %w", err)
	}
	return yamlBytes, nil
}

// FromYamlBytes 将yaml转换为结构体
func FromYamlBytes[T any](data []byte) (T, error) {
	var result T
	err := yaml.Unmarshal(data, &result)
	if err != nil {
		return result, fmt.Errorf("unmarshal yaml to struct failed: %w", err)
	}
	return result, nil
}

// AnyMapToMapAny 将map[string]T转换为map[string]any
func AnyMapToMapAny[T any](data map[string]T) map[string]any {
	result := make(map[string]any, len(data))
	for k, v := range data {
		result[k] = v
	}
	return result
}
