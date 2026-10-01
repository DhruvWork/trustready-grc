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

import { TrustReadyBanner } from "./banner";
import {
  TrustReadyAcceptButton,
  TrustReadyAcknowledgeButton,
  TrustReadyCustomizeButton,
  TrustReadyRejectButton,
} from "./buttons";
import { TrustReadyCategory } from "./category";
import { TrustReadyCategoryList } from "./category-list";
import { TrustReadyCategoryToggle } from "./category-toggle";
import { TrustReadyCookieBannerRoot } from "./cookie-banner-root";
import { TrustReadyCookie, TrustReadyCookieList } from "./cookie-list";
import { TrustReadyPreferencePanel, TrustReadySaveButton } from "./preference-panel";
import { TrustReadyPrivacyChoices } from "./privacy-choices";
import { TrustReadySettingsLink } from "./settings-link";

const elements: [string, CustomElementConstructor][] = [
  ["trustready-cookie-banner-root", TrustReadyCookieBannerRoot],
  ["trustready-banner", TrustReadyBanner],
  ["trustready-accept-button", TrustReadyAcceptButton],
  ["trustready-acknowledge-button", TrustReadyAcknowledgeButton],
  ["trustready-reject-button", TrustReadyRejectButton],
  ["trustready-customize-button", TrustReadyCustomizeButton],
  ["trustready-preference-panel", TrustReadyPreferencePanel],
  ["trustready-privacy-choices", TrustReadyPrivacyChoices],
  ["trustready-category-list", TrustReadyCategoryList],
  ["trustready-category", TrustReadyCategory],
  ["trustready-category-toggle", TrustReadyCategoryToggle],
  ["trustready-cookie-list", TrustReadyCookieList],
  ["trustready-cookie", TrustReadyCookie],
  ["trustready-save-button", TrustReadySaveButton],
  ["trustready-settings-link", TrustReadySettingsLink],
];

export function registerHeadlessComponents(): void {
  for (const [name, ctor] of elements) {
    if (!customElements.get(name)) {
      customElements.define(name, ctor);
    }
  }
}
