package configlib

import (
	"bytes"
	"fmt"
	"strings"
)

// convert key=value file to map
func ConvertStrToMap(str string) map[string]string {
	m := map[string]string{}
	strField := strings.Split(str, "\n")
	for _, line := range strField {
		keyValue := strings.Split(line, "=")
		if len(keyValue) != 2 {
			continue
		}
		m[strings.TrimSpace(keyValue[0])] = strings.TrimSpace(keyValue[1])
	}

	return m
}

// convert map[string]string to key=value string
// Input:
//
//	{
//		"key1": "val1",
//		"key2": "val2",
//	}
//
// Output:
//
//	key1=val1
//	key2=val2
func ConvertMapToStr(m map[string]string) string {
	b := new(bytes.Buffer)
	for k, v := range m {
		fmt.Fprintf(b, "%s=%s\n", k, v)
	}
	return b.String()
}
