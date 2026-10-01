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

const LEGACY_CONSOLE_HOSTS: Record<string, string> = {
  "eu.console.trustready.io": "eu.trustready.io",
  "us.console.trustready.io": "us.trustready.io",
};

// Snippets copied while the console lived on *.console.trustready.io still
// pass that host as baseUrl. Rewrite so API calls hit the current hosted
// domains; self-hosted and already-migrated URLs are left alone.
export function rewriteLegacyConsoleHost(url: URL): URL {
  const next = LEGACY_CONSOLE_HOSTS[url.hostname.toLowerCase()];
  if (!next) {
    return url;
  }
  const rewritten = new URL(url.href);
  rewritten.hostname = next;
  return rewritten;
}
