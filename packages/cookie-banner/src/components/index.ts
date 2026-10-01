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

export { TrustReadyElement } from "./base";
export type { TrustReadyState, TrustReadyRootElement, ConsentDraft } from "./base";
export { TrustReadyBanner } from "./banner";
export {
  TrustReadyAcceptButton,
  TrustReadyAcknowledgeButton,
  TrustReadyCustomizeButton,
  TrustReadyRejectButton,
} from "./buttons";
export { TrustReadyCategory } from "./category";
export { TrustReadyCategoryList } from "./category-list";
export { TrustReadyCategoryToggle } from "./category-toggle";
export { TrustReadyCookieBannerRoot } from "./cookie-banner-root";
export { TrustReadyCookie, TrustReadyCookieList } from "./cookie-list";
export { TrustReadyPreferencePanel, TrustReadySaveButton } from "./preference-panel";
export { TrustReadyPrivacyChoices } from "./privacy-choices";
export { TrustReadySettingsLink } from "./settings-link";
export { registerHeadlessComponents } from "./register";
