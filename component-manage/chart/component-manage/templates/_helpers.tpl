{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "component-manage.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "component-manage.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
Common labels
*/}}
{{- define "component-manage.labels" -}}
app: {{ include "component-manage.name" . }}
{{- end -}}


{{/*
Create the name of the service account to use
*/}}
{{- define "component-manage.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "component-manage.name" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}


{{/*
Create namespace to use
*/}}
{{- define "component-manage.namespace" -}}
{{- default .Release.Namespace .Values.namespace -}}
{{- end -}}

