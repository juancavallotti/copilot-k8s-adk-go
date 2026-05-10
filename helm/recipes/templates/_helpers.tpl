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

{{/*
  RECIPES_API_BASE for the web container.
  With Ingress: must be set in values (API is not on this Ingress) — see web-deployment validation.
  Without Ingress: empty → in-cluster backend URL for SSR + loader-injected client base.
*/}}
{{- define "recipes.webRecipesApiBase" -}}
{{- $u := .Values.web.recipesApiBase | default "" | trim -}}
{{- if ne $u "" -}}
{{- $u -}}
{{- else if not .Values.ingress.enabled -}}
{{- printf "http://%s-backend:4000" (include "recipes.fullname" .) -}}
{{- end -}}
{{- end }}
