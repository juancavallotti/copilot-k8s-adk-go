{{- define "recipes.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "recipes.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "recipes.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "recipes.labels" -}}
helm.sh/chart: {{ include "recipes.chart" . }}
{{ include "recipes.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "recipes.selectorLabels" -}}
app.kubernetes.io/name: {{ include "recipes.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "recipes.postgres.serviceName" -}}
{{ include "recipes.fullname" . }}-postgres
{{- end }}

{{- define "recipes.backend.serviceName" -}}
{{ include "recipes.fullname" . }}-backend
{{- end }}

{{- define "recipes.agent.serviceName" -}}
{{ include "recipes.fullname" . }}-agent
{{- end }}

{{- define "recipes.agent.geminiApiKeySecretName" -}}
{{- $name := .Values.agent.geminiApiKeySecret.name | default "" | trim -}}
{{- if ne $name "" -}}
{{- $name -}}
{{- else -}}
{{ include "recipes.fullname" . }}-agent
{{- end -}}
{{- end }}

{{/*
  RECIPES_API_BASE for the web container (Node server only — loaders/actions; never sent to the browser).
  Empty → in-cluster backend Service URL. Ingress only exposes the web app; the API stays internal.
  Set web.recipesApiBase only if the web pod must reach the API at a different URL.
*/}}
{{- define "recipes.webRecipesApiBase" -}}
{{- $u := .Values.web.recipesApiBase | default "" | trim -}}
{{- if ne $u "" -}}
{{- $u -}}
{{- else -}}
{{- printf "http://%s-backend:4000" (include "recipes.fullname" .) -}}
{{- end -}}
{{- end }}
