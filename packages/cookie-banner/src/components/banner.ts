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

import { TrustReadyElement } from "./base";
import type { TrustReadyRootElement } from "./base";
import type { TrustReadyCookieBannerRoot } from "./cookie-banner-root";

export class TrustReadyBanner extends TrustReadyElement {
  private root: TrustReadyRootElement | null = null;
  private onStateChange = (e: Event): void => {
    const { state, prev } = (e as CustomEvent).detail;
    this.hidden = state !== "banner";
    if (state === "banner" && prev !== "loading") {
      this.focusFirst();
    }
  };

  private onReady = (): void => {
    this.validate();
  };

  connectedCallback(): void {
    this.hidden = true;
    this.root = this.findAncestor<TrustReadyCookieBannerRoot>("trustready-cookie-banner-root");

    if (this.root) {
      this.root.addEventListener("trustready-state", this.onStateChange);
      if (this.root.layout) {
        this.validate();
      } else {
        this.root.addEventListener("trustready-ready", this.onReady, { once: true });
      }
      if (this.root.state === "banner") {
        this.hidden = false;
      }
    }
  }

  disconnectedCallback(): void {
    if (this.root) {
      this.root.removeEventListener("trustready-state", this.onStateChange);
      this.root.removeEventListener("trustready-ready", this.onReady);
    }
  }

  private validate(): void {
    const layout = this.root?.layout;
    if (!layout) return;

    const missing: string[] = [];

    // The primary action is presentation-specific: notice banners acknowledge,
    // every other presentation accepts. Requiring the exact tag stops, e.g., an
    // opt-in banner from silently recording ACKNOWLEDGE, and gives the correct
    // missing-child diagnostic.
    const primaryTag =
      layout.presentation === "NOTICE"
        ? "trustready-acknowledge-button"
        : "trustready-accept-button";
    if (!this.querySelector(primaryTag)) {
      missing.push(primaryTag);
    }

    if (layout.buttons.reject_all && !this.querySelector("trustready-reject-button")) {
      missing.push("trustready-reject-button");
    }
    if (layout.buttons.customize && !this.querySelector("trustready-customize-button")) {
      missing.push("trustready-customize-button");
    }

    if (missing.length > 0) {
      this.warn(`<trustready-banner> is missing required children: ${missing.join(", ")}`);
      this.emitValidation(missing);
    }
  }
}
