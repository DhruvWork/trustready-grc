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

package task

import (
	"github.com/spf13/cobra"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/cmdutil"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/task/activity"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/task/comment"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/task/create"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/task/delete"
	linklinear "github.com/DhruvWork/trustready-grc/pkg/cmd/task/link-linear"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/task/list"
	listlinearissues "github.com/DhruvWork/trustready-grc/pkg/cmd/task/list-linear-issues"
	listlinearteams "github.com/DhruvWork/trustready-grc/pkg/cmd/task/list-linear-teams"
	publishlinear "github.com/DhruvWork/trustready-grc/pkg/cmd/task/publish-linear"
	unlinkexternal "github.com/DhruvWork/trustready-grc/pkg/cmd/task/unlink-external"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/task/update"
	"github.com/DhruvWork/trustready-grc/pkg/cmd/task/view"
)

func NewCmdTask(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task <command>",
		Short: "Manage tasks",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(view.NewCmdView(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(delete.NewCmdDelete(f))
	cmd.AddCommand(comment.NewCmdComment(f))
	cmd.AddCommand(activity.NewCmdActivity(f))
	cmd.AddCommand(listlinearteams.NewCmdListLinearTeams(f))
	cmd.AddCommand(listlinearissues.NewCmdListLinearIssues(f))
	cmd.AddCommand(publishlinear.NewCmdPublishLinear(f))
	cmd.AddCommand(linklinear.NewCmdLinkLinear(f))
	cmd.AddCommand(unlinkexternal.NewCmdUnlinkExternal(f))

	return cmd
}
