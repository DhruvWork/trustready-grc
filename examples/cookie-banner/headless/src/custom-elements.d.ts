import "react";

type CE<T = object> = React.DetailedHTMLProps<
  React.HTMLAttributes<HTMLElement> & T,
  HTMLElement
>;

declare module "react" {
  namespace JSX {
    interface IntrinsicElements {
      "trustready-cookie-banner-root": CE<{
        "banner-id"?: string;
        "base-url"?: string;
        lang?: string;
        "gcm-enabled"?: string;
      }>;
      "trustready-banner": CE;
      "trustready-preference-panel": CE;
      "trustready-category-list": CE;
      "trustready-category-toggle": CE;
      "trustready-cookie-list": CE;
      "trustready-accept-button": CE;
      "trustready-reject-button": CE;
      "trustready-customize-button": CE;
      "trustready-acknowledge-button": CE;
      "trustready-save-button": CE;
      "trustready-privacy-choices": CE;
      "trustready-settings-link": CE;
    }
  }
}
