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

package trustreadyd

import "github.com/DhruvWork/trustready-grc/pkg/trustreadydconfig"

type (
	FullConfig                         = trustreadydconfig.FullConfig
	Config                             = trustreadydconfig.Config
	UnitConfig                         = trustreadydconfig.UnitConfig
	MetricsConfig                      = trustreadydconfig.MetricsConfig
	TracingConfig                      = trustreadydconfig.TracingConfig
	ESignConfig                        = trustreadydconfig.ESignConfig
	CompliancePortalConfig             = trustreadydconfig.CompliancePortalConfig
	CompliancePortalTLSMode            = trustreadydconfig.CompliancePortalTLSMode
	APIConfig                          = trustreadydconfig.APIConfig
	CorsConfig                         = trustreadydconfig.CorsConfig
	GraphQLConfig                      = trustreadydconfig.GraphQLConfig
	ProxyProtocolConfig                = trustreadydconfig.ProxyProtocolConfig
	AuthConfig                         = trustreadydconfig.AuthConfig
	IdentityFederationConfig           = trustreadydconfig.IdentityFederationConfig
	IdentityFederationSigningKeyConfig = trustreadydconfig.IdentityFederationSigningKeyConfig
	OAuth2ServerConfig                 = trustreadydconfig.OAuth2ServerConfig
	OAuth2SigningKeyConfig             = trustreadydconfig.OAuth2SigningKeyConfig
	CookieConfig                       = trustreadydconfig.CookieConfig
	CookieSameSite                     = trustreadydconfig.CookieSameSite
	PasswordConfig                     = trustreadydconfig.PasswordConfig
	PrivateKey                         = trustreadydconfig.PrivateKey
	RSAPrivateKey                      = trustreadydconfig.RSAPrivateKey
	AWSConfig                          = trustreadydconfig.AWSConfig
	ConnectorConfig                    = trustreadydconfig.ConnectorConfig
	ConnectorConfigOAuth2              = trustreadydconfig.ConnectorConfigOAuth2
	ConnectorConfigGitHubApp           = trustreadydconfig.ConnectorConfigGitHubApp
	CustomDomainsConfig                = trustreadydconfig.CustomDomainsConfig
	ACMEConfig                         = trustreadydconfig.ACMEConfig
	LLMProviderConfig                  = trustreadydconfig.LLMProviderConfig
	LLMAgentConfig                     = trustreadydconfig.LLMAgentConfig
	EvidenceDescriberConfig            = trustreadydconfig.EvidenceDescriberConfig
	ThirdPartyVettingWorkerConfig      = trustreadydconfig.ThirdPartyVettingWorkerConfig
	AgentsConfig                       = trustreadydconfig.AgentsConfig

	TrackerMappingWorkerConfig             = trustreadydconfig.TrackerMappingWorkerConfig
	CommonPatternEnrichmentWorkerConfig    = trustreadydconfig.CommonPatternEnrichmentWorkerConfig
	CommonThirdPartyEnrichmentWorkerConfig = trustreadydconfig.CommonThirdPartyEnrichmentWorkerConfig

	MailerConfig        = trustreadydconfig.MailerConfig
	SMTPConfig          = trustreadydconfig.SMTPConfig
	NotificationsConfig = trustreadydconfig.NotificationsConfig
	WebhookConfig       = trustreadydconfig.WebhookConfig

	DocumentNotificationConfig = trustreadydconfig.DocumentNotificationConfig
	OIDCProviderConfig         = trustreadydconfig.OIDCProviderConfig
	PgConfig                   = trustreadydconfig.PgConfig
	SAMLConfig                 = trustreadydconfig.SAMLConfig
	SCIMBridgeConfig           = trustreadydconfig.SCIMBridgeConfig
	ITAMConfig                 = trustreadydconfig.ITAMConfig
	SlackConfig                = trustreadydconfig.SlackConfig
	SlackbotConfig             = trustreadydconfig.SlackbotConfig
	CookieBannerConfig         = trustreadydconfig.CookieBannerConfig
)

const (
	CompliancePortalTLSModeDirect   = trustreadydconfig.CompliancePortalTLSModeDirect
	CompliancePortalTLSModeExternal = trustreadydconfig.CompliancePortalTLSModeExternal
	CookieSameSiteLax               = trustreadydconfig.CookieSameSiteLax
	CookieSameSiteStrict            = trustreadydconfig.CookieSameSiteStrict
	CookieSameSiteNone              = trustreadydconfig.CookieSameSiteNone
)
