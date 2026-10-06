// Copyright (c) 2026 TrustReady Inc <hello@trustready.io>.
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

import { PageHeader } from "@probo/ui";
import { useTranslation } from "react-i18next";

// The TrustReady Security audit service (run Azure scan -> findings -> PDF report).
// Embedded so the scan/report workflow lives alongside the GRC console.
const REPORT_APP_URL
  = "https://trustready-web.niceglacier-3b15d07e.uksouth.azurecontainerapps.io/";

export default function ReportCreationPage() {
  const { t } = useTranslation();

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("nav.reportCreation")}
        description={t("nav.reportCreationDescription")}
      />
      <iframe
        title={t("nav.reportCreation")}
        src={REPORT_APP_URL}
        style={{ width: "100%", height: "78vh", border: 0 }}
      />
    </div>
  );
}
