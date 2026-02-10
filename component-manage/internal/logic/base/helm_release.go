package base

import (
	"fmt"
	"strings"
)

/**
// Expand the name of the chart
{{- define "zookeeper.name" -}}
{{- if contains "zookeeper" .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name "zookeeper" | trunc 63 | trimSuffix "-" }}
{{- end -}}
{{- end -}}
*/

func TemplateName(releaseName, keyword string) string {
	result := releaseName
	if !strings.Contains(releaseName, keyword) {
		result = fmt.Sprintf("%s-%s", releaseName, keyword)
	}
	if len(result) > 63 {
		result = result[:63]
	}
	return strings.TrimSuffix(result, "-")
}
