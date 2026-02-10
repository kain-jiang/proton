{{/* Generate proton-mq-nsq names */}}
{{- define "proton-mq-nsq.name" }}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "nsq.exporter-name" }}
{{- printf "%s-%s-exporter" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/* Generate nsq exporter-image */}}
{{- define "nsq.exporter-image" }}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.exporter.repository .Values.image.exporter.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.exporter.repository .Values.image.exporter.tag -}}
{{- end -}}
{{- end -}}