// Copyright (c) 2026 TrustReady <hello@probo.com>.
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

package complianceportal_v1

import (
	"net/http"

	"go.gearno.de/kit/log"
	"github.com/DhruvWork/trustready-grc/pkg/baseurl"
	"github.com/DhruvWork/trustready-grc/pkg/complianceportal/visitor"
	"github.com/DhruvWork/trustready-grc/pkg/esign"
	"github.com/DhruvWork/trustready-grc/pkg/filemanager"
	"github.com/DhruvWork/trustready-grc/pkg/iam"
	"github.com/DhruvWork/trustready-grc/pkg/mailman"
	"github.com/DhruvWork/trustready-grc/pkg/resourcealias"
	"github.com/DhruvWork/trustready-grc/pkg/securecookie"
	"github.com/DhruvWork/trustready-grc/pkg/server/api/authn"
	"github.com/DhruvWork/trustready-grc/pkg/server/api/complianceportal/v1/schema"
	"github.com/DhruvWork/trustready-grc/pkg/server/gqlutils"
	"github.com/DhruvWork/trustready-grc/pkg/server/gqlutils/directives/authentication"
	"github.com/DhruvWork/trustready-grc/pkg/server/gqlutils/directives/session"
)

func NewGraphQLHandler(
	iamSvc *iam.Service,
	visitorSvc *visitor.Service,
	resourceAliasSvc *resourcealias.Service,
	fileManagerSvc *filemanager.Service,
	esignSvc *esign.Service,
	mailmanSvc *mailman.Service,
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	cookieConfig securecookie.Config,
	tokenSecret string,
	limits gqlutils.Limits,
) http.Handler {
	config := schema.Config{
		Resolvers: &Resolver{
			iam:           iamSvc,
			visitor:       visitorSvc,
			resourceAlias: resourceAliasSvc,
			fileManager:   fileManagerSvc,
			esign:         esignSvc,
			mailman:       mailmanSvc,
			logger:        logger,
			baseURL:       baseURL,
			sessionCookie: authn.NewCookie(&cookieConfig),
		},
		Directives: schema.DirectiveRoot{
			Authentication: authentication.Directive,
			SessionOnly:    session.Directive,
		},
	}

	es := schema.NewExecutableSchema(config)
	gqlh := gqlutils.NewHandler(es, logger, limits)

	return gqlh
}
