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

package trustready_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/DhruvWork/trustready-grc/pkg/gid"
	"github.com/DhruvWork/trustready-grc/pkg/iam/policy"
	"github.com/DhruvWork/trustready-grc/pkg/trustready"
)

func TestAuditorPolicy_ProcessingActivityPageReadAccess(t *testing.T) {
	t.Parallel()

	organizationID := gid.New(gid.NewTenantID(), 1)
	evaluator := policy.NewEvaluator()
	conditionContext := policy.ConditionContext{
		Principal: map[string]string{
			"organization_id": organizationID.String(),
		},
		Resource: map[string]string{
			"organization_id": organizationID.String(),
		},
	}

	tests := []struct {
		name   string
		action string
	}{
		{
			name:   "list processing activities",
			action: trustready.ActionProcessingActivityList,
		},
		{
			name:   "list data protection impact assessments",
			action: trustready.ActionDataProtectionImpactAssessmentList,
		},
		{
			name:   "list transfer impact assessments",
			action: trustready.ActionTransferImpactAssessmentList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := evaluator.Evaluate(
				policy.AuthorizationRequest{
					Principal:        organizationID,
					Resource:         organizationID,
					Action:           tt.action,
					ConditionContext: conditionContext,
				},
				[]*policy.Policy{trustready.AuditorPolicy},
			)

			assert.True(t, result.IsAllowed())
		})
	}
}

func TestAuditorPolicy_OrganizationContextReadAccess(t *testing.T) {
	t.Parallel()

	organizationID := gid.New(gid.NewTenantID(), 1)
	evaluator := policy.NewEvaluator()
	conditionContext := policy.ConditionContext{
		Principal: map[string]string{
			"organization_id": organizationID.String(),
		},
		Resource: map[string]string{
			"organization_id": organizationID.String(),
		},
	}

	result := evaluator.Evaluate(
		policy.AuthorizationRequest{
			Principal:        organizationID,
			Resource:         organizationID,
			Action:           trustready.ActionOrganizationContextGet,
			ConditionContext: conditionContext,
		},
		[]*policy.Policy{trustready.AuditorPolicy},
	)

	assert.True(t, result.IsAllowed())
}
