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

package iam_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/DhruvWork/trustready-grc/pkg/accessreview"
	"github.com/DhruvWork/trustready-grc/pkg/agentexecution"
	"github.com/DhruvWork/trustready-grc/pkg/coredata"
	"github.com/DhruvWork/trustready-grc/pkg/iam"
	"github.com/DhruvWork/trustready-grc/pkg/iam/oauth2scope"
	"github.com/DhruvWork/trustready-grc/pkg/itam"
	"github.com/DhruvWork/trustready-grc/pkg/trustready"
	"github.com/DhruvWork/trustready-grc/pkg/riskmanagement"
	"github.com/DhruvWork/trustready-grc/pkg/task"
)

func allRegisteredOAuth2ScopeRegistries() *oauth2scope.Registry {
	return oauth2scope.NewRegistry().
		Register(iam.IAMOAuth2ScopeMappings).
		Register(trustready.OAuth2ScopeMappings).
		Register(accessreview.OAuth2ScopeMappings).
		Register(agentexecution.OAuth2ScopeMappings).
		Register(itam.OAuth2ScopeMappings).
		Register(riskmanagement.OAuth2ScopeMappings).
		Register(task.OAuth2ScopeMappings)
}

func TestRegisteredOAuth2ScopeRegistries_OrganizationRead(t *testing.T) {
	t.Parallel()

	reg := allRegisteredOAuth2ScopeRegistries()
	tokenScopes := coredata.OAuth2Scopes{trustready.ScopeV1OrgRead}

	assert.True(t, reg.Allows(tokenScopes, trustready.ActionOrganizationGet))
	assert.False(t, reg.Allows(tokenScopes, trustready.ActionOrganizationUpdate))
	assert.False(t, reg.Allows(tokenScopes, trustready.ActionThirdPartyList))
}

func TestRegisteredOAuth2ScopeRegistries_UnmappedActionDenies(t *testing.T) {
	t.Parallel()

	reg := allRegisteredOAuth2ScopeRegistries()
	tokenScopes := coredata.OAuth2Scopes{
		trustready.ScopeV1OrgRead,
		trustready.ScopeV1ThirdPartyRead,
	}

	assert.False(t, reg.Allows(tokenScopes, "core:unmapped:action"))
}

func TestRegisteredOAuth2ScopeRegistries_ITAMDeviceDelete(t *testing.T) {
	t.Parallel()

	reg := allRegisteredOAuth2ScopeRegistries()

	assert.True(t, reg.Allows(coredata.OAuth2Scopes{itam.ScopeV1ITAM}, itam.ActionDeviceDelete))
	assert.False(t, reg.Allows(coredata.OAuth2Scopes{itam.ScopeV1ITAMRead}, itam.ActionDeviceDelete))
	assert.True(t, reg.Allows(coredata.OAuth2Scopes{itam.ScopeV1ITAMRead}, itam.ActionDeviceGet))
}

func TestRegisteredOAuth2ScopeRegistries_EvidenceOnControlNotTask(t *testing.T) {
	t.Parallel()

	reg := allRegisteredOAuth2ScopeRegistries()

	assert.False(t, reg.Allows(coredata.OAuth2Scopes{trustready.ScopeV1TaskRead}, trustready.ActionEvidenceList))
	assert.False(t, reg.Allows(coredata.OAuth2Scopes{trustready.ScopeV1Task}, trustready.ActionEvidenceList))
	assert.False(t, reg.Allows(coredata.OAuth2Scopes{trustready.ScopeV1Task}, trustready.ActionEvidenceDelete))

	assert.True(t, reg.Allows(coredata.OAuth2Scopes{trustready.ScopeV1ControlRead}, trustready.ActionEvidenceList))
	assert.False(t, reg.Allows(coredata.OAuth2Scopes{trustready.ScopeV1ControlRead}, trustready.ActionEvidenceDelete))
	assert.True(t, reg.Allows(coredata.OAuth2Scopes{trustready.ScopeV1Control}, trustready.ActionEvidenceList))
	assert.True(t, reg.Allows(coredata.OAuth2Scopes{trustready.ScopeV1Control}, trustready.ActionEvidenceDelete))
	assert.True(t, reg.Allows(coredata.OAuth2Scopes{trustready.ScopeV1Control}, trustready.ActionMeasureEvidenceUpload))
}
