// Copyright (c) 2026 TrustReady <hello@trustready.io>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudaws "github.com/DhruvWork/trustready-grc/pkg/cloud/aws"
	cloudazure "github.com/DhruvWork/trustready-grc/pkg/cloud/azure"
	cloudgcp "github.com/DhruvWork/trustready-grc/pkg/cloud/gcp"
	"github.com/DhruvWork/trustready-grc/pkg/connector"
	"github.com/DhruvWork/trustready-grc/pkg/crypto/keys"
	"github.com/DhruvWork/trustready-grc/pkg/crypto/pem"
	"github.com/DhruvWork/trustready-grc/pkg/trustreadydconfig"
)

// The configuration decodes key material as it is read, so a test that reaches
// Build must hand it a real key. Generating one costs enough that the suite
// shares them; a second signing key lets a test tell two keys apart.
var (
	testSigningKeyPEM = sync.OnceValue(func() string {
		keyPEM, err := GenerateOAuth2SigningKey()
		if err != nil {
			panic(err)
		}

		return keyPEM
	})

	testOtherSigningKeyPEM = sync.OnceValue(func() string {
		keyPEM, err := GenerateOAuth2SigningKey()
		if err != nil {
			panic(err)
		}

		return keyPEM
	})

	testECKeyPEM = sync.OnceValue(func() string {
		key, err := keys.Generate(keys.TypeEC256)
		if err != nil {
			panic(err)
		}

		keyPEM, err := pem.EncodePrivateKey(key)
		if err != nil {
			panic(err)
		}

		return string(keyPEM)
	})
)

func mockEnv(env map[string]string) EnvGetter {
	return func(key string) string {
		return env[key]
	}
}

// mockEnumerator wraps env in os.Environ's "KEY=VALUE" form, so a test can
// give a mockEnv-backed Resolver the enumeration capability NewResolver(nil)
// gives the real process environment for free (see Resolver.SetEnumerator).
func mockEnumerator(env map[string]string) EnvEnumerator {
	return func() []string {
		pairs := make([]string, 0, len(env))
		for k, v := range env {
			pairs = append(pairs, k+"="+v)
		}

		return pairs
	}
}

func requiredEnv() map[string]string {
	return map[string]string{
		"TRUSTREADYD_ENCRYPTION_KEY":            "test-encryption-key-32-bytes-long",
		"TRUSTREADYD_AUTH_COOKIE_SECRET":        "test-cookie-secret-32-bytes-long!",
		"TRUSTREADYD_AUTH_PASSWORD_PEPPER":      "test-password-pepper-32-bytes-lo",
		"TRUSTREADYD_OAUTH2_SERVER_SIGNING_KEY": testSigningKeyPEM(),
	}
}

func TestBuilder_Build_MissingRequiredEnvVars(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantMissing []string
	}{
		{
			name:        "all missing",
			env:         map[string]string{},
			wantMissing: []string{"TRUSTREADYD_ENCRYPTION_KEY", "TRUSTREADYD_AUTH_COOKIE_SECRET", "TRUSTREADYD_AUTH_PASSWORD_PEPPER", "TRUSTREADYD_OAUTH2_SERVER_SIGNING_KEY"},
		},
		{
			name: "missing oauth2 signing key",
			env: map[string]string{
				"TRUSTREADYD_ENCRYPTION_KEY":       "key",
				"TRUSTREADYD_AUTH_COOKIE_SECRET":   "secret",
				"TRUSTREADYD_AUTH_PASSWORD_PEPPER": "pepper",
			},
			wantMissing: []string{"TRUSTREADYD_OAUTH2_SERVER_SIGNING_KEY"},
		},
		{
			name: "missing encryption key",
			env: map[string]string{
				"TRUSTREADYD_AUTH_COOKIE_SECRET":   "secret",
				"TRUSTREADYD_AUTH_PASSWORD_PEPPER": "pepper",
			},
			wantMissing: []string{"TRUSTREADYD_ENCRYPTION_KEY"},
		},
		{
			name: "missing cookie secret",
			env: map[string]string{
				"TRUSTREADYD_ENCRYPTION_KEY":       "key",
				"TRUSTREADYD_AUTH_PASSWORD_PEPPER": "pepper",
			},
			wantMissing: []string{"TRUSTREADYD_AUTH_COOKIE_SECRET"},
		},
		{
			name: "slack connector missing required fields",
			env: map[string]string{
				"TRUSTREADYD_ENCRYPTION_KEY":            "key",
				"TRUSTREADYD_AUTH_COOKIE_SECRET":        "secret",
				"TRUSTREADYD_AUTH_PASSWORD_PEPPER":      "pepper",
				"TRUSTREADYD_CONNECTOR_SLACK_CLIENT_ID": "client-id",
			},
			wantMissing: []string{"TRUSTREADYD_CONNECTOR_SLACK_CLIENT_SECRET", "TRUSTREADYD_CONNECTOR_SLACK_SIGNING_SECRET"},
		},
		{
			name: "enabled slackbot missing required fields",
			env: map[string]string{
				"TRUSTREADYD_ENCRYPTION_KEY":            "key",
				"TRUSTREADYD_AUTH_COOKIE_SECRET":        "secret",
				"TRUSTREADYD_AUTH_PASSWORD_PEPPER":      "pepper",
				"TRUSTREADYD_SLACKBOT_ENABLED":          "true",
				"TRUSTREADYD_OAUTH2_SERVER_SIGNING_KEY": "signing-key",
			},
			wantMissing: []string{
				"TRUSTREADYD_SLACKBOT_SIGNING_SECRET",
				"TRUSTREADYD_SLACKBOT_CLIENT_ID",
				"TRUSTREADYD_SLACKBOT_CLIENT_SECRET",
				"TRUSTREADYD_SLACKBOT_REDIRECT_URI",
				"TRUSTREADYD_OPENAI_API_KEY",
			},
		},
		{
			name: "google workspace connector missing required fields",
			env: map[string]string{
				"TRUSTREADYD_ENCRYPTION_KEY":                       "key",
				"TRUSTREADYD_AUTH_COOKIE_SECRET":                   "secret",
				"TRUSTREADYD_AUTH_PASSWORD_PEPPER":                 "pepper",
				"TRUSTREADYD_CONNECTOR_GOOGLE_WORKSPACE_CLIENT_ID": "client-id",
			},
			wantMissing: []string{"TRUSTREADYD_CONNECTOR_GOOGLE_WORKSPACE_CLIENT_SECRET"},
		},
		{
			name: "microsoft 365 connector missing required fields",
			env: map[string]string{
				"TRUSTREADYD_ENCRYPTION_KEY":                    "key",
				"TRUSTREADYD_AUTH_COOKIE_SECRET":                "secret",
				"TRUSTREADYD_AUTH_PASSWORD_PEPPER":              "pepper",
				"TRUSTREADYD_CONNECTOR_MICROSOFT_365_CLIENT_ID": "client-id",
			},
			wantMissing: []string{"TRUSTREADYD_CONNECTOR_MICROSOFT_365_CLIENT_SECRET"},
		},
		{
			name: "cal.com connector missing required fields",
			env: map[string]string{
				"TRUSTREADYD_ENCRYPTION_KEY":              "key",
				"TRUSTREADYD_AUTH_COOKIE_SECRET":          "secret",
				"TRUSTREADYD_AUTH_PASSWORD_PEPPER":        "pepper",
				"TRUSTREADYD_CONNECTOR_CAL_COM_CLIENT_ID": "client-id",
			},
			wantMissing: []string{"TRUSTREADYD_CONNECTOR_CAL_COM_CLIENT_SECRET"},
		},
		{
			name: "calendly connector missing required fields",
			env: map[string]string{
				"TRUSTREADYD_ENCRYPTION_KEY":               "key",
				"TRUSTREADYD_AUTH_COOKIE_SECRET":           "secret",
				"TRUSTREADYD_AUTH_PASSWORD_PEPPER":         "pepper",
				"TRUSTREADYD_CONNECTOR_CALENDLY_CLIENT_ID": "client-id",
			},
			wantMissing: []string{"TRUSTREADYD_CONNECTOR_CALENDLY_CLIENT_SECRET"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(NewResolver(mockEnv(tt.env)))
			_, err := b.Build()

			require.Error(t, err)

			for _, missing := range tt.wantMissing {
				assert.Contains(t, err.Error(), missing)
			}
		})
	}
}

func TestBuilder_Build_InvalidCookieBannerTCFCmpID(t *testing.T) {
	t.Parallel()

	for _, id := range []string{"1", "4096"} {
		t.Run(
			id,
			func(t *testing.T) {
				t.Parallel()

				env := requiredEnv()
				env["TRUSTREADYD_COOKIE_BANNER_TCF_CMP_ID"] = id

				b := NewBuilder(NewResolver(mockEnv(env)))
				_, err := b.Build()

				require.Error(t, err)
				assert.Contains(t, err.Error(), "TRUSTREADYD_COOKIE_BANNER_TCF_CMP_ID must be between 2 and 4095")
			},
		)
	}
}

func TestBuilder_Build_InvalidCompliancePortalTLSMode(t *testing.T) {
	t.Parallel()

	env := requiredEnv()
	env["TRUSTREADYD_TRUST_CENTER_TLS_MODE"] = "passthrough"

	b := NewBuilder(NewResolver(mockEnv(env)))
	_, err := b.Build()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse TRUSTREADYD_TRUST_CENTER_TLS_MODE")
}

func TestBuilder_Build_Defaults(t *testing.T) {
	b := NewBuilder(NewResolver(mockEnv(requiredEnv())))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	// Unit config
	assert.Equal(t, "localhost:8081", cfg.Unit.Metrics.Addr)
	assert.Equal(t, "localhost:4318", cfg.Unit.Tracing.Addr)
	assert.Equal(t, 512, cfg.Unit.Tracing.MaxBatchSize)
	assert.Equal(t, 5, cfg.Unit.Tracing.BatchTimeout)
	assert.Equal(t, 30, cfg.Unit.Tracing.ExportTimeout)
	assert.Equal(t, 2048, cfg.Unit.Tracing.MaxQueueSize)

	// TrustReadyd base config
	assert.Empty(t, cfg.TrustReadyd.BaseURL)
	assert.Empty(t, cfg.TrustReadyd.ChromeDPAddr)

	// API config
	assert.Empty(t, cfg.TrustReadyd.Api.Addr)
	assert.Nil(t, cfg.TrustReadyd.Api.ProxyProtocol.TrustedProxies)
	assert.Nil(t, cfg.TrustReadyd.Api.Cors.AllowedOrigins)
	assert.Equal(t, 15000, cfg.TrustReadyd.Api.GraphQL.ParserTokenLimit)
	assert.Equal(t, 2000, cfg.TrustReadyd.Api.GraphQL.ComplexityLimit)
	assert.Equal(t, 1000, cfg.TrustReadyd.Api.GraphQL.QueryCacheSize)
	assert.True(t, cfg.TrustReadyd.Api.GraphQL.DisableSuggestion)

	// PG config
	assert.Empty(t, cfg.TrustReadyd.Pg.Addr)
	assert.Empty(t, cfg.TrustReadyd.Pg.Username)
	assert.Empty(t, cfg.TrustReadyd.Pg.Password)
	assert.Empty(t, cfg.TrustReadyd.Pg.Database)
	assert.Equal(t, int32(100), cfg.TrustReadyd.Pg.PoolSize)
	assert.Equal(t, int32(10), cfg.TrustReadyd.Pg.MinPoolSize)
	assert.Equal(t, 1800, cfg.TrustReadyd.Pg.MaxConnIdleTimeSeconds)
	assert.Equal(t, 3600, cfg.TrustReadyd.Pg.MaxConnLifetimeSeconds)
	assert.Equal(t, 300, cfg.TrustReadyd.Pg.MaxConnLifetimeJitterSeconds)
	assert.Equal(t, 60, cfg.TrustReadyd.Pg.HealthCheckPeriodSeconds)
	assert.False(t, cfg.TrustReadyd.Pg.Debug)

	// Auth config
	assert.False(t, cfg.TrustReadyd.Auth.DisableSignup)
	assert.Equal(t, 3600, cfg.TrustReadyd.Auth.InvitationConfirmationTokenValidity)
	assert.Equal(t, 3600, cfg.TrustReadyd.Auth.PasswordResetTokenValidity)
	assert.Equal(t, 900, cfg.TrustReadyd.Auth.MagicLinkTokenValidity)
	assert.Equal(t, 3600, cfg.TrustReadyd.Auth.EmailConfirmationTokenValidity)
	assert.Empty(t, cfg.TrustReadyd.Auth.Cookie.Name)
	assert.Empty(t, cfg.TrustReadyd.Auth.Cookie.Domain)
	assert.Equal(t, 24, cfg.TrustReadyd.Auth.Cookie.Duration)
	assert.True(t, cfg.TrustReadyd.Auth.Cookie.Secure)
	assert.Equal(t, 1000000, cfg.TrustReadyd.Auth.Password.Iterations)

	// SAML config
	assert.Equal(t, 604800, cfg.TrustReadyd.Auth.SAML.SessionDuration)
	assert.Equal(t, 0, cfg.TrustReadyd.Auth.SAML.CleanupIntervalSeconds)
	assert.Equal(t, 60, cfg.TrustReadyd.Auth.SAML.DomainVerificationIntervalSeconds)
	assert.Empty(t, cfg.TrustReadyd.Auth.SAML.DomainVerificationResolverAddr)

	// ITAM config
	assert.Equal(t, 604800, cfg.TrustReadyd.ITAM.DeviceEnrollmentTokenValidity)

	// Trust center config
	assert.Empty(t, cfg.TrustReadyd.CompliancePortal.HTTPAddr)
	assert.Empty(t, cfg.TrustReadyd.CompliancePortal.HTTPSAddr)
	assert.Empty(t, cfg.TrustReadyd.CompliancePortal.BaseDomain)
	assert.Equal(t, trustreadydconfig.CompliancePortalTLSModeDirect, cfg.TrustReadyd.CompliancePortal.TLSMode)
	assert.Nil(t, cfg.TrustReadyd.CompliancePortal.ProxyProtocol.TrustedProxies)

	// AWS config
	assert.Empty(t, cfg.TrustReadyd.AWS.Region)
	assert.Empty(t, cfg.TrustReadyd.AWS.Bucket)
	assert.False(t, cfg.TrustReadyd.AWS.UsePathStyle)

	// Notifications config
	assert.Empty(t, cfg.TrustReadyd.Notifications.Mailer.SenderName)
	assert.Empty(t, cfg.TrustReadyd.Notifications.Mailer.SenderEmail)
	assert.Empty(t, cfg.TrustReadyd.Notifications.Mailer.SMTP.Addr)
	assert.False(t, cfg.TrustReadyd.Notifications.Mailer.SMTP.TLSRequired)
	assert.Empty(t, cfg.TrustReadyd.Notifications.Mailer.SMTP.HelloName)
	assert.Equal(t, 60, cfg.TrustReadyd.Notifications.Mailer.MailerInterval)
	assert.Equal(t, 60, cfg.TrustReadyd.Notifications.Slack.SenderInterval)
	assert.Empty(t, cfg.TrustReadyd.Notifications.Slack.SigningSecret)
	assert.Equal(t, 5, cfg.TrustReadyd.Notifications.Webhook.SenderInterval)
	assert.Equal(t, 15, cfg.TrustReadyd.Notifications.Webhook.RequestTimeout)
	assert.Equal(t, 86400, cfg.TrustReadyd.Notifications.Webhook.CacheTTL)
	assert.Equal(t, 300, cfg.TrustReadyd.Notifications.Webhook.StaleAfter)
	assert.Equal(t, 30, cfg.TrustReadyd.Notifications.Webhook.RetryBase)
	assert.Equal(t, 14400, cfg.TrustReadyd.Notifications.Webhook.RetryMax)
	assert.Equal(t, 5, cfg.TrustReadyd.Notifications.Webhook.MaxConcurrency)
	assert.Equal(t, 300, cfg.TrustReadyd.Notifications.Document.Interval)
	assert.Equal(t, 900, cfg.TrustReadyd.Notifications.Document.DebounceDelay)
	assert.Equal(t, 86400, cfg.TrustReadyd.Notifications.Document.ReminderInterval)

	// Agents tools — Firecrawl empty by default
	assert.Empty(t, cfg.TrustReadyd.Agents.Tools.FirecrawlAPIKey)

	// Agents config — default
	assert.Equal(t, "openai", cfg.TrustReadyd.Agents.Default.Provider)
	assert.Equal(t, "gpt-4o", cfg.TrustReadyd.Agents.Default.ModelName)
	assert.Equal(t, new(0.1), cfg.TrustReadyd.Agents.Default.Temperature)
	assert.Equal(t, new(4096), cfg.TrustReadyd.Agents.Default.MaxTokens)
	assert.Nil(t, cfg.TrustReadyd.Agents.Providers)
	// Agents config — per-agent overrides are empty (inherit from default)
	assert.Empty(t, cfg.TrustReadyd.Agents.TrustReady.Provider)
	assert.Empty(t, cfg.TrustReadyd.Agents.TrustReady.ModelName)
	assert.Nil(t, cfg.TrustReadyd.Agents.TrustReady.Temperature)
	assert.Nil(t, cfg.TrustReadyd.Agents.TrustReady.MaxTokens)
	assert.Empty(t, cfg.TrustReadyd.Agents.EvidenceDescriber.Provider)
	assert.Empty(t, cfg.TrustReadyd.Agents.EvidenceDescriber.ModelName)
	assert.Nil(t, cfg.TrustReadyd.Agents.EvidenceDescriber.Temperature)
	assert.Nil(t, cfg.TrustReadyd.Agents.EvidenceDescriber.MaxTokens)
	assert.Empty(t, cfg.TrustReadyd.Agents.ThirdPartyVetter.Provider)
	assert.Empty(t, cfg.TrustReadyd.Agents.ThirdPartyVetter.ModelName)
	assert.Nil(t, cfg.TrustReadyd.Agents.ThirdPartyVetter.Temperature)
	assert.Nil(t, cfg.TrustReadyd.Agents.ThirdPartyVetter.MaxTokens)
	assert.Empty(t, cfg.TrustReadyd.Agents.TrackerMapping.Provider)
	assert.Empty(t, cfg.TrustReadyd.Agents.TrackerMapping.ModelName)
	assert.Nil(t, cfg.TrustReadyd.Agents.TrackerMapping.Temperature)
	assert.Equal(t, new(4096), cfg.TrustReadyd.Agents.TrackerMapping.MaxTokens)
	assert.Empty(t, cfg.TrustReadyd.Agents.TrackerEnrichment.Provider)
	assert.Empty(t, cfg.TrustReadyd.Agents.TrackerEnrichment.ModelName)
	assert.Nil(t, cfg.TrustReadyd.Agents.TrackerEnrichment.Temperature)
	assert.Equal(t, new(4096), cfg.TrustReadyd.Agents.TrackerEnrichment.MaxTokens)

	// Tracker worker tuning — defaults
	assert.Equal(t, 10, cfg.TrustReadyd.TrackerMappingWorker.Interval)
	assert.Equal(t, 3, cfg.TrustReadyd.TrackerMappingWorker.MaxConcurrency)
	assert.Equal(t, 600, cfg.TrustReadyd.TrackerMappingWorker.StaleAfter)
	assert.Equal(t, 45, cfg.TrustReadyd.TrackerMappingWorker.AgentTimeout)
	assert.Equal(t, 10, cfg.TrustReadyd.TrackerMappingWorker.AgentMaxTurns)
	assert.Equal(t, 10, cfg.TrustReadyd.CommonPatternEnrichmentWorker.Interval)
	assert.Equal(t, 2, cfg.TrustReadyd.CommonPatternEnrichmentWorker.MaxConcurrency)
	assert.Equal(t, 600, cfg.TrustReadyd.CommonPatternEnrichmentWorker.StaleAfter)
	assert.Equal(t, 45, cfg.TrustReadyd.CommonPatternEnrichmentWorker.AgentTimeout)
	assert.Equal(t, 10, cfg.TrustReadyd.CommonPatternEnrichmentWorker.AgentMaxTurns)
	assert.Equal(t, 10, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.Interval)
	assert.Equal(t, 1, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.MaxConcurrency)
	assert.Equal(t, 900, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.StaleAfter)
	assert.Equal(t, 90, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.AgentTimeout)
	assert.Equal(t, 12, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.AgentMaxTurns)
	assert.Equal(t, 0.7, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.ConfidenceThreshold)
	assert.Equal(t, 3, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.MaxAttempts)
	assert.Equal(t, 10, cfg.TrustReadyd.ThirdPartyVetting.Interval)
	assert.Equal(t, 1500, cfg.TrustReadyd.ThirdPartyVetting.StaleAfter)
	assert.Equal(t, 1, cfg.TrustReadyd.ThirdPartyVetting.MaxConcurrency)

	// Custom domains config
	assert.Equal(t, 3600, cfg.TrustReadyd.CustomDomains.RenewalInterval)
	assert.Equal(t, 30, cfg.TrustReadyd.CustomDomains.ProvisionInterval)
	assert.Equal(t, "custom.trustready.io", cfg.TrustReadyd.CustomDomains.CnameTarget)
	assert.Empty(t, cfg.TrustReadyd.CustomDomains.ResolverAddr)
	assert.Empty(t, cfg.TrustReadyd.CustomDomains.ACME.Directory)
	assert.Empty(t, cfg.TrustReadyd.CustomDomains.ACME.Email)
	assert.Empty(t, cfg.TrustReadyd.CustomDomains.ACME.KeyType)

	// SCIM bridge config
	assert.Equal(t, 900, cfg.TrustReadyd.SCIMBridge.SyncInterval)
	assert.Equal(t, 30, cfg.TrustReadyd.SCIMBridge.PollInterval)

	// ESign config
	assert.Empty(t, cfg.TrustReadyd.ESign.TSAURL)

	// Branding
	assert.True(t, cfg.TrustReadyd.Branding)

	// Cookie banner
	assert.Equal(t, 4095, cfg.TrustReadyd.CookieBanner.TCFCMPID)

	// No connectors by default
	assert.Empty(t, cfg.TrustReadyd.Connectors)
}

func TestBuilder_Build_CustomValues(t *testing.T) {
	env := requiredEnv()
	// Unit
	env["TRUSTREADYD_METRICS_ADDR"] = "0.0.0.0:9090"
	env["TRUSTREADYD_TRACING_ADDR"] = "jaeger:4317"
	env["TRUSTREADYD_TRACING_MAX_BATCH_SIZE"] = "1024"
	// TrustReadyd
	env["TRUSTREADYD_BASE_URL"] = "https://app.example.com"
	env["TRUSTREADYD_CHROME_DP_ADDR"] = "chrome:9222"
	// API
	env["TRUSTREADYD_API_ADDR"] = "0.0.0.0:8080"
	env["TRUSTREADYD_API_CORS_ALLOWED_ORIGINS"] = "https://app.example.com,https://admin.example.com"
	env["TRUSTREADYD_API_PROXY_PROTOCOL_TRUSTED_PROXIES"] = "10.0.0.1,10.0.0.2"
	env["TRUSTREADYD_API_GRAPHQL_PARSER_TOKEN_LIMIT"] = "20000"
	env["TRUSTREADYD_API_GRAPHQL_COMPLEXITY_LIMIT"] = "5000"
	env["TRUSTREADYD_API_GRAPHQL_QUERY_CACHE_SIZE"] = "2000"
	env["TRUSTREADYD_API_GRAPHQL_DISABLE_SUGGESTION"] = "false"
	// PG
	env["TRUSTREADYD_PG_ADDR"] = "postgres.example.com:5432"
	env["TRUSTREADYD_PG_USERNAME"] = "trustready"
	env["TRUSTREADYD_PG_PASSWORD"] = "secret123"
	env["TRUSTREADYD_PG_DATABASE"] = "trustready_prod"
	env["TRUSTREADYD_PG_POOL_SIZE"] = "200"
	env["TRUSTREADYD_PG_MIN_POOL_SIZE"] = "25"
	env["TRUSTREADYD_PG_MAX_CONN_IDLE_TIME_SECONDS"] = "900"
	env["TRUSTREADYD_PG_MAX_CONN_LIFETIME_SECONDS"] = "7200"
	env["TRUSTREADYD_PG_MAX_CONN_LIFETIME_JITTER_SECONDS"] = "600"
	env["TRUSTREADYD_PG_HEALTH_CHECK_PERIOD_SECONDS"] = "30"
	env["TRUSTREADYD_PG_DEBUG"] = "true"
	// Auth
	env["TRUSTREADYD_AUTH_DISABLE_SIGNUP"] = "true"
	env["TRUSTREADYD_AUTH_INVITATION_TOKEN_VALIDITY"] = "7200"
	env["TRUSTREADYD_AUTH_PASSWORD_RESET_TOKEN_VALIDITY"] = "1800"
	env["TRUSTREADYD_AUTH_MAGIC_LINK_TOKEN_VALIDITY"] = "600"
	env["TRUSTREADYD_AUTH_EMAIL_CONFIRMATION_TOKEN_VALIDITY"] = "43200"
	env["TRUSTREADYD_AUTH_COOKIE_DOMAIN"] = ".example.com"
	env["TRUSTREADYD_AUTH_COOKIE_DURATION"] = "48"
	env["TRUSTREADYD_AUTH_COOKIE_SAMESITE"] = "strict"
	// SAML
	env["TRUSTREADYD_SAML_DOMAIN_VERIFICATION_INTERVAL_SECONDS"] = "120"
	env["TRUSTREADYD_SAML_DOMAIN_VERIFICATION_RESOLVER_ADDR"] = "1.1.1.1:53"
	// Trust center
	env["TRUSTREADYD_TRUST_CENTER_HTTP_ADDR"] = ":8080"
	env["TRUSTREADYD_TRUST_CENTER_HTTPS_ADDR"] = ":8443"
	env["TRUSTREADYD_TRUST_CENTER_BASE_DOMAIN"] = "trustreadypage.example.com"
	env["TRUSTREADYD_TRUST_CENTER_TLS_MODE"] = "external"
	env["TRUSTREADYD_TRUST_CENTER_PROXY_PROTOCOL_TRUSTED_PROXIES"] = "10.0.1.1,10.0.1.2"
	// AWS
	env["TRUSTREADYD_AWS_REGION"] = "eu-west-1"
	env["TRUSTREADYD_AWS_BUCKET"] = "trustready-files"
	env["TRUSTREADYD_AWS_ACCESS_KEY_ID"] = "AKIAIOSFODNN7EXAMPLE"
	env["TRUSTREADYD_AWS_SECRET_ACCESS_KEY"] = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	env["TRUSTREADYD_AWS_ENDPOINT"] = "https://s3.example.com"
	env["TRUSTREADYD_AWS_USE_PATH_STYLE"] = "true"
	// Notifications
	env["TRUSTREADYD_WEBHOOK_SENDER_INTERVAL"] = "10"
	env["TRUSTREADYD_WEBHOOK_REQUEST_TIMEOUT"] = "20"
	env["TRUSTREADYD_WEBHOOK_CACHE_TTL"] = "3600"
	env["TRUSTREADYD_WEBHOOK_STALE_AFTER"] = "600"
	env["TRUSTREADYD_WEBHOOK_RETRY_BASE"] = "60"
	env["TRUSTREADYD_WEBHOOK_RETRY_MAX"] = "7200"
	env["TRUSTREADYD_WEBHOOK_MAX_CONCURRENCY"] = "8"
	env["TRUSTREADYD_CONNECTOR_SLACK_SIGNING_SECRET"] = "slack-signing-secret"
	env["TRUSTREADYD_DOCUMENT_NOTIFICATION_INTERVAL"] = "120"
	env["TRUSTREADYD_DOCUMENT_NOTIFICATION_DEBOUNCE_DELAY"] = "60"
	env["TRUSTREADYD_DOCUMENT_NOTIFICATION_REMINDER_INTERVAL"] = "43200"
	// Firecrawl
	env["TRUSTREADYD_FIRECRAWL_API_KEY"] = "fc-test-key"
	// Agents — providers
	env["TRUSTREADYD_OPENAI_API_KEY"] = "sk-test-key"
	env["TRUSTREADYD_ANTHROPIC_API_KEY"] = "sk-ant-test-key"
	// Agents — default
	env["TRUSTREADYD_AGENT_DEFAULT_PROVIDER"] = "openai"
	env["TRUSTREADYD_AGENT_DEFAULT_MODEL_NAME"] = "gpt-4-turbo"
	env["TRUSTREADYD_AGENT_DEFAULT_TEMPERATURE"] = "0.5"
	env["TRUSTREADYD_AGENT_DEFAULT_MAX_TOKENS"] = "8192"
	// Agents — evidence-describer override
	env["TRUSTREADYD_AGENT_EVIDENCE_DESCRIBER_PROVIDER"] = "anthropic"
	env["TRUSTREADYD_AGENT_EVIDENCE_DESCRIBER_MODEL_NAME"] = "claude-sonnet-4-20250514"
	env["TRUSTREADYD_AGENT_EVIDENCE_DESCRIBER_TEMPERATURE"] = "0.2"
	env["TRUSTREADYD_AGENT_EVIDENCE_DESCRIBER_MAX_TOKENS"] = "4096"
	// Agents — third-party-vetter override
	env["TRUSTREADYD_AGENT_THIRD_PARTY_VETTER_PROVIDER"] = "openai"
	env["TRUSTREADYD_AGENT_THIRD_PARTY_VETTER_MODEL_NAME"] = "gpt-4o"
	env["TRUSTREADYD_AGENT_THIRD_PARTY_VETTER_TEMPERATURE"] = "0.3"
	env["TRUSTREADYD_AGENT_THIRD_PARTY_VETTER_MAX_TOKENS"] = "8192"
	// Agents — tracker-mapping override
	env["TRUSTREADYD_AGENT_TRACKER_MAPPING_PROVIDER"] = "openai"
	env["TRUSTREADYD_AGENT_TRACKER_MAPPING_MODEL_NAME"] = "gpt-4o-mini"
	env["TRUSTREADYD_AGENT_TRACKER_MAPPING_TEMPERATURE"] = "0.1"
	env["TRUSTREADYD_AGENT_TRACKER_MAPPING_MAX_TOKENS"] = "1024"
	// Agents — tracker-enrichment override
	env["TRUSTREADYD_AGENT_TRACKER_ENRICHMENT_PROVIDER"] = "openai"
	env["TRUSTREADYD_AGENT_TRACKER_ENRICHMENT_MODEL_NAME"] = "gpt-4o"
	env["TRUSTREADYD_AGENT_TRACKER_ENRICHMENT_TEMPERATURE"] = "0.2"
	env["TRUSTREADYD_AGENT_TRACKER_ENRICHMENT_MAX_TOKENS"] = "2048"
	// Tracker worker tuning override
	env["TRUSTREADYD_TRACKER_MAPPING_INTERVAL"] = "20"
	env["TRUSTREADYD_TRACKER_MAPPING_MAX_CONCURRENCY"] = "5"
	env["TRUSTREADYD_TRACKER_MAPPING_STALE_AFTER"] = "1200"
	env["TRUSTREADYD_TRACKER_MAPPING_AGENT_TIMEOUT"] = "30"
	env["TRUSTREADYD_TRACKER_MAPPING_AGENT_MAX_TURNS"] = "6"
	env["TRUSTREADYD_COMMON_PATTERN_ENRICHMENT_INTERVAL"] = "15"
	env["TRUSTREADYD_COMMON_PATTERN_ENRICHMENT_MAX_CONCURRENCY"] = "4"
	env["TRUSTREADYD_COMMON_PATTERN_ENRICHMENT_STALE_AFTER"] = "900"
	env["TRUSTREADYD_COMMON_PATTERN_ENRICHMENT_AGENT_TIMEOUT"] = "50"
	env["TRUSTREADYD_COMMON_PATTERN_ENRICHMENT_AGENT_MAX_TURNS"] = "5"
	// Common third party enrichment agent + worker tuning override
	env["TRUSTREADYD_AGENT_COMMON_THIRD_PARTY_ENRICHMENT_PROVIDER"] = "openai"
	env["TRUSTREADYD_AGENT_COMMON_THIRD_PARTY_ENRICHMENT_MODEL_NAME"] = "gpt-4o"
	env["TRUSTREADYD_AGENT_COMMON_THIRD_PARTY_ENRICHMENT_MAX_TOKENS"] = "16384"
	env["TRUSTREADYD_COMMON_THIRD_PARTY_ENRICHMENT_INTERVAL"] = "25"
	env["TRUSTREADYD_COMMON_THIRD_PARTY_ENRICHMENT_MAX_CONCURRENCY"] = "2"
	env["TRUSTREADYD_COMMON_THIRD_PARTY_ENRICHMENT_STALE_AFTER"] = "1200"
	env["TRUSTREADYD_COMMON_THIRD_PARTY_ENRICHMENT_AGENT_TIMEOUT"] = "120"
	env["TRUSTREADYD_COMMON_THIRD_PARTY_ENRICHMENT_AGENT_MAX_TURNS"] = "8"
	env["TRUSTREADYD_COMMON_THIRD_PARTY_ENRICHMENT_CONFIDENCE_THRESHOLD"] = "0.85"
	env["TRUSTREADYD_COMMON_THIRD_PARTY_ENRICHMENT_MAX_ATTEMPTS"] = "5"
	env["TRUSTREADYD_THIRD_PARTY_VETTING_INTERVAL"] = "15"
	env["TRUSTREADYD_THIRD_PARTY_VETTING_STALE_AFTER"] = "1800"
	env["TRUSTREADYD_THIRD_PARTY_VETTING_MAX_CONCURRENCY"] = "2"
	// Custom domains
	env["TRUSTREADYD_CUSTOM_DOMAINS_RESOLVER_ADDR"] = "1.1.1.1:53"
	env["TRUSTREADYD_ACME_ACCOUNT_KEY"] = testECKeyPEM()
	// SCIM bridge
	env["TRUSTREADYD_SCIM_BRIDGE_SYNC_INTERVAL"] = "1800"
	env["TRUSTREADYD_SCIM_BRIDGE_POLL_INTERVAL"] = "60"
	// ESign
	env["TRUSTREADYD_ESIGN_TSA_URL"] = "http://custom.tsa.example.com"
	// Branding
	env["TRUSTREADYD_BRANDING"] = "false"
	// Cookie banner
	env["TRUSTREADYD_COOKIE_BANNER_TCF_CMP_ID"] = "123"
	// Slackbot
	env["TRUSTREADYD_SLACKBOT_ENABLED"] = "true"
	env["TRUSTREADYD_SLACKBOT_SIGNING_SECRET"] = "slackbot-signing-secret"
	env["TRUSTREADYD_SLACKBOT_CLIENT_ID"] = "slackbot-client-id"
	env["TRUSTREADYD_SLACKBOT_CLIENT_SECRET"] = "slackbot-client-secret"
	env["TRUSTREADYD_SLACKBOT_REDIRECT_URI"] = "https://console.example.com/api/console/v1/slackbot/install/complete"
	env["TRUSTREADYD_AGENT_SLACKBOT_PROVIDER"] = "openai"
	env["TRUSTREADYD_AGENT_SLACKBOT_MODEL_NAME"] = "gpt-4o-mini"
	env["TRUSTREADYD_AGENT_SLACKBOT_MAX_TOKENS"] = "2048"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	// Unit
	assert.Equal(t, "0.0.0.0:9090", cfg.Unit.Metrics.Addr)
	assert.Equal(t, "jaeger:4317", cfg.Unit.Tracing.Addr)
	assert.Equal(t, 1024, cfg.Unit.Tracing.MaxBatchSize)
	// TrustReadyd
	assert.Equal(t, "https://app.example.com", cfg.TrustReadyd.BaseURL)
	assert.Equal(t, "chrome:9222", cfg.TrustReadyd.ChromeDPAddr)
	// API
	assert.Equal(t, "0.0.0.0:8080", cfg.TrustReadyd.Api.Addr)
	assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, cfg.TrustReadyd.Api.ProxyProtocol.TrustedProxies)
	assert.Equal(t, []string{"https://app.example.com", "https://admin.example.com"}, cfg.TrustReadyd.Api.Cors.AllowedOrigins)
	assert.Equal(t, 20000, cfg.TrustReadyd.Api.GraphQL.ParserTokenLimit)
	assert.Equal(t, 5000, cfg.TrustReadyd.Api.GraphQL.ComplexityLimit)
	assert.Equal(t, 2000, cfg.TrustReadyd.Api.GraphQL.QueryCacheSize)
	assert.False(t, cfg.TrustReadyd.Api.GraphQL.DisableSuggestion)
	// PG
	assert.Equal(t, "postgres.example.com:5432", cfg.TrustReadyd.Pg.Addr)
	assert.Equal(t, "trustready", cfg.TrustReadyd.Pg.Username)
	assert.Equal(t, "secret123", cfg.TrustReadyd.Pg.Password)
	assert.Equal(t, "trustready_prod", cfg.TrustReadyd.Pg.Database)
	assert.Equal(t, int32(200), cfg.TrustReadyd.Pg.PoolSize)
	assert.Equal(t, int32(25), cfg.TrustReadyd.Pg.MinPoolSize)
	assert.Equal(t, 900, cfg.TrustReadyd.Pg.MaxConnIdleTimeSeconds)
	assert.Equal(t, 7200, cfg.TrustReadyd.Pg.MaxConnLifetimeSeconds)
	assert.Equal(t, 600, cfg.TrustReadyd.Pg.MaxConnLifetimeJitterSeconds)
	assert.Equal(t, 30, cfg.TrustReadyd.Pg.HealthCheckPeriodSeconds)
	assert.True(t, cfg.TrustReadyd.Pg.Debug)
	// Auth
	assert.True(t, cfg.TrustReadyd.Auth.DisableSignup)
	assert.Equal(t, 7200, cfg.TrustReadyd.Auth.InvitationConfirmationTokenValidity)
	assert.Equal(t, 1800, cfg.TrustReadyd.Auth.PasswordResetTokenValidity)
	assert.Equal(t, 600, cfg.TrustReadyd.Auth.MagicLinkTokenValidity)
	assert.Equal(t, 43200, cfg.TrustReadyd.Auth.EmailConfirmationTokenValidity)
	assert.Equal(t, ".example.com", cfg.TrustReadyd.Auth.Cookie.Domain)
	assert.Equal(t, 48, cfg.TrustReadyd.Auth.Cookie.Duration)
	assert.Equal(t, trustreadydconfig.CookieSameSiteStrict, cfg.TrustReadyd.Auth.Cookie.SameSite)
	// SAML
	assert.Equal(t, 120, cfg.TrustReadyd.Auth.SAML.DomainVerificationIntervalSeconds)
	assert.Equal(t, "1.1.1.1:53", cfg.TrustReadyd.Auth.SAML.DomainVerificationResolverAddr)
	// Trust center
	assert.Equal(t, ":8080", cfg.TrustReadyd.CompliancePortal.HTTPAddr)
	assert.Equal(t, ":8443", cfg.TrustReadyd.CompliancePortal.HTTPSAddr)
	assert.Equal(t, "trustreadypage.example.com", cfg.TrustReadyd.CompliancePortal.BaseDomain)
	assert.Equal(t, trustreadydconfig.CompliancePortalTLSModeExternal, cfg.TrustReadyd.CompliancePortal.TLSMode)
	assert.Equal(t, []string{"10.0.1.1", "10.0.1.2"}, cfg.TrustReadyd.CompliancePortal.ProxyProtocol.TrustedProxies)
	// AWS
	assert.Equal(t, "eu-west-1", cfg.TrustReadyd.AWS.Region)
	assert.Equal(t, "trustready-files", cfg.TrustReadyd.AWS.Bucket)
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", cfg.TrustReadyd.AWS.AccessKeyID)
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", cfg.TrustReadyd.AWS.SecretAccessKey)
	assert.Equal(t, "https://s3.example.com", cfg.TrustReadyd.AWS.Endpoint)
	assert.True(t, cfg.TrustReadyd.AWS.UsePathStyle)
	// Notifications
	assert.Equal(t, "slack-signing-secret", cfg.TrustReadyd.Notifications.Slack.SigningSecret)
	assert.Equal(t, 10, cfg.TrustReadyd.Notifications.Webhook.SenderInterval)
	assert.Equal(t, 20, cfg.TrustReadyd.Notifications.Webhook.RequestTimeout)
	assert.Equal(t, 3600, cfg.TrustReadyd.Notifications.Webhook.CacheTTL)
	assert.Equal(t, 600, cfg.TrustReadyd.Notifications.Webhook.StaleAfter)
	assert.Equal(t, 60, cfg.TrustReadyd.Notifications.Webhook.RetryBase)
	assert.Equal(t, 7200, cfg.TrustReadyd.Notifications.Webhook.RetryMax)
	assert.Equal(t, 8, cfg.TrustReadyd.Notifications.Webhook.MaxConcurrency)
	assert.Equal(t, 120, cfg.TrustReadyd.Notifications.Document.Interval)
	assert.Equal(t, 60, cfg.TrustReadyd.Notifications.Document.DebounceDelay)
	assert.Equal(t, 43200, cfg.TrustReadyd.Notifications.Document.ReminderInterval)
	// Agents tools — Firecrawl
	assert.Equal(t, "fc-test-key", cfg.TrustReadyd.Agents.Tools.FirecrawlAPIKey)
	// Agents — providers
	assert.Equal(t, "openai", cfg.TrustReadyd.Agents.Providers["openai"].Type)
	assert.Equal(t, "sk-test-key", cfg.TrustReadyd.Agents.Providers["openai"].APIKey)
	assert.Equal(t, "anthropic", cfg.TrustReadyd.Agents.Providers["anthropic"].Type)
	assert.Equal(t, "sk-ant-test-key", cfg.TrustReadyd.Agents.Providers["anthropic"].APIKey)
	// Agents — default
	assert.Equal(t, "openai", cfg.TrustReadyd.Agents.Default.Provider)
	assert.Equal(t, "gpt-4-turbo", cfg.TrustReadyd.Agents.Default.ModelName)
	assert.Equal(t, new(0.5), cfg.TrustReadyd.Agents.Default.Temperature)
	assert.Equal(t, new(8192), cfg.TrustReadyd.Agents.Default.MaxTokens)
	// Agents — trustready inherits default (no overrides set)
	assert.Empty(t, cfg.TrustReadyd.Agents.TrustReady.Provider)
	assert.Empty(t, cfg.TrustReadyd.Agents.TrustReady.ModelName)
	// Agents — evidence-describer overrides
	assert.Equal(t, "anthropic", cfg.TrustReadyd.Agents.EvidenceDescriber.Provider)
	assert.Equal(t, "claude-sonnet-4-20250514", cfg.TrustReadyd.Agents.EvidenceDescriber.ModelName)
	assert.Equal(t, new(0.2), cfg.TrustReadyd.Agents.EvidenceDescriber.Temperature)
	assert.Equal(t, new(4096), cfg.TrustReadyd.Agents.EvidenceDescriber.MaxTokens)
	// Agents — third-party-vetter overrides
	assert.Equal(t, "openai", cfg.TrustReadyd.Agents.ThirdPartyVetter.Provider)
	assert.Equal(t, "gpt-4o", cfg.TrustReadyd.Agents.ThirdPartyVetter.ModelName)
	assert.Equal(t, new(0.3), cfg.TrustReadyd.Agents.ThirdPartyVetter.Temperature)
	assert.Equal(t, new(8192), cfg.TrustReadyd.Agents.ThirdPartyVetter.MaxTokens)
	// Agents — tracker-mapping overrides
	assert.Equal(t, "openai", cfg.TrustReadyd.Agents.TrackerMapping.Provider)
	assert.Equal(t, "gpt-4o-mini", cfg.TrustReadyd.Agents.TrackerMapping.ModelName)
	assert.Equal(t, new(0.1), cfg.TrustReadyd.Agents.TrackerMapping.Temperature)
	assert.Equal(t, new(1024), cfg.TrustReadyd.Agents.TrackerMapping.MaxTokens)
	// Agents — tracker-enrichment overrides
	assert.Equal(t, "openai", cfg.TrustReadyd.Agents.TrackerEnrichment.Provider)
	assert.Equal(t, "gpt-4o", cfg.TrustReadyd.Agents.TrackerEnrichment.ModelName)
	assert.Equal(t, new(0.2), cfg.TrustReadyd.Agents.TrackerEnrichment.Temperature)
	assert.Equal(t, new(2048), cfg.TrustReadyd.Agents.TrackerEnrichment.MaxTokens)
	// Tracker worker tuning — overrides
	assert.Equal(t, 20, cfg.TrustReadyd.TrackerMappingWorker.Interval)
	assert.Equal(t, 5, cfg.TrustReadyd.TrackerMappingWorker.MaxConcurrency)
	assert.Equal(t, 1200, cfg.TrustReadyd.TrackerMappingWorker.StaleAfter)
	assert.Equal(t, 30, cfg.TrustReadyd.TrackerMappingWorker.AgentTimeout)
	assert.Equal(t, 6, cfg.TrustReadyd.TrackerMappingWorker.AgentMaxTurns)
	assert.Equal(t, 15, cfg.TrustReadyd.CommonPatternEnrichmentWorker.Interval)
	assert.Equal(t, 4, cfg.TrustReadyd.CommonPatternEnrichmentWorker.MaxConcurrency)
	assert.Equal(t, 900, cfg.TrustReadyd.CommonPatternEnrichmentWorker.StaleAfter)
	assert.Equal(t, 50, cfg.TrustReadyd.CommonPatternEnrichmentWorker.AgentTimeout)
	assert.Equal(t, 5, cfg.TrustReadyd.CommonPatternEnrichmentWorker.AgentMaxTurns)
	assert.Equal(t, "openai", cfg.TrustReadyd.Agents.CommonThirdPartyEnrichment.Provider)
	assert.Equal(t, "gpt-4o", cfg.TrustReadyd.Agents.CommonThirdPartyEnrichment.ModelName)
	require.NotNil(t, cfg.TrustReadyd.Agents.CommonThirdPartyEnrichment.MaxTokens)
	assert.Equal(t, 16384, *cfg.TrustReadyd.Agents.CommonThirdPartyEnrichment.MaxTokens)
	assert.Equal(t, 25, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.Interval)
	assert.Equal(t, 2, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.MaxConcurrency)
	assert.Equal(t, 1200, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.StaleAfter)
	assert.Equal(t, 120, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.AgentTimeout)
	assert.Equal(t, 8, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.AgentMaxTurns)
	assert.Equal(t, 0.85, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.ConfidenceThreshold)
	assert.Equal(t, 5, cfg.TrustReadyd.CommonThirdPartyEnrichmentWorker.MaxAttempts)
	assert.Equal(t, 15, cfg.TrustReadyd.ThirdPartyVetting.Interval)
	assert.Equal(t, 1800, cfg.TrustReadyd.ThirdPartyVetting.StaleAfter)
	assert.Equal(t, 2, cfg.TrustReadyd.ThirdPartyVetting.MaxConcurrency)
	// Custom domains
	assert.Equal(t, "1.1.1.1:53", cfg.TrustReadyd.CustomDomains.ResolverAddr)
	assert.Equal(t, testECKeyPEM(), cfg.TrustReadyd.CustomDomains.ACME.AccountKey.PEM())
	// SCIM bridge
	assert.Equal(t, 1800, cfg.TrustReadyd.SCIMBridge.SyncInterval)
	assert.Equal(t, 60, cfg.TrustReadyd.SCIMBridge.PollInterval)
	// ESign
	assert.Equal(t, "http://custom.tsa.example.com", cfg.TrustReadyd.ESign.TSAURL)
	// Branding
	assert.False(t, cfg.TrustReadyd.Branding)
	// Cookie banner
	assert.Equal(t, 123, cfg.TrustReadyd.CookieBanner.TCFCMPID)
	// Slackbot
	assert.True(t, cfg.TrustReadyd.Slackbot.Enabled)
	assert.Equal(t, "slackbot-signing-secret", cfg.TrustReadyd.Slackbot.SigningSecret)
	assert.Equal(t, "slackbot-client-id", cfg.TrustReadyd.Slackbot.ClientID)
	assert.Equal(t, "slackbot-client-secret", cfg.TrustReadyd.Slackbot.ClientSecret)
	assert.Equal(t, "https://console.example.com/api/console/v1/slackbot/install/complete", cfg.TrustReadyd.Slackbot.RedirectURI)
	assert.Equal(t, "openai", cfg.TrustReadyd.Agents.Slackbot.Provider)
	assert.Equal(t, "gpt-4o-mini", cfg.TrustReadyd.Agents.Slackbot.ModelName)
	require.NotNil(t, cfg.TrustReadyd.Agents.Slackbot.MaxTokens)
	assert.Equal(t, 2048, *cfg.TrustReadyd.Agents.Slackbot.MaxTokens)
}

func TestBuilder_Build_GoogleWorkspaceConnector(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_CONNECTOR_GOOGLE_WORKSPACE_CLIENT_ID"] = "gw-client-id"
	env["TRUSTREADYD_CONNECTOR_GOOGLE_WORKSPACE_CLIENT_SECRET"] = "gw-client-secret"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.Connectors, 1)
	connector := cfg.TrustReadyd.Connectors[0]
	assert.Equal(t, "GOOGLE_WORKSPACE", connector.Provider)
	assert.Equal(t, "oauth2", string(connector.Protocol))
	rawConfig := connector.RawConfig.(trustreadydconfig.ConnectorConfigOAuth2)
	assert.Equal(t, "gw-client-id", rawConfig.ClientID)
	assert.Equal(t, "gw-client-secret", rawConfig.ClientSecret)
}

func TestBuilder_Build_GitHubAppConnector(t *testing.T) {
	t.Parallel()

	env := requiredEnv()
	env["TRUSTREADYD_CONNECTOR_GITHUB_APP_ID"] = "123456"
	env["TRUSTREADYD_CONNECTOR_GITHUB_APP_CLIENT_ID"] = "Iv1.example"
	env["TRUSTREADYD_CONNECTOR_GITHUB_APP_CLIENT_SECRET"] = "client-secret"
	env["TRUSTREADYD_CONNECTOR_GITHUB_APP_SLUG"] = "trustready"
	env["TRUSTREADYD_CONNECTOR_GITHUB_APP_PRIVATE_KEY"] = "private-key"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.Connectors, 1)
	c := cfg.TrustReadyd.Connectors[0]
	assert.Equal(t, "GITHUB", c.Provider)
	assert.Equal(t, connector.ProtocolGitHubApp, c.Protocol)

	raw := c.RawConfig.(trustreadydconfig.ConnectorConfigGitHubApp)
	assert.Equal(t, "123456", raw.AppID)
	assert.Equal(t, "Iv1.example", raw.ClientID)
	assert.Equal(t, "client-secret", raw.ClientSecret)
	assert.Equal(t, "trustready", raw.Slug)
	assert.Equal(t, "private-key", raw.PrivateKey)
}

func TestBuilder_Build_Microsoft365Connector(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_CONNECTOR_MICROSOFT_365_CLIENT_ID"] = "ms365-client-id"
	env["TRUSTREADYD_CONNECTOR_MICROSOFT_365_CLIENT_SECRET"] = "ms365-client-secret"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.Connectors, 1)
	connector := cfg.TrustReadyd.Connectors[0]
	assert.Equal(t, "MICROSOFT_365", connector.Provider)
	assert.Equal(t, "oauth2", string(connector.Protocol))
	rawConfig := connector.RawConfig.(trustreadydconfig.ConnectorConfigOAuth2)
	assert.Equal(t, "ms365-client-id", rawConfig.ClientID)
	assert.Equal(t, "ms365-client-secret", rawConfig.ClientSecret)
}

func TestBuilder_Build_AccessReviewConnectors(t *testing.T) {
	// All non-Vercel access-review providers added by this PR. Vercel
	// has its own dedicated test because it carries an additional
	// CONNECTOR_VERCEL_INTEGRATION_SLUG env var.
	providers := []string{
		"GITLAB", "BITBUCKET", "HEROKU", "PAGERDUTY",
		"ASANA", "NETLIFY", "CLICKUP", "MONDAY", "DATADOG",
		"ZENDESK", "LINEAR", "GOOGLE_ANALYTICS", "SQUARE",
		"CAL_COM", "CALENDLY", "ATTIO",
		"OVHCLOUD",
	}

	env := requiredEnv()
	for _, provider := range providers {
		env["TRUSTREADYD_CONNECTOR_"+provider+"_CLIENT_ID"] = strings.ToLower(provider) + "-id"
		env["TRUSTREADYD_CONNECTOR_"+provider+"_CLIENT_SECRET"] = strings.ToLower(provider) + "-secret"
	}

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.Connectors, len(providers))

	byProvider := make(map[string]trustreadydconfig.ConnectorConfig, len(cfg.TrustReadyd.Connectors))
	for _, c := range cfg.TrustReadyd.Connectors {
		byProvider[c.Provider] = c
	}

	for _, provider := range providers {
		c, ok := byProvider[provider]
		require.True(t, ok, "missing %s connector", provider)
		assert.Equal(t, "oauth2", string(c.Protocol))
		raw := c.RawConfig.(trustreadydconfig.ConnectorConfigOAuth2)
		assert.NotEmpty(t, raw.ClientID, "%s client-id", provider)
		assert.NotEmpty(t, raw.ClientSecret, "%s client-secret", provider)
		assert.Empty(t, raw.IntegrationSlug, "%s should not carry integration-slug", provider)
	}
}

func TestBuilder_Build_VercelConnector(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_CONNECTOR_VERCEL_CLIENT_ID"] = "vercel-id"
	env["TRUSTREADYD_CONNECTOR_VERCEL_CLIENT_SECRET"] = "vercel-secret"
	env["TRUSTREADYD_CONNECTOR_VERCEL_INTEGRATION_SLUG"] = "trustready-app"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.Connectors, 1)
	c := cfg.TrustReadyd.Connectors[0]
	assert.Equal(t, "VERCEL", c.Provider)
	assert.Equal(t, "oauth2", string(c.Protocol))
	raw := c.RawConfig.(trustreadydconfig.ConnectorConfigOAuth2)
	assert.Equal(t, "vercel-id", raw.ClientID)
	assert.Equal(t, "vercel-secret", raw.ClientSecret)
	assert.Equal(t, "trustready-app", raw.IntegrationSlug)
}

func TestBuilder_Build_LinearWebhookSecret(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_CONNECTOR_LINEAR_SYNC_CLIENT_ID"] = "linear-sync-client-id"
	env["TRUSTREADYD_CONNECTOR_LINEAR_SYNC_CLIENT_SECRET"] = "linear-sync-client-secret"
	env["TRUSTREADYD_CONNECTOR_LINEAR_SYNC_WEBHOOK_SECRET"] = "linear-sync-webhook-secret"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.Connectors, 1)
	c := cfg.TrustReadyd.Connectors[0]
	assert.Equal(t, "LINEAR_SYNC", c.Provider)
	assert.Equal(t, "oauth2", string(c.Protocol))
	raw := c.RawConfig.(trustreadydconfig.ConnectorConfigOAuth2)
	assert.Equal(t, "linear-sync-client-id", raw.ClientID)
	assert.Equal(t, "linear-sync-client-secret", raw.ClientSecret)
	assert.Equal(t, "linear-sync-webhook-secret", cfg.TrustReadyd.GetLinearWebhookSecret())
}

func TestBuilder_Build_SlackConnector(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_CONNECTOR_SLACK_CLIENT_ID"] = "slack-client-id"
	env["TRUSTREADYD_CONNECTOR_SLACK_CLIENT_SECRET"] = "slack-client-secret"
	env["TRUSTREADYD_CONNECTOR_SLACK_SIGNING_SECRET"] = "slack-signing-secret"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.Connectors, 1)
	connector := cfg.TrustReadyd.Connectors[0]
	assert.Equal(t, "SLACK", connector.Provider)
	assert.Equal(t, "oauth2", string(connector.Protocol))
	rawConfig := connector.RawConfig.(trustreadydconfig.ConnectorConfigOAuth2)
	assert.Equal(t, "slack-client-id", rawConfig.ClientID)
	assert.Equal(t, "slack-client-secret", rawConfig.ClientSecret)

	rawSettings := connector.RawSettings.(map[string]any)
	assert.Equal(t, "slack-signing-secret", rawSettings["signing-secret"])
}

func TestBuilder_Build_CrispConnector(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_CONNECTOR_CRISP_PLUGIN_TOKEN"] = "plugin-identifier:plugin-key"
	env["TRUSTREADYD_CONNECTOR_CRISP_PLUGIN_ID"] = "e979a1c3-2c41-4e93-a8ed-410ace27318e"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.Connectors, 1)
	connector := cfg.TrustReadyd.Connectors[0]
	assert.Equal(t, "CRISP", connector.Provider)
	assert.Equal(t, "api_key", string(connector.Protocol))
	rawConfig := connector.RawConfig.(trustreadydconfig.ConnectorConfigAPIKey)
	assert.Equal(t, "plugin-identifier:plugin-key", rawConfig.APIKey)
	assert.Equal(t, "e979a1c3-2c41-4e93-a8ed-410ace27318e", rawConfig.ResourceID)
}

func TestBuilder_Build_CrispConnectorAbsentWithoutToken(t *testing.T) {
	// Without the plugin token the connector must not be configured, which
	// is what keeps Crisp deactivated until Crisp validates the plugin.
	b := NewBuilder(NewResolver(mockEnv(requiredEnv())))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	for _, c := range cfg.TrustReadyd.Connectors {
		assert.NotEqual(t, "CRISP", c.Provider)
	}
}

func TestBuilder_Build_CrispConnectorAbsentWithoutPluginID(t *testing.T) {
	// The plugin token alone is not enough: the per-website plugin API needs
	// the plugin ID to verify website ownership, so a half-configured Crisp
	// connector stays hidden rather than activating in a broken state.
	env := requiredEnv()
	env["TRUSTREADYD_CONNECTOR_CRISP_PLUGIN_TOKEN"] = "plugin-identifier:plugin-key"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	for _, c := range cfg.TrustReadyd.Connectors {
		assert.NotEqual(t, "CRISP", c.Provider)
	}
}

func TestBuilder_Build_SAMLAutoGeneration(t *testing.T) {
	b := NewBuilder(NewResolver(mockEnv(requiredEnv())))

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Contains(t, cfg.TrustReadyd.Auth.SAML.Certificate, "-----BEGIN CERTIFICATE-----")
	assert.Contains(t, cfg.TrustReadyd.Auth.SAML.Certificate, "-----END CERTIFICATE-----")
	assert.Contains(t, cfg.TrustReadyd.Auth.SAML.PrivateKey.PEM(), "-----BEGIN RSA PRIVATE KEY-----")
	assert.Contains(t, cfg.TrustReadyd.Auth.SAML.PrivateKey.PEM(), "-----END RSA PRIVATE KEY-----")
	assert.NotNil(t, cfg.TrustReadyd.Auth.SAML.PrivateKey.PrivateKey())
}

func TestBuilder_Build_SAMLFromEnv(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_SAML_CERTIFICATE"] = "env-cert"
	env["TRUSTREADYD_SAML_PRIVATE_KEY"] = testOtherSigningKeyPEM()

	b := NewBuilder(NewResolver(mockEnv(env)))

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "env-cert", cfg.TrustReadyd.Auth.SAML.Certificate)
	assert.Equal(t, testOtherSigningKeyPEM(), cfg.TrustReadyd.Auth.SAML.PrivateKey.PEM())
}

func TestBuilder_Build_SAMLPreset(t *testing.T) {
	b := NewBuilder(NewResolver(mockEnv(requiredEnv())))
	b.samlCertificate = "preset-cert"
	b.samlPrivateKey = testOtherSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "preset-cert", cfg.TrustReadyd.Auth.SAML.Certificate)
	assert.Equal(t, testOtherSigningKeyPEM(), cfg.TrustReadyd.Auth.SAML.PrivateKey.PEM())
}

func TestBuilder_Build_OAuth2Defaults(t *testing.T) {
	b := NewBuilder(NewResolver(mockEnv(requiredEnv())))

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.Auth.OAuth2Server.SigningKeys, 1)
	sk := cfg.TrustReadyd.Auth.OAuth2Server.SigningKeys[0]
	assert.Equal(t, testSigningKeyPEM(), sk.PrivateKey.PEM())
	assert.NotNil(t, sk.PrivateKey.PrivateKey())
	assert.Equal(t, "default", sk.KID)
	assert.True(t, sk.Active)

	assert.Equal(t, 3600, cfg.TrustReadyd.Auth.OAuth2Server.AccessTokenDuration)
	assert.Equal(t, 2592000, cfg.TrustReadyd.Auth.OAuth2Server.RefreshTokenDuration)
	assert.Equal(t, 600, cfg.TrustReadyd.Auth.OAuth2Server.AuthorizationCodeDuration)
	assert.Equal(t, 600, cfg.TrustReadyd.Auth.OAuth2Server.DeviceCodeDuration)
	assert.Nil(t, cfg.TrustReadyd.Auth.OAuth2Server.CIMDAllowedClientIDs)
}

func TestBuilder_Build_OAuth2FromEnv(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_OAUTH2_SERVER_SIGNING_KEY"] = testOtherSigningKeyPEM()
	env["TRUSTREADYD_OAUTH2_SERVER_SIGNING_KEY_KID"] = "env-kid"
	env["TRUSTREADYD_OAUTH2_SERVER_ACCESS_TOKEN_DURATION"] = "10"
	env["TRUSTREADYD_OAUTH2_SERVER_REFRESH_TOKEN_DURATION"] = "20"
	env["TRUSTREADYD_OAUTH2_SERVER_AUTHORIZATION_CODE_DURATION"] = "30"
	env["TRUSTREADYD_OAUTH2_SERVER_DEVICE_CODE_DURATION"] = "40"
	env["TRUSTREADYD_OAUTH2_SERVER_CIMD_ALLOWED_CLIENT_IDS"] = "https://chatgpt.com/oauth/client.json,https://claude.ai/oauth/client.json"

	b := NewBuilder(NewResolver(mockEnv(env)))

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.Auth.OAuth2Server.SigningKeys, 1)
	sk := cfg.TrustReadyd.Auth.OAuth2Server.SigningKeys[0]
	assert.Equal(t, testOtherSigningKeyPEM(), sk.PrivateKey.PEM())
	assert.Equal(t, "env-kid", sk.KID)
	assert.True(t, sk.Active)

	assert.Equal(t, 10, cfg.TrustReadyd.Auth.OAuth2Server.AccessTokenDuration)
	assert.Equal(t, 20, cfg.TrustReadyd.Auth.OAuth2Server.RefreshTokenDuration)
	assert.Equal(t, 30, cfg.TrustReadyd.Auth.OAuth2Server.AuthorizationCodeDuration)
	assert.Equal(t, 40, cfg.TrustReadyd.Auth.OAuth2Server.DeviceCodeDuration)
	assert.Equal(
		t,
		[]string{
			"https://chatgpt.com/oauth/client.json",
			"https://claude.ai/oauth/client.json",
		},
		cfg.TrustReadyd.Auth.OAuth2Server.CIMDAllowedClientIDs,
	)
}

func TestBuilder_Build_OAuth2Preset(t *testing.T) {
	env := requiredEnv()
	delete(env, "OAUTH2_SERVER_SIGNING_KEY")

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.oauth2SigningKey = testOtherSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.Auth.OAuth2Server.SigningKeys, 1)
	assert.Equal(t, testOtherSigningKeyPEM(), cfg.TrustReadyd.Auth.OAuth2Server.SigningKeys[0].PrivateKey.PEM())
}

// A key that cannot be decoded fails in trustreadyd-bootstrap, rather than at the
// next start of trustreadyd.
func TestBuilder_Build_InvalidSigningKey(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_OAUTH2_SERVER_SIGNING_KEY"] = "not-a-pem-key"

	b := NewBuilder(NewResolver(mockEnv(env)))

	_, err := b.Build()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot get OAuth2 server signing key")
}

func TestBuilder_Build_IdentityFederationDisabledByDefault(t *testing.T) {
	b := NewBuilder(NewResolver(mockEnv(requiredEnv())))

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.False(t, cfg.TrustReadyd.IdentityFederation.Enabled)
	assert.Empty(t, cfg.TrustReadyd.IdentityFederation.IssuerBaseURL)
	assert.Empty(t, cfg.TrustReadyd.IdentityFederation.SigningKeys)
	assert.Equal(t, cloudaws.DefaultCloudFormationTemplateURL, cfg.TrustReadyd.IdentityFederation.CloudFormationTemplateURL)
	assert.Equal(t, cloudaws.DefaultTerraformModuleSource, cfg.TrustReadyd.IdentityFederation.TerraformModuleSource)
	assert.Equal(t, cloudgcp.DefaultTerraformModuleSource, cfg.TrustReadyd.IdentityFederation.GCPTerraformModuleSource)
	assert.Equal(t, cloudazure.DefaultTerraformModuleSource, cfg.TrustReadyd.IdentityFederation.AzureTerraformModuleSource)
	assert.Empty(t, cfg.TrustReadyd.IdentityFederation.AWSEndpoint)
}

func TestBuilder_Build_IdentityFederationInstallArtifactsFromEnv(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_IDENTITY_FEDERATION_CLOUDFORMATION_TEMPLATE_URL"] = "https://example.com/audit-role.yaml"
	env["TRUSTREADYD_IDENTITY_FEDERATION_TERRAFORM_MODULE_SOURCE"] = "example/terraform-aws-audit-role"
	env["TRUSTREADYD_IDENTITY_FEDERATION_GCP_TERRAFORM_MODULE_SOURCE"] = "example/terraform-gcp-audit-role"
	env["TRUSTREADYD_IDENTITY_FEDERATION_AZURE_TERRAFORM_MODULE_SOURCE"] = "example/terraform-azurerm-audit-role"
	env["TRUSTREADYD_IDENTITY_FEDERATION_AWS_ENDPOINT"] = "http://127.0.0.1:4566"

	b := NewBuilder(NewResolver(mockEnv(env)))

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "https://example.com/audit-role.yaml", cfg.TrustReadyd.IdentityFederation.CloudFormationTemplateURL)
	assert.Equal(t, "example/terraform-aws-audit-role", cfg.TrustReadyd.IdentityFederation.TerraformModuleSource)
	assert.Equal(t, "example/terraform-gcp-audit-role", cfg.TrustReadyd.IdentityFederation.GCPTerraformModuleSource)
	assert.Equal(t, "example/terraform-azurerm-audit-role", cfg.TrustReadyd.IdentityFederation.AzureTerraformModuleSource)
	assert.Equal(t, "http://127.0.0.1:4566", cfg.TrustReadyd.IdentityFederation.AWSEndpoint)
}

func TestBuilder_Build_IdentityFederationDisabledSkipsSigningKey(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_IDENTITY_FEDERATION_ENABLED"] = "false"
	env["TRUSTREADYD_IDENTITY_FEDERATION_SIGNING_KEY"] = "unused-identity-federation-key"

	b := NewBuilder(NewResolver(mockEnv(env)))

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.False(t, cfg.TrustReadyd.IdentityFederation.Enabled)
	assert.Empty(t, cfg.TrustReadyd.IdentityFederation.SigningKeys)
}

func TestBuilder_Build_IdentityFederationEnabledFromEnv(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_IDENTITY_FEDERATION_ENABLED"] = "true"
	env["TRUSTREADYD_IDENTITY_FEDERATION_ISSUER_BASE_URL"] = "https://trustreadyidentity.com"
	env["TRUSTREADYD_IDENTITY_FEDERATION_SIGNING_KEY"] = testSigningKeyPEM()
	env["TRUSTREADYD_IDENTITY_FEDERATION_SIGNING_KEY_KID"] = "env-identity-federation-kid"

	b := NewBuilder(NewResolver(mockEnv(env)))

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.True(t, cfg.TrustReadyd.IdentityFederation.Enabled)
	assert.Equal(t, "https://trustreadyidentity.com", cfg.TrustReadyd.IdentityFederation.IssuerBaseURL)

	require.Len(t, cfg.TrustReadyd.IdentityFederation.SigningKeys, 1)
	sk := cfg.TrustReadyd.IdentityFederation.SigningKeys[0]
	assert.Equal(t, testSigningKeyPEM(), sk.PrivateKey.PEM())
	assert.NotNil(t, sk.PrivateKey.PrivateKey())
	assert.Equal(t, "env-identity-federation-kid", sk.KID)
	assert.True(t, sk.Active)
}

func TestBuilder_Build_IdentityFederationEnabledDefaultsKID(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_IDENTITY_FEDERATION_ENABLED"] = "true"
	env["TRUSTREADYD_IDENTITY_FEDERATION_SIGNING_KEY"] = testSigningKeyPEM()

	b := NewBuilder(NewResolver(mockEnv(env)))

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.IdentityFederation.SigningKeys, 1)
	assert.Equal(t, "default", cfg.TrustReadyd.IdentityFederation.SigningKeys[0].KID)
}

func TestBuilder_Build_IdentityFederationEnabledRequiresSigningKey(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_IDENTITY_FEDERATION_ENABLED"] = "true"

	b := NewBuilder(NewResolver(mockEnv(env)))

	_, err := b.Build()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "TRUSTREADYD_IDENTITY_FEDERATION_SIGNING_KEY")
}

func TestBuilder_Build_IdentityFederationPreviousSigningKey(t *testing.T) {
	t.Parallel()

	enabledEnv := func() map[string]string {
		env := requiredEnv()
		env["TRUSTREADYD_IDENTITY_FEDERATION_ENABLED"] = "true"
		env["TRUSTREADYD_IDENTITY_FEDERATION_SIGNING_KEY"] = testSigningKeyPEM()
		env["TRUSTREADYD_IDENTITY_FEDERATION_SIGNING_KEY_KID"] = "current"

		return env
	}

	t.Run(
		"retired key stays published as inactive",
		func(t *testing.T) {
			t.Parallel()

			env := enabledEnv()
			env["TRUSTREADYD_IDENTITY_FEDERATION_PREVIOUS_SIGNING_KEY"] = testOtherSigningKeyPEM()
			env["TRUSTREADYD_IDENTITY_FEDERATION_PREVIOUS_SIGNING_KEY_KID"] = "retired"

			b := NewBuilder(NewResolver(mockEnv(env)))

			cfg, err := b.Build()
			require.NoError(t, err)

			require.Len(t, cfg.TrustReadyd.IdentityFederation.SigningKeys, 2)

			active := cfg.TrustReadyd.IdentityFederation.SigningKeys[0]
			assert.Equal(t, testSigningKeyPEM(), active.PrivateKey.PEM())
			assert.Equal(t, "current", active.KID)
			assert.True(t, active.Active)

			// Published but never signing: a cloud provider that already cached
			// the key set can still verify a token minted before the rotation.
			retired := cfg.TrustReadyd.IdentityFederation.SigningKeys[1]
			assert.Equal(t, testOtherSigningKeyPEM(), retired.PrivateKey.PEM())
			assert.Equal(t, "retired", retired.KID)
			assert.False(t, retired.Active)
		},
	)

	t.Run(
		"omitted leaves a single active key",
		func(t *testing.T) {
			t.Parallel()

			b := NewBuilder(NewResolver(mockEnv(enabledEnv())))

			cfg, err := b.Build()
			require.NoError(t, err)

			require.Len(t, cfg.TrustReadyd.IdentityFederation.SigningKeys, 1)
			assert.True(t, cfg.TrustReadyd.IdentityFederation.SigningKeys[0].Active)
		},
	)

	t.Run(
		"key without kid",
		func(t *testing.T) {
			t.Parallel()

			env := enabledEnv()
			env["TRUSTREADYD_IDENTITY_FEDERATION_PREVIOUS_SIGNING_KEY"] = testOtherSigningKeyPEM()

			b := NewBuilder(NewResolver(mockEnv(env)))

			_, err := b.Build()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "TRUSTREADYD_IDENTITY_FEDERATION_PREVIOUS_SIGNING_KEY_KID is required")
		},
	)

	t.Run(
		"kid without key",
		func(t *testing.T) {
			t.Parallel()

			env := enabledEnv()
			env["TRUSTREADYD_IDENTITY_FEDERATION_PREVIOUS_SIGNING_KEY_KID"] = "retired"

			b := NewBuilder(NewResolver(mockEnv(env)))

			_, err := b.Build()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "TRUSTREADYD_IDENTITY_FEDERATION_PREVIOUS_SIGNING_KEY is required")
		},
	)

	t.Run(
		"kid equal to the active kid",
		func(t *testing.T) {
			t.Parallel()

			env := enabledEnv()
			env["TRUSTREADYD_IDENTITY_FEDERATION_PREVIOUS_SIGNING_KEY"] = testOtherSigningKeyPEM()
			env["TRUSTREADYD_IDENTITY_FEDERATION_PREVIOUS_SIGNING_KEY_KID"] = "current"

			b := NewBuilder(NewResolver(mockEnv(env)))

			_, err := b.Build()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "must differ from TRUSTREADYD_IDENTITY_FEDERATION_SIGNING_KEY_KID")
		},
	)
}

func TestBuilder_Build_PgCABundleFromEnv(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_PG_CA_BUNDLE"] = "test-ca-bundle-content"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "test-ca-bundle-content", cfg.TrustReadyd.Pg.CACertBundle)
}

func TestBuilder_Build_PgCABundleFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	caFile := filepath.Join(tmpDir, "ca-bundle.pem")
	err := os.WriteFile(caFile, []byte("ca-bundle-from-file"), 0644)
	require.NoError(t, err)

	env := requiredEnv()
	env["TRUSTREADYD_PG_CA_BUNDLE_PATH"] = caFile

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Equal(t, "ca-bundle-from-file", cfg.TrustReadyd.Pg.CACertBundle)
}

func TestBuilder_Build_AuthCookieSameSiteInvalid(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_AUTH_COOKIE_SAMESITE"] = "invalid"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	_, err := b.Build()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TRUSTREADYD_AUTH_COOKIE_SAMESITE")
}

func TestBuilder_Build_AuthCookieSameSiteNoneRequiresSecure(t *testing.T) {
	env := requiredEnv()
	env["TRUSTREADYD_AUTH_COOKIE_SAMESITE"] = "none"
	env["TRUSTREADYD_AUTH_COOKIE_SECURE"] = "false"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	_, err := b.Build()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "secure")
}

func TestBuilder_parseOriginsList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single origin",
			input: "http://localhost:8080",
			want:  []string{"http://localhost:8080"},
		},
		{
			name:  "multiple origins",
			input: "http://localhost:8080,https://example.com",
			want:  []string{"http://localhost:8080", "https://example.com"},
		},
		{
			name:  "quoted origins",
			input: `"http://localhost:8080","https://example.com"`,
			want:  []string{"http://localhost:8080", "https://example.com"},
		},
		{
			name:  "with spaces",
			input: "http://localhost:8080 , https://example.com",
			want:  []string{"http://localhost:8080", "https://example.com"},
		},
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(nil)
			got := b.parseOriginsList(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestBuilder_Build_ConnectorEndpointOverrides covers the staging case the
// override exists for: pointing DocuSign at the vendor's demo hosts from the
// deployment's environment, with no code change and no effect on any other
// provider.
func TestBuilder_Build_ConnectorEndpointOverrides(t *testing.T) {
	t.Parallel()

	env := requiredEnv()
	env["TRUSTREADYD_CONNECTOR_DOCUSIGN_ENDPOINT_AUTH"] = "https://account-d.docusign.com/oauth/auth"
	env["TRUSTREADYD_CONNECTOR_DOCUSIGN_ENDPOINT_TOKEN"] = "https://account-d.docusign.com/oauth/token"
	env["TRUSTREADYD_CONNECTOR_DOCUSIGN_ENDPOINT_PROBE"] = "https://account-d.docusign.com/oauth/userinfo"
	env["TRUSTREADYD_CONNECTOR_DOCUSIGN_ENDPOINT_IDENTITY"] = "https://account-d.docusign.com/oauth/userinfo"
	env["TRUSTREADYD_CONNECTOR_SLACK_ENDPOINT_AUTH"] = "https://auth.slack.test/oauth/v2/authorize"
	env["TRUSTREADYD_CONNECTOR_SLACK_ENDPOINT_TOKEN"] = "https://api.slack.test/oauth.v2.access"
	env["TRUSTREADYD_CONNECTOR_SLACK_ENDPOINT_API_BASE"] = "https://api.slack.test"

	b := NewBuilder(NewResolver(mockEnv(env)))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	require.Len(t, cfg.TrustReadyd.ConnectorEndpoints, 2, "only the configured providers should appear")

	got, ok := cfg.TrustReadyd.ConnectorEndpoints["DOCUSIGN"]
	require.True(t, ok)

	assert.Equal(t, "https://account-d.docusign.com/oauth/auth", got.Auth)
	assert.Equal(t, "https://account-d.docusign.com/oauth/token", got.Token)
	assert.Equal(t, "https://account-d.docusign.com/oauth/userinfo", got.Probe)
	assert.Equal(t, "https://account-d.docusign.com/oauth/userinfo", got.Identity)
	assert.Empty(t, got.APIBase, "an unset field must stay empty so the compiled default survives")

	got, ok = cfg.TrustReadyd.ConnectorEndpoints["SLACK"]
	require.True(t, ok)

	assert.Equal(t, "https://auth.slack.test/oauth/v2/authorize", got.Auth)
	assert.Equal(t, "https://api.slack.test/oauth.v2.access", got.Token)
	assert.Equal(t, "https://api.slack.test", got.APIBase)
}

// TestBuilder_Build_NoConnectorEndpointOverrides pins the default: a
// deployment that configures nothing emits no override map at all, so every
// connector keeps the endpoints compiled into it.
func TestBuilder_Build_NoConnectorEndpointOverrides(t *testing.T) {
	t.Parallel()

	b := NewBuilder(NewResolver(mockEnv(requiredEnv())))
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	cfg, err := b.Build()
	require.NoError(t, err)

	assert.Empty(t, cfg.TrustReadyd.ConnectorEndpoints)
}

// TestBuilder_Build_ConnectorEndpointTypoUnknownProvider pins the fix for a
// typo the operator would otherwise never see: getEnv resolves an unset key
// to "" exactly like a typo'd one, so
// TRUSTREADYD_CONNECTOR_GITHBU_ENDPOINT_API_BASE silently did nothing before this
// scan existed, while the operator believed GitHub was repointed.
func TestBuilder_Build_ConnectorEndpointTypoUnknownProvider(t *testing.T) {
	t.Parallel()

	env := requiredEnv()
	env["TRUSTREADYD_CONNECTOR_GITHBU_ENDPOINT_API_BASE"] = "https://api.github.com"

	resolver := NewResolver(mockEnv(env))
	resolver.SetEnumerator(mockEnumerator(env))

	b := NewBuilder(resolver)
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	_, err := b.Build()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TRUSTREADYD_CONNECTOR_GITHBU_ENDPOINT_API_BASE")
	assert.Contains(t, err.Error(), "GITHBU")
}

// TestBuilder_Build_ConnectorEndpointTypoUnknownField mirrors the provider-typo
// case for a misspelled field segment: GITHUB is a real provider, but BASEURL
// is not one of the five endpoint fields trustreadyd resolves.
func TestBuilder_Build_ConnectorEndpointTypoUnknownField(t *testing.T) {
	t.Parallel()

	env := requiredEnv()
	env["TRUSTREADYD_CONNECTOR_GITHUB_ENDPOINT_BASEURL"] = "https://api.github.com"

	resolver := NewResolver(mockEnv(env))
	resolver.SetEnumerator(mockEnumerator(env))

	b := NewBuilder(resolver)
	b.samlCertificate = "test-cert"
	b.samlPrivateKey = testSigningKeyPEM()

	_, err := b.Build()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TRUSTREADYD_CONNECTOR_GITHUB_ENDPOINT_BASEURL")
	assert.Contains(t, err.Error(), "BASEURL")
}
