{{- define "sentinel-agent.name" -}}
sentinel-agent
{{- end }}

{{- define "sentinel-agent.namespace" -}}
{{ .Values.namespace.name }}
{{- end }}

{{- define "sentinel-agent.labels" -}}
app.kubernetes.io/name: sentinel-agent
app.kubernetes.io/part-of: sentinel
{{- end }}
