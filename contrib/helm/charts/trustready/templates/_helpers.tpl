{{/*
Expand the name of the chart.
*/}}
{{- define "trustready.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "trustready.fullname" -}}
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

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "trustready.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "trustready.labels" -}}
helm.sh/chart: {{ include "trustready.chart" . }}
{{ include "trustready.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "trustready.selectorLabels" -}}
app.kubernetes.io/name: {{ include "trustready.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "trustready.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "trustready.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
PostgreSQL hostname
*/}}
{{- define "trustready.postgresql.host" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql-demo" .Release.Name }}
{{- else }}
{{- .Values.postgresql.host | required "postgresql.host is required when postgresql.enabled=false" }}
{{- end }}
{{- end }}

{{/*
PostgreSQL port
*/}}
{{- define "trustready.postgresql.port" -}}
{{- if .Values.postgresql.enabled }}
{{- 5432 }}
{{- else }}
{{- .Values.postgresql.port | default 5432 }}
{{- end }}
{{- end }}

{{/*
PostgreSQL database name
*/}}
{{- define "trustready.postgresql.database" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.database | default "trustreadyd" }}
{{- else }}
{{- .Values.postgresql.database | default "trustreadyd" }}
{{- end }}
{{- end }}

{{/*
PostgreSQL username
*/}}
{{- define "trustready.postgresql.username" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.postgresUser | default "trustreadyd" }}
{{- else }}
{{- .Values.postgresql.username | default "trustreadyd" }}
{{- end }}
{{- end }}

{{/*
PostgreSQL password (from subchart or external config)
*/}}
{{- define "trustready.postgresql.password" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.postgresPassword | required "postgresql.auth.postgresPassword is required when postgresql.enabled=true" }}
{{- else }}
{{- .Values.postgresql.password | required "postgresql.password is required when postgresql.enabled=false" }}
{{- end }}
{{- end }}

{{/*
S3 endpoint
*/}}
{{- define "trustready.s3.endpoint" -}}
{{- if .Values.seaweedfs.enabled }}
{{- printf "http://%s-seaweedfs:8333" (include "trustready.fullname" .) }}
{{- else }}
{{- .Values.s3.endpoint }}
{{- end }}
{{- end }}

{{/*
S3 access key
*/}}
{{- define "trustready.s3.accessKeyId" -}}
{{- if .Values.seaweedfs.enabled }}
{{- .Values.seaweedfs.auth.accessKey | required "seaweedfs.auth.accessKey is required when seaweedfs.enabled=true" }}
{{- else }}
{{- .Values.s3.accessKeyId | required "s3.accessKeyId is required when seaweedfs.enabled=false" }}
{{- end }}
{{- end }}

{{/*
S3 secret key
*/}}
{{- define "trustready.s3.secretAccessKey" -}}
{{- if .Values.seaweedfs.enabled }}
{{- .Values.seaweedfs.auth.secretKey | required "seaweedfs.auth.secretKey is required when seaweedfs.enabled=true" }}
{{- else }}
{{- .Values.s3.secretAccessKey | required "s3.secretAccessKey is required when seaweedfs.enabled=false" }}
{{- end }}
{{- end }}

{{/*
Chrome DevTools Protocol address
*/}}
{{- define "trustready.chrome.addr" -}}
{{- if .Values.chrome.enabled }}
{{- printf "%s-chrome:9222" (include "trustready.fullname" .) }}
{{- else }}
{{- .Values.chrome.external.addr | required "chrome.external.addr is required when chrome.enabled=false" }}
{{- end }}
{{- end }}

{{/*
Validate a required 32-byte base64 secret.
*/}}
{{- define "trustready.requireBase64Key32" -}}
{{- $name := .name -}}
{{- $value := .value | required (printf "%s is required" $name) -}}
{{- if not (regexMatch "^[A-Za-z0-9+/]{43}=$" $value) -}}
{{- fail (printf "%s must be a base64-encoded 32-byte secret (example: openssl rand -base64 32)" $name) -}}
{{- end -}}
{{- $decoded := b64dec $value -}}
{{- if ne (len $decoded) 32 -}}
{{- fail (printf "%s must decode to exactly 32 bytes" $name) -}}
{{- end -}}
{{- $value -}}
{{- end }}

{{/*
Validate a required PEM private key.
*/}}
{{- define "trustready.requirePEMPrivateKey" -}}
{{- $name := .name -}}
{{- $value := .value | required (printf "%s is required" $name) -}}
{{- if not (contains "-----BEGIN" $value) -}}
{{- fail (printf "%s must be a PEM-encoded private key" $name) -}}
{{- end -}}
{{- if not (contains "PRIVATE KEY-----" $value) -}}
{{- fail (printf "%s must be a PEM-encoded private key" $name) -}}
{{- end -}}
{{- $value -}}
{{- end }}

{{/*
Validate the trust center TLS mode.
*/}}
{{- define "trustready.validateTrustCenterTLSMode" -}}
{{- $mode := .Values.trustready.trustCenter.tlsMode -}}
{{- if not (has $mode (list "direct" "external")) -}}
{{- fail (printf "trustready.trustCenter.tlsMode must be direct or external, got %q" $mode) -}}
{{- end -}}
{{- end }}
