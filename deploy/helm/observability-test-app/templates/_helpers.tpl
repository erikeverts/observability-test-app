{{- define "app.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}

{{- define "app.chaosEnv" -}}
- name: CHAOS_ERROR_ROUTES
  value: {{ .Values.chaos.errorRoutes | quote }}
- name: CHAOS_LATENCY_ROUTES
  value: {{ .Values.chaos.latencyRoutes | quote }}
- name: CHAOS_CPU_LOAD_ENABLED
  value: {{ .Values.chaos.cpuLoad.enabled | quote }}
- name: CHAOS_CPU_LOAD_PERCENT
  value: {{ .Values.chaos.cpuLoad.percent | quote }}
- name: CHAOS_MEM_LOAD_ENABLED
  value: {{ .Values.chaos.memLoad.enabled | quote }}
- name: CHAOS_MEM_LOAD_MB
  value: {{ .Values.chaos.memLoad.mb | quote }}
- name: CHAOS_LOG_VOLUME_ENABLED
  value: {{ .Values.chaos.logVolume.enabled | quote }}
- name: CHAOS_LOG_RATE_PER_SEC
  value: {{ .Values.chaos.logVolume.ratePerSec | quote }}
- name: CHAOS_LOG_PATTERN
  value: {{ .Values.chaos.logVolume.pattern | quote }}
{{- end -}}

{{- define "app.otelEnv" -}}
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: {{ .Values.otel.endpoint | quote }}
- name: OTEL_EXPORTER_OTLP_PROTOCOL
  value: {{ .Values.otel.protocol | quote }}
{{- if .Values.grafana.endpoint }}
- name: GRAFANA_OTLP_ENDPOINT
  value: {{ .Values.grafana.endpoint | quote }}
- name: GRAFANA_API_TOKEN
  value: {{ .Values.grafana.apiToken | quote }}
{{- end }}
{{- end -}}
