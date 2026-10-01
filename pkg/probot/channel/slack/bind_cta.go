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

package slack

const (
	SlashCommandName = "/probot"

	bindRequiredText = "I can't help until your TrustReady account is linked. " +
		"Run `/probot login` — I'll send you a private link."

	bindSlashAlreadyLinkedText = "Your TrustReady account is already linked."
	bindSlashLinkedText        = "Your TrustReady account is linked."
	bindSlashUsageText         = "Usage: `/probot login`"
	bindSlashUnavailableText   = "Probot is not available in this workspace."
	bindSlashFallbackText      = "Link your TrustReady account to use the TrustReady Slack assistant."
	bindSlashFailedText        = "I couldn't create a link right now. Try `/probot login` again."

	interactiveForbiddenText = "You don't have permission to do that."
	interactiveFailedText    = "I couldn't complete that action."
)

func bindRequiredBlocks(bindURL string) []any {
	return []any{
		map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": "Link your *TrustReady* account before I can help you here. " +
					"You'll confirm the connection while signed in to TrustReady.",
			},
		},
		map[string]any{
			"type": "actions",
			"elements": []any{
				map[string]any{
					"type": "button",
					"text": map[string]any{
						"type": "plain_text",
						"text": "Link TrustReady account",
					},
					"url":   bindURL,
					"style": "primary",
				},
			},
		},
	}
}
