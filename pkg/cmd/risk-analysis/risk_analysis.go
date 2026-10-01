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

package riskanalysis

import (
	"github.com/spf13/cobra"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/cmdutil"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/boundary"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/create"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/delete"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/diagram"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/fork"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/list"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/node"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/process"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/publish"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/scenario"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/threat"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/update"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/risk-analysis/view"
)

func NewCmdRiskAnalysis(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "risk-analysis <command>",
		Short: "Manage risk analyses",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(fork.NewCmdFork(f))
	cmd.AddCommand(view.NewCmdView(f))
	cmd.AddCommand(publish.NewCmdPublish(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(delete.NewCmdDelete(f))
	cmd.AddCommand(diagram.NewCmdDiagram(f))
	cmd.AddCommand(node.NewCmdNode(f))
	cmd.AddCommand(boundary.NewCmdBoundary(f))
	cmd.AddCommand(process.NewCmdProcess(f))
	cmd.AddCommand(threat.NewCmdThreat(f))
	cmd.AddCommand(scenario.NewCmdScenario(f))

	return cmd
}
