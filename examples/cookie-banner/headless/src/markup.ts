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

function escapeAttr(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

export function headlessRootHTML(
  bannerId: string,
  baseUrl: string,
  gcmEnabled: boolean,
): string {
  return `
      <style>trustready-banner, trustready-preference-panel, trustready-privacy-choices { display: block !important; }</style>
      <trustready-cookie-banner-root banner-id="${escapeAttr(bannerId)}" base-url="${escapeAttr(baseUrl)}" gcm-enabled="${gcmEnabled ? "true" : "false"}">
        <trustready-banner>
          <div style="border:2px solid #333;padding:12px;margin-bottom:8px;">
            <strong>[trustready-banner]</strong>
            <div style="margin-top:8px;">
              <trustready-acknowledge-button><button>Acknowledge</button></trustready-acknowledge-button>
              <trustready-accept-button><button style="margin-left:8px;">Accept All</button></trustready-accept-button>
              <trustready-reject-button><button style="margin-left:8px;">Reject All</button></trustready-reject-button>
              <trustready-customize-button><button style="margin-left:8px;">Customize</button></trustready-customize-button>
            </div>
          </div>
        </trustready-banner>

        <trustready-preference-panel>
          <div style="border:2px dashed #666;padding:12px;margin-bottom:8px;">
            <strong>[trustready-preference-panel]</strong>
            <trustready-category-list>
              <template>
                <div style="border:1px solid #aaa;padding:8px;margin:4px 0;">
                  <span data-slot="name" style="font-weight:bold;"></span>:
                  <span data-slot="description"></span>
                  <trustready-category-toggle>
                    <label style="margin-left:8px;"><input type="checkbox" /> toggle</label>
                  </trustready-category-toggle>
                  <trustready-cookie-list hidden>
                    <template>
                      <div style="padding:4px 0 4px 16px;font-size:13px;">
                        <span data-slot="name" style="font-weight:bold;"></span>
                        &mdash; <span data-slot="description"></span>
                      </div>
                    </template>
                  </trustready-cookie-list>
                </div>
              </template>
            </trustready-category-list>
            <div style="margin-top:8px;">
              <trustready-accept-button><button>Accept All</button></trustready-accept-button>
              <trustready-reject-button><button style="margin-left:8px;">Reject All</button></trustready-reject-button>
              <trustready-save-button><button style="margin-left:8px;">Save Preferences</button></trustready-save-button>
            </div>
          </div>
        </trustready-preference-panel>

        <trustready-privacy-choices>
          <div style="border:2px solid #1d4ed8;padding:12px;margin-bottom:8px;">
            <strong>[trustready-privacy-choices]</strong>
            <p style="margin:8px 0;font-size:14px;">
              Right to opt out of sale/sharing and right to limit sensitive
              personal information (CCPA).
            </p>
            <trustready-reject-button>
              <button>Do Not Sell or Share My Personal Information</button>
            </trustready-reject-button>
          </div>
        </trustready-privacy-choices>
      </trustready-cookie-banner-root>
    `;
}
