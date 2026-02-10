{{/* vim: set filetype=mustache: */}}
{{/* Expand the name of the chart. */}}

{{/* Generate opensearch name */}}
{{- define "opensearch.name" -}}
{{ .Values.config.clusterName }}-{{ .Values.config.nodeGroup }}
{{- end -}}

{{/* Generate opensearch-exporter name */}}
{{- define "opensearch-exporter.name" -}}
{{ .Values.config.clusterName }}-exporter
{{- end -}}

{{/* Generate opensearch image */}}
{{- define "opensearch.image" -}}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.opensearch.repository .Values.image.opensearch.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.opensearch.repository .Values.image.opensearch.tag -}}
{{- end -}}
{{- end -}}

{{/* Generate exporter image */}}
{{- define "opensearch-exporter.image" }}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.exporter.repository .Values.image.exporter.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.exporter.repository .Values.image.exporter.tag -}}
{{- end -}}
{{- end -}}

{{- define "opensearch.masterNodes" -}}
{{- $name := (include "opensearch.name" .) }}
{{- range $i := .Values.replicaCount | int | until }}
{{- if $i }},{{ end }}{{ $name }}-{{ $i }}
{{- end -}}
{{- end -}}

{{- define "opensearch.seedHosts" -}}
{{ .Values.config.clusterName }}-master-headless:{{ .Values.service.transport.port }}
{{- end -}}

{{- define "opensearch.secretName" -}}
{{ .Values.config.clusterName }}-certs
{{- end -}}

{{- define "opensearch.roles" -}}
{{- if (eq .Values.config.nodeGroup "master") }}
      - cluster_manager
{{- end -}}
{{- if .Values.config.roles.data }}
      - data
{{- end -}}
{{- if .Values.config.roles.search }}
      - search
{{- end -}}
{{- if .Values.config.roles.ingest }}
      - ingest
{{- end -}}
{{- if .Values.config.roles.remoteClusterClient }}
      - remote_cluster_client
{{- end -}}
{{- end -}}

{{- define "opensearch.boxType" -}}
{{- if and (eq .Values.config.nodeGroup "master") .Values.config.roles.data }}
{{- printf "hot" -}}
{{- else if or (eq .Values.config.nodeGroup "hot") (eq .Values.config.nodeGroup "warm") }}
{{- printf "%s" .Values.config.nodeGroup -}}
{{- end -}}
{{- end -}}

{{/* Expand the uri of the OpenSearch.*/}}
{{- define "opensearch-exporter.opensearch-uri" -}}
{{- $name := (include "opensearch.name" .) }}
{{- $portStr := toString .Values.service.http.port }}
{{- if .Values.config.enableSSL }}
{{- printf "https://%s:%s" $name $portStr -}}
{{- else }}
{{- printf "http://%s:%s" $name $portStr -}}
{{- end }}
{{- end -}}
