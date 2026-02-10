{{/* Generate policy-engine names */}}
{{- define "policy-engine.name" }}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/* Generate policy-engine image */}}
{{- define "policy-engine.engine-image" }}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.engine.repository .Values.image.engine.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.engine.repository .Values.image.engine.tag -}}
{{- end -}}
{{- end -}}

{{- define "policy-engine.etcd-name" }}
{{- printf "%s-%s-etcd" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/* Generate policy-engine etcd-image */}}
{{- define "policy-engine.etcd-image" }}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.etcd.repository .Values.image.etcd.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.etcd.repository .Values.image.etcd.tag -}}
{{- end -}}
{{- end -}}

{{- define "policy-engine.exporter-name" }}
{{- printf "%s-%s-exporter" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/* Generate policy-engine exporter-image */}}
{{- define "policy-engine.exporter-image" }}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.exporter.repository .Values.image.exporter.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.exporter.repository .Values.image.exporter.tag -}}
{{- end -}}
{{- end -}}

{{/* Generate policy-engine upgrade-job-image */}}
{{- define "policy-engine.post-upgrade-job-image" }}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.postUpgradeJob.repository .Values.image.postUpgradeJob.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.postUpgradeJob.repository .Values.image.postUpgradeJob.tag -}}
{{- end -}}
{{- end -}}

{{/*
Renders a value that contains template.
Usage:
{{ include "policy-engine.tplValue" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "policy-engine.tplValue" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}
