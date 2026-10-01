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

import type { Category } from "../types";
import { TrustReadyElement } from "./base";
import type { TrustReadyRootElement } from "./base";
import type { TrustReadyCookieBannerRoot } from "./cookie-banner-root";

export class TrustReadyCategoryList extends TrustReadyElement {
  private root: TrustReadyRootElement | null = null;
  private template: HTMLTemplateElement | null = null;
  private onReady = (e: Event): void => {
    const { config } = (e as CustomEvent).detail;
    this.stamp(config.categories);
  };

  connectedCallback(): void {
    this.template = this.querySelector("template");
    if (!this.template) {
      this.warn("<trustready-category-list> requires a <template> child");
      return;
    }

    this.root = this.findAncestor<TrustReadyCookieBannerRoot>("trustready-cookie-banner-root");
    if (!this.root) return;

    this.validateTemplate();

    try {
      const config = this.root.bannerConfig;
      this.stamp(config.categories);
    } catch {
      this.root.addEventListener("trustready-ready", this.onReady, { once: true });
    }
  }

  disconnectedCallback(): void {
    if (this.root) {
      this.root.removeEventListener("trustready-ready", this.onReady);
    }
  }

  private stamp(categories: Category[]): void {
    if (!this.template) return;

    for (const cat of categories) {
      const wrapper = document.createElement("trustready-category");
      wrapper.setAttribute("name", cat.name);
      wrapper.setAttribute("slug", cat.slug);
      wrapper.setAttribute("kind", cat.kind);
      wrapper.setAttribute("description", cat.description);
      wrapper.setAttribute("cookies", JSON.stringify(cat.cookies));

      const clone = this.template.content.cloneNode(true) as DocumentFragment;
      this.fillSlots(clone, {
        name: cat.name,
        description: cat.description,
      });

      const hasCookies = cat.cookies && cat.cookies.length > 0;
      if (!hasCookies) {
        clone.querySelector("[data-action=toggle-cookies]")?.remove();
        clone.querySelector("trustready-cookie-list")?.remove();
      }

      wrapper.appendChild(clone);
      this.appendChild(wrapper);
    }
  }

  private validateTemplate(): void {
    if (!this.template) return;
    const content = this.template.content;
    const missing: string[] = [];
    if (!content.querySelector("trustready-category-toggle")) {
      missing.push("trustready-category-toggle");
    }
    if (!content.querySelector("trustready-cookie-list")) {
      missing.push("trustready-cookie-list");
    }
    if (missing.length > 0) {
      this.warn(`<trustready-category-list> template is missing required elements: ${missing.join(", ")}`);
      this.emitValidation(missing);
    }
  }

  private fillSlots(
    fragment: DocumentFragment,
    data: Record<string, string>,
  ): void {
    for (const [key, value] of Object.entries(data)) {
      const els = fragment.querySelectorAll(`[data-slot="${key}"]`);
      for (const el of els) {
        el.textContent = value;
      }
    }
  }
}
