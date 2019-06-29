{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "approval-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "approval-service.fullname" -}}
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
{{- define "approval-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create approval-service hostname
*/}}
{{- define "approval-service-host" -}}
{{- $namespace := default "" .Values.namespace -}}
{{- $stackname := default "" .Values.stackname -}}
{{ if eq $stackname "" "stag" "sand" "prod" }}
{{- printf "approval-service.%s.%s.%s" .Values.region .Values.environment .Values.site -}}
{{else if and (ne $namespace "") (ne $namespace "default")}}
{{- printf "approval-service-%s.%s.%s.%s" $namespace .Values.region .Values.environment .Values.site -}}
{{else}}
{{- printf "approval-service-%s.%s.%s.%s" $stackname .Values.region .Values.environment .Values.site -}}
{{ end }}
{{- end -}}



{{/*
Create approval-service backend
*/}}
{{- define "approval-service-backend" -}}
{{- $namespace := default "" .Values.namespace -}}
{{- $stackname := default "" .Values.stackname -}}
{{ if eq $stackname "" "stag" "sand" "prod" }}
{{- printf "approval-service" -}}
{{else if ne $namespace "" }}
{{- printf "approval-service-%s" $namespace -}}
{{else}}
{{- printf "approval-service" -}}
{{ end }}
{{- end -}}
