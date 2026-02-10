{{/* vim: set filetype=mustache: */}}
{{/* Expand the name of the chart. */}}
{{- define "kafka.name" -}}
{{- if contains "kafka" .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name "kafka" | trunc 63 | trimSuffix "-" }}
{{- end -}}
{{- end -}}


{{/* Return the proper kafka image name */}}
{{- define "kafka.image" -}}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.kafka.repository .Values.image.kafka.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.kafka.repository .Values.image.kafka.tag -}}
{{- end -}}
{{- end -}}


{{/* Return the proper kafka data dir */}}
{{- define "kafka.data_dir" -}}
{{- print "/opt/kafka/work_path" -}}
{{- end -}}


{{/* Generate default_replication_factor */}}
{{- define "kafka.default_replication_factor" }}
{{- if .Values.config.kafkaENV.KAFKA_DEFAULT_REPLICATION_FACTOR }}
{{- .Values.config.kafkaENV.KAFKA_DEFAULT_REPLICATION_FACTOR -}}
{{- else if gt (int .Values.replicaCount) 4 }}
{{- print "3" -}}
{{- else if gt (int .Values.replicaCount) 2 }}
{{- print "2" -}}
{{- else -}}
{{- print "1" -}}
{{- end -}}
{{- end -}}


{{/* Generate offsets_topic_replication_factor */}}
{{- define "kafka.offsets_topic_replication_factor" }}
{{- if .Values.config.kafkaENV.KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR }}
{{- .Values.config.kafkaENV.KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR -}}
{{- else if gt (int .Values.replicaCount) 4 }}
{{- print "3" -}}
{{- else if gt (int .Values.replicaCount) 2 }}
{{- print "2" -}}
{{- else -}}
{{- print "1" -}}
{{- end -}}
{{- end -}}


{{/* Generate num_partitions */}}
{{- define "kafka.num_partitions" }}
{{- if .Values.config.kafkaENV.KAFKA_NUM_PARTITIONS }}
{{- .Values.config.kafkaENV.KAFKA_NUM_PARTITIONS -}}
{{- else }}
{{- .Values.replicaCount -}}
{{- end -}}
{{- end -}}


{{/* Return the default kafka ssl secret name */}}
{{- define "kafka.defaultSSLSecretName" -}}
{{- printf "%s-ssl" (include "kafka.name" .) -}}
{{- end -}}


{{/* Return the kafka headless svc name */}}
{{- define "kafka.headless-svc" -}}
{{- printf "%s-headless.%s:%d" (include "kafka.name" .) .Values.namespace (int .Values.service.kafka.port) -}}
{{- end -}}


{{/* Return the kafka internal security protocol */}}
{{- define "kafka.internal-protocol" -}}
{{- if .Values.config.sasl.enabled -}}
{{- print "SASL_PLAINTEXT" -}}
{{- else -}}
{{- print "PLAINTEXT" -}}
{{- end -}}
{{- end -}}


{{/* Return the kafka external security protocol */}}
{{/* Call this function with "include" and boolean parameters "enableSSL" "enableSASL" */}}
{{/* Example: {{ include "kafka.external-protocol" (dict "enableSSL" $value.enableSSL "enableSASL" $.Values.config.sasl.enabled) }} */}}
{{- define "kafka.external-protocol" -}}
{{- if .enableSASL -}}
  {{- if .enableSSL -}}
    {{- print "SASL_SSL" -}}
  {{- else }}
    {{- print "SASL_PLAINTEXT" -}}
  {{- end -}}
{{- else }}
  {{ if .enableSSL -}}
    {{- print "SSL" -}}
  {{- else -}}
    {{- print "PLAINTEXT" -}}
  {{- end -}}
{{- end -}}
{{- end -}}


{{/* Expand the name of the chart.*/}}
{{- define "kafka-exporter.name" -}}
{{- printf "%s-%s" (include "kafka.name" .) "exporter" -}}
{{- end -}}


{{/* Generate exporter image */}}
{{- define "kafka-exporter.image" }}
{{- if .Values.image.registry }}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.exporter.repository .Values.image.exporter.tag -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.exporter.repository .Values.image.exporter.tag -}}
{{- end -}}
{{- end -}}