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

package mcp_v1

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
	mcpgenmcp "go.probo.inc/mcpgen/mcp"
	"github.com/DhruvWork/trustready-grc/pkg/accessreview"
	"github.com/DhruvWork/trustready-grc/pkg/baseurl"
	"github.com/DhruvWork/trustready-grc/pkg/certmanager"
	cloudaws "github.com/DhruvWork/trustready-grc/pkg/cloud/aws"
	cloudazure "github.com/DhruvWork/trustready-grc/pkg/cloud/azure"
	cloudgcp "github.com/DhruvWork/trustready-grc/pkg/cloud/gcp"
	"github.com/DhruvWork/trustready-grc/pkg/complianceportal/management"
	"github.com/DhruvWork/trustready-grc/pkg/cookiebanner"
	"github.com/DhruvWork/trustready-grc/pkg/filemanager"
	"github.com/DhruvWork/trustready-grc/pkg/iam"
	"github.com/DhruvWork/trustready-grc/pkg/identityfederation"
	"github.com/DhruvWork/trustready-grc/pkg/itam"
	"github.com/DhruvWork/trustready-grc/pkg/mailman"
	"github.com/DhruvWork/trustready-grc/pkg/trustready"
	"github.com/DhruvWork/trustready-grc/pkg/resourcealias"
	"github.com/DhruvWork/trustready-grc/pkg/riskmanagement"
	"github.com/DhruvWork/trustready-grc/pkg/server/api/authn"
	"github.com/DhruvWork/trustready-grc/pkg/server/api/mcp/mcputils"
	"github.com/DhruvWork/trustready-grc/pkg/server/api/mcp/v1/server"
	"github.com/DhruvWork/trustready-grc/pkg/task"
	"github.com/DhruvWork/trustready-grc/pkg/thirdparty"
)

func NewMux(
	logger *log.Logger,
	trustreadySvc *trustready.Service,
	managementSvc *management.Service,
	certManagerSvc *certmanager.Service,
	resourceAliasSvc *resourcealias.Service,
	thirdPartySvc *thirdparty.Service,
	iamSvc *iam.Service,
	accessReviewSvc *accessreview.Service,
	cookieBannerSvc *cookiebanner.Service,
	riskManagementSvc *riskmanagement.Service,
	itamSvc *itam.Service,
	taskSvc *task.Service,
	mailmanSvc *mailman.Service,
	tokenSecret string,
	fileManagerSvc *filemanager.Service,
	baseURL *baseurl.BaseURL,
	identityFederation *identityfederation.Issuer,
	awsConnectorInstall cloudaws.ConnectorInstallConfig,
	gcpConnectorInstall cloudgcp.ConnectorInstallConfig,
	azureConnectorInstall cloudazure.ConnectorInstallConfig,
) *chi.Mux {
	logger = logger.Named("mcp.v1")

	logger.Info("initializing MCP server")

	resolver := &Resolver{
		trustreadySvc:              trustreadySvc,
		management:            managementSvc,
		certManager:           certManagerSvc,
		resourceAlias:         resourceAliasSvc,
		thirdPartySvc:         thirdPartySvc,
		iamSvc:                iamSvc,
		accessReview:          accessReviewSvc,
		cookieBanner:          cookieBannerSvc,
		riskManagement:        riskManagementSvc,
		itamSvc:               itamSvc,
		task:                  taskSvc,
		mailman:               mailmanSvc,
		logger:                logger,
		fileManager:           fileManagerSvc,
		baseURL:               baseURL,
		identityFederation:    identityFederation,
		awsConnectorInstall:   awsConnectorInstall,
		gcpConnectorInstall:   gcpConnectorInstall,
		azureConnectorInstall: azureConnectorInstall,
	}

	mcpServer := server.New(resolver, mcpgenmcp.WithRecoverFunc(mcputils.NewRecoverFunc(logger)))

	mcpServer.AddReceivingMiddleware(mcputils.LoggingMiddleware(logger))
	mcpServer.AddReceivingMiddleware(mcputils.ListToolsCacheMiddleware(5 * time.Minute))

	getServer := func(r *http.Request) *mcp.Server { return mcpServer }
	eventStore := mcp.NewMemoryEventStore(nil)

	handler := mcp.NewStreamableHTTPHandler(
		getServer,
		&mcp.StreamableHTTPOptions{
			Stateless: true,
			// SessionTimeout: 30 * time.Minute,
			EventStore: eventStore,
			Logger:     nil, // TODO put logger here
		},
	)

	r := chi.NewMux()
	r.Use(authn.NewAPIKeyMiddleware(iamSvc, tokenSecret))
	r.Use(authn.NewOAuth2AccessTokenMiddleware(iamSvc))
	r.Use(authn.NewIdentityPresenceMiddleware(baseURL))
	r.Handle("/", handler)

	logger.Info("MCP server initialized successfully")

	return r
}

func UnwrapOmittable[T any](field mcpgenmcp.Omittable[T]) *T {
	if !field.IsSet() {
		return nil
	}

	value, _ := field.Value()

	return &value
}

func optionalIntSlice(values *[]any) (*[]int, error) {
	if values == nil {
		return nil, nil
	}

	ids := make([]int, 0, len(*values))
	for _, item := range *values {
		n, err := intFromAny(item)
		if err != nil {
			return nil, err
		}

		ids = append(ids, n)
	}

	return &ids, nil
}

func intFromAny(item any) (int, error) {
	switch n := item.(type) {
	case int:
		return n, nil
	case int32:
		return int(n), nil
	case int64:
		return int(n), nil
	case float64:
		if n != math.Trunc(n) {
			return 0, fmt.Errorf("tcf_purpose_ids must contain integers")
		}

		return int(n), nil
	default:
		return 0, fmt.Errorf("tcf_purpose_ids must contain integers")
	}
}

func optionalPtr[T any](value *T) **T {
	if value == nil {
		return nil
	}

	return &value
}
