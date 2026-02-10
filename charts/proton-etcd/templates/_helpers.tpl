{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "etcd.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "etcd.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "etcd.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "etcd.labels" -}}
app.kubernetes.io/name: {{ include "etcd.name" . }}
helm.sh/chart: {{ include "etcd.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Labels to use on deploy.spec.selector.matchLabels and svc.spec.selector
*/}}
{{- define "etcd.matchLabels" -}}
app.kubernetes.io/name: {{ include "etcd.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}


{{/*
Return the proper etcd image name
*/}}
{{- define "etcd.image" -}}
{{- $registryName := .Values.image.registry -}}
{{- $repositoryName := .Values.image.repository -}}
{{- $tag := .Values.image.tag | toString -}}
    {{- printf "%s/%s:%s" $registryName $repositoryName $tag -}}
{{- end -}}

{{/*
Return the proper etcd peer protocol
*/}}
{{- define "etcd.peerProtocol" -}}
{{- if .Values.auth.peer.enableAuthentication -}}
{{- print "https" -}}
{{- else -}}
{{- print "http" -}}
{{- end -}}
{{- end -}}

{{/*
Return the proper etcd client protocol
*/}}
{{- define "etcd.clientProtocol" -}}
{{- if .Values.auth.client.enableAuthentication -}}
{{- print "https" -}}
{{- else -}}
{{- print "http" -}}
{{- end -}}
{{- end -}}

{{/*
Return the proper etcd data dir
*/}}
{{- define "etcd.dataDir" -}}
{{- print "/var/lib/etcd/data" -}}
{{- end -}}

{{/*
Return the proper etcdctl authentication options
*/}}
{{- define "etcd.authOptions" -}}
{{- $rbacOption := "--user root:$ETCD_ROOT_PASSWORD" -}}
{{- $certsOption := " --cert $ETCD_CERT_FILE --key $ETCD_KEY_FILE" -}}
{{- $caOption := " --cacert $ETCD_TRUSTED_CA_FILE" -}}
{{- if .Values.auth.rbac.enabled -}}
{{- printf "%s" $rbacOption -}}
{{- end -}}
{{- if .Values.auth.client.enableAuthentication -}}
{{- printf "%s" $certsOption -}}
{{- printf "%s" $caOption -}}
{{- end -}}
{{- end -}}

{{/*
Return the etcd env vars ConfigMap name
*/}}
{{- define "etcd.envVarsCM" -}}
{{- printf "%s" .Values.envVarsConfigMap -}}
{{- end -}}

{{/*
Return the etcd env vars ConfigMap name
*/}}
{{- define "etcd.configFileCM" -}}
{{- printf "%s" .Values.configFileConfigMap -}}
{{- end -}}


{{/*
Return the proper Storage Class
*/}}
{{- define "etcd.storageClass" -}}
  {{- if .Values.persistence.storageClass -}}
      {{- if (eq "-" .Values.persistence.storageClass) -}}
          {{- printf "storageClassName: \"\"" -}}
      {{- else }}
          {{- printf "storageClassName: %s" .Values.persistence.storageClass -}}
      {{- end -}}
  {{- end -}}
{{- end -}}

{{/*
Renders a value that contains template.
Usage:
{{ include "etcd.tplValue" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "etcd.tplValue" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}
