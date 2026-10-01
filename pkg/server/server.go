// Copyright (c) 2025-2026 TrustReady <hello@trustready.io>.
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

package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"github.com/DhruvWork/trustready-grc/apps/console"
	employeeportalstatics "github.com/DhruvWork/trustready-grc/apps/employee-portal"
	"github.com/DhruvWork/trustready-grc/pkg/accessreview"
	"github.com/DhruvWork/trustready-grc/pkg/agentexecution"
	"github.com/DhruvWork/trustready-grc/pkg/baseurl"
	"github.com/DhruvWork/trustready-grc/pkg/certmanager"
	cloudaws "github.com/DhruvWork/trustready-grc/pkg/cloud/aws"
	cloudazure "github.com/DhruvWork/trustready-grc/pkg/cloud/azure"
	cloudgcp "github.com/DhruvWork/trustready-grc/pkg/cloud/gcp"
	"github.com/DhruvWork/trustready-grc/pkg/complianceportal/management"
	"github.com/DhruvWork/trustready-grc/pkg/complianceportal/visitor"
	"github.com/DhruvWork/trustready-grc/pkg/connector"
	"github.com/DhruvWork/trustready-grc/pkg/connector/provider"
	"github.com/DhruvWork/trustready-grc/pkg/cookiebanner"
	"github.com/DhruvWork/trustready-grc/pkg/esign"
	"github.com/DhruvWork/trustready-grc/pkg/filemanager"
	"github.com/DhruvWork/trustready-grc/pkg/geoloc"
	"github.com/DhruvWork/trustready-grc/pkg/iam"
	"github.com/DhruvWork/trustready-grc/pkg/identityfederation"
	"github.com/DhruvWork/trustready-grc/pkg/itam"
	"github.com/DhruvWork/trustready-grc/pkg/mailman"
	"github.com/DhruvWork/trustready-grc/pkg/trustready"
	"github.com/DhruvWork/trustready-grc/pkg/probot"
	slackchannel "github.com/DhruvWork/trustready-grc/pkg/probot/channel/slack"
	"github.com/DhruvWork/trustready-grc/pkg/probot/identitybinding"
	"github.com/DhruvWork/trustready-grc/pkg/resourcealias"
	"github.com/DhruvWork/trustready-grc/pkg/riskmanagement"
	"github.com/DhruvWork/trustready-grc/pkg/securecookie"
	"github.com/DhruvWork/trustready-grc/pkg/server/api"
	connect_v1 "github.com/DhruvWork/trustready-grc/pkg/server/api/connect/v1"
	"github.com/DhruvWork/trustready-grc/pkg/server/gqlutils"
	server_identityfederation "github.com/DhruvWork/trustready-grc/pkg/server/identityfederation"
	"github.com/DhruvWork/trustready-grc/pkg/server/mailactions"
	console_web "github.com/DhruvWork/trustready-grc/pkg/server/web"
	employeeportal_web "github.com/DhruvWork/trustready-grc/pkg/server/web/employeeportal"
	"github.com/DhruvWork/trustready-grc/pkg/slack"
	"github.com/DhruvWork/trustready-grc/pkg/task"
	"github.com/DhruvWork/trustready-grc/pkg/thirdparty"
	"github.com/DhruvWork/trustready-grc/pkg/uri"
)

type Config struct {
	BaseURL                 *baseurl.BaseURL
	FileStorageOrigin       string
	AllowedOrigins          []string
	ExtraHeaderFields       map[string]string
	TrustReady                   *trustready.Service
	ResourceAlias           *resourcealias.Service
	File                    *filemanager.Service
	IAM                     *iam.Service
	Visitor                 *visitor.Service
	ESign                   *esign.Service
	Management              *management.Service
	CertManager             *certmanager.Service
	AccessReview            *accessreview.Service
	AgentExecution          *agentexecution.Service
	Slack                   *slack.Service
	BotDeliveryDestinations api.BotDeliveryDestinations
	ComplianceMessages      api.ComplianceMessages
	Slackbot                *slackchannel.Service
	SlackInteractiveInbox   *slackchannel.InteractiveCommandInbox
	ProbotIdentityBindings  *identitybinding.Service
	SlackbotInstallations   *slackchannel.InstallationService
	ProbotCapabilities      *probot.CapabilityRegistry
	Mailman                 *mailman.Service
	CookieBanner            *cookiebanner.Service
	Geoloc                  *geoloc.Service
	ThirdParty              *thirdparty.Service
	RiskManagement          *riskmanagement.Service
	ITAM                    *itam.Service
	Task                    *task.Service
	Cookie                  securecookie.Config
	TokenSecret             string
	// InstallStateKey signs the app-install state tokens the connector install
	// ceremony carries across the vendor.
	InstallStateKey   string
	ConnectorRegistry *connector.Registry
	ProviderRegistry  *provider.Registry
	CustomDomainCname string
	GraphQLLimits     gqlutils.Limits
	Logger            *log.Logger

	// IdentityFederationIssuer serves the outbound OIDC documents. It is nil when the
	// identity federation issuer is disabled, in which case no /federation route exists.
	IdentityFederationIssuer *identityfederation.Issuer
	AWSConnectorInstall      cloudaws.ConnectorInstallConfig
	GCPConnectorInstall      cloudgcp.ConnectorInstallConfig
	AzureConnectorInstall    cloudazure.ConnectorInstallConfig
	LinearWebhookSecret      string
}

type Server struct {
	cfg                          Config
	apiServer                    *api.Server
	mailActionsHandler           http.Handler
	identityFederationHandler    http.Handler
	consoleWebServer             *console_web.Server
	consoleSecurityPolicy        string
	employeePortalWebServer      *employeeportal_web.Server
	employeePortalSecurityPolicy string
	router                       *chi.Mux
	extraHeaderFields            map[string]string
	baseURL                      string
	trustreadyService                 *trustready.Service
	iamService                   *iam.Service
	logger                       *log.Logger
}

func NewServer(cfg Config) (*Server, error) {
	apiCfg := api.Config{
		BaseURL:                  cfg.BaseURL,
		AllowedOrigins:           cfg.AllowedOrigins,
		TrustReady:                    cfg.TrustReady,
		ResourceAlias:            cfg.ResourceAlias,
		File:                     cfg.File,
		IAM:                      cfg.IAM,
		Visitor:                  cfg.Visitor,
		ESign:                    cfg.ESign,
		Management:               cfg.Management,
		CertManager:              cfg.CertManager,
		AccessReview:             cfg.AccessReview,
		AgentExecution:           cfg.AgentExecution,
		Slack:                    cfg.Slack,
		BotDeliveryDestinations:  cfg.BotDeliveryDestinations,
		ComplianceMessages:       cfg.ComplianceMessages,
		Slackbot:                 cfg.Slackbot,
		SlackInteractiveInbox:    cfg.SlackInteractiveInbox,
		ProbotIdentityBindings:   cfg.ProbotIdentityBindings,
		SlackbotInstallations:    cfg.SlackbotInstallations,
		ProbotCapabilities:       cfg.ProbotCapabilities,
		Mailman:                  cfg.Mailman,
		CookieBanner:             cfg.CookieBanner,
		Geoloc:                   cfg.Geoloc,
		ThirdParty:               cfg.ThirdParty,
		RiskManagement:           cfg.RiskManagement,
		ITAM:                     cfg.ITAM,
		Task:                     cfg.Task,
		Cookie:                   cfg.Cookie,
		TokenSecret:              cfg.TokenSecret,
		InstallStateKey:          cfg.InstallStateKey,
		ConnectorRegistry:        cfg.ConnectorRegistry,
		ProviderRegistry:         cfg.ProviderRegistry,
		CustomDomainCname:        cfg.CustomDomainCname,
		GraphQLLimits:            cfg.GraphQLLimits,
		Logger:                   cfg.Logger.Named("api"),
		IdentityFederationIssuer: cfg.IdentityFederationIssuer,
		AWSConnectorInstall:      cfg.AWSConnectorInstall,
		GCPConnectorInstall:      cfg.GCPConnectorInstall,
		AzureConnectorInstall:    cfg.AzureConnectorInstall,
		LinearWebhookSecret:      cfg.LinearWebhookSecret,
	}

	apiServer, err := api.NewServer(apiCfg)
	if err != nil {
		return nil, err
	}

	consoleWebServer, err := console_web.NewServer()
	if err != nil {
		return nil, err
	}

	employeePortalWebServer, err := employeeportal_web.NewServer()
	if err != nil {
		return nil, err
	}

	appOrigin := ""

	if cfg.BaseURL != nil {
		var originErr error

		appOrigin, originErr = cfg.BaseURL.CSPOrigin()
		if originErr != nil {
			return nil, fmt.Errorf("cannot build console content security policy: %w", originErr)
		}
	}

	consoleCSP, err := console.ContentSecurityPolicy(appOrigin, cfg.FileStorageOrigin)
	if err != nil {
		return nil, fmt.Errorf("cannot build console content security policy: %w", err)
	}

	employeePortalCSP, err := employeeportalstatics.ContentSecurityPolicy(appOrigin, cfg.FileStorageOrigin)
	if err != nil {
		return nil, fmt.Errorf("cannot build employee portal content security policy: %w", err)
	}

	router := chi.NewRouter()

	var identityFederationHandler http.Handler

	if cfg.IdentityFederationIssuer != nil {
		if cfg.TrustReady == nil || cfg.TrustReady.Organizations == nil {
			return nil, fmt.Errorf("cannot create server: identity federation issuer needs an organization service")
		}

		identityFederationHandler = server_identityfederation.NewMux(
			cfg.Logger.Named("identityfederation"),
			cfg.IdentityFederationIssuer,
			cfg.TrustReady.Organizations,
		)
	}

	server := &Server{
		cfg:                          cfg,
		apiServer:                    apiServer,
		mailActionsHandler:           mailactions.NewMux(cfg.Mailman, cfg.TokenSecret),
		identityFederationHandler:    identityFederationHandler,
		consoleWebServer:             consoleWebServer,
		consoleSecurityPolicy:        consoleCSP,
		employeePortalWebServer:      employeePortalWebServer,
		employeePortalSecurityPolicy: employeePortalCSP,
		router:                       router,
		extraHeaderFields:            cfg.ExtraHeaderFields,
		baseURL:                      cfg.BaseURL.String(),
		trustreadyService:                 cfg.TrustReady,
		iamService:                   cfg.IAM,
		logger:                       cfg.Logger,
	}

	server.setupRoutes()

	return server, nil
}

func (s *Server) setupRoutes() {
	// OIDC Discovery 1.0 §4 and RFC 8414 §3 both require the metadata
	// document at the issuer root under well-known paths.
	s.router.Get("/.well-known/openid-configuration", s.oidcDiscoveryHandler)
	s.router.Get("/.well-known/oauth-authorization-server", s.oidcDiscoveryHandler)
	s.router.Get("/.well-known/oauth-protected-resource", s.protectedResourceMetadataHandler)
	s.router.Get(
		"/.well-known/oauth-protected-resource/api/mcp/v1",
		s.mcpProtectedResourceMetadataHandler,
	)

	s.router.Mount("/api", http.StripPrefix("/api", s.apiServer))
	s.router.Mount("/mail-actions", http.StripPrefix("/mail-actions", s.mailActionsHandler))

	// The identity federation route tree is mounted at the same prefix in every
	// deployment; only the advertised issuer differs, and it comes from
	// configuration. The SaaS edge maps its public apex onto this prefix.
	if s.identityFederationHandler != nil {
		s.router.Mount(
			identityfederation.PathPrefix,
			http.StripPrefix(identityfederation.PathPrefix, s.identityFederationHandler),
		)
	}

	s.router.Mount(
		employeeportal_web.PathPrefix,
		NewSecurityHeadersMiddleware(
			SecurityHeadersOptions{
				ExtraHeaderFields:     s.extraHeaderFields,
				ContentSecurityPolicy: s.employeePortalSecurityPolicy,
			},
		)(http.StripPrefix(employeeportal_web.PathPrefix, s.employeePortalWebServer)),
	)

	s.router.Mount(
		"/",
		NewSecurityHeadersMiddleware(
			SecurityHeadersOptions{
				ExtraHeaderFields:     s.extraHeaderFields,
				ContentSecurityPolicy: s.consoleSecurityPolicy,
			},
		)(employeeportal_web.LegacyRedirectMiddleware(s.consoleWebServer)),
	)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.setExtraHeaders(w)
	s.router.ServeHTTP(w, r)
}

func (s *Server) setExtraHeaders(w http.ResponseWriter) {
	ApplyExtraHeaders(w, s.extraHeaderFields)
}

func (s *Server) oidcDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	metadata := connect_v1.OAuth2ServerMetadata(
		s.cfg.BaseURL,
		s.iamService.OAuth2ScopeRegistry.RegisteredScopes(),
	)

	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpserver.RenderJSON(w, http.StatusOK, metadata)
}

func (s *Server) protectedResourceMetadataHandler(w http.ResponseWriter, r *http.Request) {
	s.renderProtectedResourceMetadata(w, uri.URI(s.baseURL))
}

func (s *Server) mcpProtectedResourceMetadataHandler(w http.ResponseWriter, r *http.Request) {
	resource, err := s.iamService.OAuth2MCPResource()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.renderProtectedResourceMetadata(w, resource)
}

func (s *Server) renderProtectedResourceMetadata(w http.ResponseWriter, resource uri.URI) {
	metadata := s.iamService.OAuth2ProtectedResourceMetadata(resource)

	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpserver.RenderJSON(w, http.StatusOK, metadata)
}
