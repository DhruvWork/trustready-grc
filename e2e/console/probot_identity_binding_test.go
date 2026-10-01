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

package console_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/DhruvWork/trustready-grc/e2e/internal/factory"
	"github.com/DhruvWork/trustready-grc/e2e/internal/testutil"
	"github.com/DhruvWork/trustready-grc/internal/test"
	"github.com/DhruvWork/trustready-grc/pkg/baseurl"
	"github.com/DhruvWork/trustready-grc/pkg/gid"
	slackchannel "github.com/DhruvWork/trustready-grc/pkg/probot/channel/slack"
	"github.com/DhruvWork/trustready-grc/pkg/probot/identitybinding"
)

func newTrustReadytBindToken(
	t *testing.T,
	subject identitybinding.Subject,
	organizationID gid.GID,
) string {
	t.Helper()

	baseURL, err := baseurl.Parse("https://console.example.com")
	require.NoError(t, err)
	service := identitybinding.NewService(test.PGClient(t), baseURL)
	bindURL, err := service.BindURL(context.Background(), subject, organizationID)
	require.NoError(t, err)
	parsed, err := url.Parse(bindURL)
	require.NoError(t, err)

	return parsed.Query().Get("token")
}

func TestTrustReadytIdentityBinding_ConfirmWhileLoggedIn(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	other := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

	externalTenantID := "T-e2e-" + factory.SafeName("team")
	externalUserID := "U-e2e-" + factory.SafeName("user")
	token := newTrustReadytBindToken(t, identitybinding.Subject{
		Provider:           slackchannel.ProviderName,
		ExternalTenantID:   externalTenantID,
		ExternalUserID:     externalUserID,
		ExternalTenantName: "acme-workspace",
		ExternalUserName:   "ada",
	}, owner.GetOrganizationID())

	const previewQuery = `
		query($token: String!) {
			trustreadytIdentityBindPreview(token: $token) {
				provider
				externalTenantId
				externalUserId
				externalTenantName
				externalUserName
			}
		}
	`

	var previewResult struct {
		TrustReadytIdentityBindPreview struct {
			Provider           string `json:"provider"`
			ExternalTenantID   string `json:"externalTenantId"`
			ExternalUserID     string `json:"externalUserId"`
			ExternalTenantName string `json:"externalTenantName"`
			ExternalUserName   string `json:"externalUserName"`
		} `json:"trustreadytIdentityBindPreview"`
	}

	err := owner.Execute(
		previewQuery,
		map[string]any{"token": token},
		&previewResult,
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		slackchannel.ProviderName,
		previewResult.TrustReadytIdentityBindPreview.Provider,
	)
	assert.Equal(
		t,
		externalTenantID,
		previewResult.TrustReadytIdentityBindPreview.ExternalTenantID,
	)
	assert.Equal(
		t,
		externalUserID,
		previewResult.TrustReadytIdentityBindPreview.ExternalUserID,
	)
	assert.Equal(
		t,
		"acme-workspace",
		previewResult.TrustReadytIdentityBindPreview.ExternalTenantName,
	)
	assert.Equal(
		t,
		"ada",
		previewResult.TrustReadytIdentityBindPreview.ExternalUserName,
	)

	const confirmMutation = `
		mutation($input: ConfirmTrustReadytIdentityBindingInput!) {
			confirmTrustReadytIdentityBinding(input: $input) {
				trustreadytIdentityBinding {
					id
					provider
					externalTenantId
					externalUserId
				}
				viewer {
					trustreadytIdentityBindings {
						id
						provider
						externalTenantId
						externalUserId
					}
				}
			}
		}
	`

	var confirmResult struct {
		ConfirmTrustReadytIdentityBinding struct {
			TrustReadytIdentityBinding struct {
				ID               string `json:"id"`
				Provider         string `json:"provider"`
				ExternalTenantID string `json:"externalTenantId"`
				ExternalUserID   string `json:"externalUserId"`
			} `json:"trustreadytIdentityBinding"`
			Viewer struct {
				TrustReadytIdentityBindings []struct {
					ID               string `json:"id"`
					Provider         string `json:"provider"`
					ExternalTenantID string `json:"externalTenantId"`
					ExternalUserID   string `json:"externalUserId"`
				} `json:"trustreadytIdentityBindings"`
			} `json:"viewer"`
		} `json:"confirmTrustReadytIdentityBinding"`
	}

	err = owner.Execute(
		confirmMutation,
		map[string]any{"input": map[string]any{"token": token}},
		&confirmResult,
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		externalTenantID,
		confirmResult.ConfirmTrustReadytIdentityBinding.TrustReadytIdentityBinding.ExternalTenantID,
	)
	assert.Equal(
		t,
		externalUserID,
		confirmResult.ConfirmTrustReadytIdentityBinding.TrustReadytIdentityBinding.ExternalUserID,
	)
	assert.NotEmpty(
		t,
		confirmResult.ConfirmTrustReadytIdentityBinding.TrustReadytIdentityBinding.ID,
	)
	require.Len(
		t,
		confirmResult.ConfirmTrustReadytIdentityBinding.Viewer.TrustReadytIdentityBindings,
		1,
	)
	assert.Equal(
		t,
		confirmResult.ConfirmTrustReadytIdentityBinding.TrustReadytIdentityBinding.ID,
		confirmResult.ConfirmTrustReadytIdentityBinding.Viewer.TrustReadytIdentityBindings[0].ID,
	)

	const deleteMutation = `
		mutation($input: DeleteTrustReadytIdentityBindingInput!) {
			deleteTrustReadytIdentityBinding(input: $input) {
				trustreadytIdentityBindingId
				viewer {
					trustreadytIdentityBindings {
						id
					}
				}
			}
		}
	`

	var deleteResult struct {
		DeleteTrustReadytIdentityBinding struct {
			TrustReadytIdentityBindingID string `json:"trustreadytIdentityBindingId"`
			Viewer                  struct {
				TrustReadytIdentityBindings []struct {
					ID string `json:"id"`
				} `json:"trustreadytIdentityBindings"`
			} `json:"viewer"`
		} `json:"deleteTrustReadytIdentityBinding"`
	}

	err = owner.Execute(
		deleteMutation,
		map[string]any{
			"input": map[string]any{
				"id": confirmResult.ConfirmTrustReadytIdentityBinding.TrustReadytIdentityBinding.ID,
			},
		},
		&deleteResult,
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		confirmResult.ConfirmTrustReadytIdentityBinding.TrustReadytIdentityBinding.ID,
		deleteResult.DeleteTrustReadytIdentityBinding.TrustReadytIdentityBindingID,
	)
	assert.Empty(t, deleteResult.DeleteTrustReadytIdentityBinding.Viewer.TrustReadytIdentityBindings)

	err = owner.Execute(
		previewQuery,
		map[string]any{"token": token},
		&previewResult,
	)
	testutil.RequireErrorCode(t, err, "INVALID")

	err = owner.Execute(
		confirmMutation,
		map[string]any{"input": map[string]any{"token": token}},
		&confirmResult,
	)
	testutil.RequireErrorCode(t, err, "CONFLICT")

	err = other.Execute(
		confirmMutation,
		map[string]any{"input": map[string]any{"token": token}},
		&confirmResult,
	)
	testutil.RequireErrorCode(t, err, "CONFLICT")
}
