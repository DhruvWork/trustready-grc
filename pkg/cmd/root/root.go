// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package root

import (
	"github.com/spf13/cobra"
	accessreview "github.com/DhruvWork/trustready-grc/pkg/cmd/access-review"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/aisystem"
	cmdapi "github.com/DhruvWork/trustready-grc/pkg/cmd/api"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/asset"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/audit"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/auditlog"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/auth"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/browse"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/businessfunction"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/cmdutil"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/completion"
	complianceportal "github.com/DhruvWork/trustready-grc/pkg/cmd/compliance-portal"
	cmdconfig "github.com/DhruvWork/trustready-grc/pkg/cmd/config"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/connector"
	consentrecord "github.com/DhruvWork/trustready-grc/pkg/cmd/consent-record"
	cmdcontext "github.com/DhruvWork/trustready-grc/pkg/cmd/context"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/control"
	cookiebanner "github.com/DhruvWork/trustready-grc/pkg/cmd/cookie-banner"
	cookiecategory "github.com/DhruvWork/trustready-grc/pkg/cmd/cookie-category"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/datum"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/device"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/document"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/dpia"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/evidence"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/finding"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/framework"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/measure"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/obligation"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/org"
	processingactivity "github.com/DhruvWork/trustready-grc/pkg/cmd/processing-activity"
	resourcealias "github.com/DhruvWork/trustready-grc/pkg/cmd/resource-alias"
	rightsrequest "github.com/DhruvWork/trustready-grc/pkg/cmd/rights-request"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk"
	riskanalysis "github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/scim"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/soa"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/task"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/thirdpartymgmt"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/tia"
	trackerpattern "github.com/DhruvWork/trustready-grc/pkg/cmd/tracker-pattern"
	trackerresource "github.com/DhruvWork/trustready-grc/pkg/cmd/tracker-resource"
	treatmentplan "github.com/DhruvWork/trustready-grc/pkg/cmd/treatment-plan"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/user"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/version"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/webhook"
)

func NewCmdRoot(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "prb <command> [flags]",
		Short:         "Probo CLI",
		Long:          "prb is a command-line tool for interacting with the Probo platform.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if noInteractive, _ := cmd.Flags().GetBool("no-interactive"); noInteractive {
				f.IOStreams.ForceNonInteractive = true
			}

			if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
				f.IOStreams.ForceNoColor = true
			}

			f.IOStreams.ApplyColorProfile()
		},
	}

	cmd.PersistentFlags().Bool(
		"no-interactive",
		false,
		"Disable interactive prompts (also set via PROBO_NO_INTERACTIVE=1, CI=true, or TERM=dumb)",
	)

	cmd.PersistentFlags().Bool(
		"no-color",
		false,
		"Disable ANSI color output (also set via NO_COLOR or TERM=dumb)",
	)

	cmd.AddCommand(accessreview.NewCmdAccessReview(f))
	cmd.AddCommand(aisystem.NewCmdAiSystem(f))
	cmd.AddCommand(cmdapi.NewCmdAPI(f))
	cmd.AddCommand(asset.NewCmdAsset(f))
	cmd.AddCommand(audit.NewCmdAudit(f))
	cmd.AddCommand(auditlog.NewCmdAuditLog(f))
	cmd.AddCommand(auth.NewCmdAuth(f))
	cmd.AddCommand(browse.NewCmdBrowse(f))
	cmd.AddCommand(businessfunction.NewCmdBusinessFunction(f))
	cmd.AddCommand(completion.NewCmdCompletion(f))
	cmd.AddCommand(cmdconfig.NewCmdConfig(f))
	cmd.AddCommand(consentrecord.NewCmdConsentRecord(f))
	cmd.AddCommand(cmdcontext.NewCmdContext(f))
	cmd.AddCommand(connector.NewCmdConnector(f))
	cmd.AddCommand(control.NewCmdControl(f))
	cmd.AddCommand(cookiebanner.NewCmdCookieBanner(f))
	cmd.AddCommand(cookiecategory.NewCmdCookieCategory(f))
	cmd.AddCommand(trackerpattern.NewCmdTrackerPattern(f))
	cmd.AddCommand(trackerresource.NewCmdTrackerResource(f))
	cmd.AddCommand(datum.NewCmdDatum(f))
	cmd.AddCommand(device.NewCmdDevice(f))
	cmd.AddCommand(document.NewCmdDocument(f))
	cmd.AddCommand(dpia.NewCmdDPIA(f))
	cmd.AddCommand(evidence.NewCmdEvidence(f))
	cmd.AddCommand(finding.NewCmdFinding(f))
	cmd.AddCommand(framework.NewCmdFramework(f))
	cmd.AddCommand(measure.NewCmdMeasure(f))
	cmd.AddCommand(obligation.NewCmdObligation(f))
	cmd.AddCommand(org.NewCmdOrg(f))
	cmd.AddCommand(processingactivity.NewCmdProcessingActivity(f))
	cmd.AddCommand(rightsrequest.NewCmdRightsRequest(f))
	cmd.AddCommand(risk.NewCmdRisk(f))
	cmd.AddCommand(riskanalysis.NewCmdRiskAnalysis(f))
	cmd.AddCommand(resourcealias.NewCmdResourceAlias(f))
	cmd.AddCommand(scim.NewCmdScim(f))
	cmd.AddCommand(soa.NewCmdSoa(f))
	cmd.AddCommand(task.NewCmdTask(f))
	cmd.AddCommand(treatmentplan.NewCmdTreatmentPlan(f))
	cmd.AddCommand(tia.NewCmdTIA(f))
	cmd.AddCommand(complianceportal.NewCmdCompliancePortal(f))
	cmd.AddCommand(user.NewCmdUser(f))
	cmd.AddCommand(thirdpartymgmt.NewCmdThirdParty(f))
	cmd.AddCommand(version.NewCmdVersion(f))
	cmd.AddCommand(webhook.NewCmdWebhook(f))

	return cmd
}
