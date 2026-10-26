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

import { Button, Card, Field, Label, PageHeader, Spinner } from "@probo/ui";
import { type ChangeEvent, type FormEvent, useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

// Native Cloud Scanner. The GRC backend proxies these calls to the TrustReady
// Security API with a service credential, so the signed-in user needs no
// separate login. Endpoints: /frameworks, /audits, /scans, /audits/{id}/report.
const API = "/api/cloud-scanner/v1";

type Framework = { name: string };

type Audit = {
  audit_id: string;
  environment?: string | null;
  status?: string | null;
  conclusion?: string | null;
  started_at?: string | null;
  completed_at?: string | null;
  total_findings?: number | null;
  alarm_findings?: number | null;
  report_profile_label?: string | null;
  report_available?: boolean | null;
};

const selectClass = "w-full py-[6px] px-3 h-[34px] bg-secondary border border-border-mid "
  + "hover:border-border-strong rounded-[10px] text-sm text-txt-primary outline-none";

function statusTone(audit: Audit): string {
  const status = (audit.status ?? "").toLowerCase();
  if (status.includes("fail") || status === "error" || status === "cancelled") {
    return "text-txt-danger";
  }
  if (status.includes("success") || status.includes("complet")) {
    return "text-txt-success";
  }
  return "text-txt-tertiary";
}

export default function ReportCreationPage() {
  const { t } = useTranslation();

  const [frameworks, setFrameworks] = useState<Framework[]>([]);
  const [audits, setAudits] = useState<Audit[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [subscriptionId, setSubscriptionId] = useState("");
  const [environment, setEnvironment] = useState<"dev" | "production">("dev");
  const [framework, setFramework] = useState("");
  const [audience, setAudience] = useState<"dev" | "manager">("manager");

  const load = useCallback(async () => {
    try {
      const [frameworksRes, auditsRes] = await Promise.all([
        fetch(`${API}/frameworks`, { credentials: "include" }),
        fetch(`${API}/audits`, { credentials: "include" }),
      ]);
      if (frameworksRes.ok) {
        const body = (await frameworksRes.json()) as { frameworks?: Framework[] } | Framework[];
        const list = Array.isArray(body) ? body : (body.frameworks ?? []);
        setFrameworks(list);
        if (!framework && list.length) {
          setFramework(list[0].name);
        }
      }
      if (auditsRes.ok) {
        const body = (await auditsRes.json()) as { audits?: Audit[] } | Audit[];
        setAudits(Array.isArray(body) ? body : (body.audits ?? []));
      }
    }
    catch {
      // Non-fatal: keep whatever we already have on screen.
    }
    finally {
      setLoading(false);
    }
  }, [framework]);

  useEffect(() => {
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      const res = await fetch(`${API}/scans`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          subscription_ids: [subscriptionId.trim()],
          environment,
          report_audience: audience,
          framework: framework || null,
        }),
      });
      if (!res.ok) {
        const body = (await res.json().catch(() => null)) as { detail?: unknown; error?: string } | null;
        const detail = body?.error
          ?? (typeof body?.detail === "string" ? body.detail : null)
          ?? `Scan could not be started (${res.status}).`;
        throw new Error(detail);
      }
      await load();
    }
    catch (err) {
      setError(err instanceof Error ? err.message : "Scan could not be started.");
    }
    finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("nav.reportCreation")}
        description={t("nav.reportCreationDescription")}
      />

      <Card padded>
        <div className="mb-4">
          <h2 className="text-base font-semibold text-txt-primary">Run a scan</h2>
          <p className="text-sm text-txt-tertiary">
            Scan an Azure subscription, map findings to controls, and produce a report.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Field
              label="Azure subscription ID"
              name="subscriptionId"
              placeholder="00000000-0000-0000-0000-000000000000"
              value={subscriptionId}
              onChange={(event: ChangeEvent<HTMLInputElement>) => setSubscriptionId(event.target.value)}
              required
            />

            <div className="flex flex-col">
              <Label htmlFor="environment" className="mb-[6px]">Environment</Label>
              <select
                id="environment"
                className={selectClass}
                value={environment}
                onChange={event => setEnvironment(event.target.value as "dev" | "production")}
              >
                <option value="dev">Development</option>
                <option value="production">Production</option>
              </select>
            </div>

            <div className="flex flex-col">
              <Label htmlFor="framework" className="mb-[6px]">Framework</Label>
              <select
                id="framework"
                className={selectClass}
                value={framework}
                onChange={event => setFramework(event.target.value)}
              >
                {frameworks.length === 0 && <option value="">No frameworks configured</option>}
                {frameworks.map(item => (
                  <option key={item.name} value={item.name}>{item.name}</option>
                ))}
              </select>
            </div>

            <div className="flex flex-col">
              <Label htmlFor="audience" className="mb-[6px]">Report audience</Label>
              <select
                id="audience"
                className={selectClass}
                value={audience}
                onChange={event => setAudience(event.target.value as "dev" | "manager")}
              >
                <option value="manager">Manager report</option>
                <option value="dev">Developer summary</option>
              </select>
            </div>
          </div>

          {error && <p className="text-sm text-txt-danger">{error}</p>}

          <div className="flex items-center gap-3">
            <Button type="submit" disabled={submitting || !subscriptionId.trim()}>
              {submitting ? <Spinner size={16} centered /> : "Run scan"}
            </Button>
            <span className="text-xs text-txt-tertiary">
              Scans run in the background; refresh to see new results.
            </span>
          </div>
        </form>
      </Card>

      <Card>
        <div className="flex items-center justify-between border-b border-border-low px-4 py-3">
          <h2 className="text-base font-semibold text-txt-primary">Recent scans</h2>
          <Button variant="tertiary" onClick={() => void load()} disabled={loading}>
            Refresh
          </Button>
        </div>

        {loading
          ? (
              <div className="flex justify-center py-10">
                <Spinner size={24} />
              </div>
            )
          : audits.length === 0
              ? (
                  <p className="px-4 py-10 text-center text-sm text-txt-tertiary">
                    No scans yet. Start one above.
                  </p>
                )
              : (
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead className="text-left text-xs uppercase tracking-wide text-txt-tertiary">
                        <tr className="border-b border-border-low">
                          <th className="px-4 py-2 font-medium">Scan</th>
                          <th className="px-4 py-2 font-medium">Environment</th>
                          <th className="px-4 py-2 font-medium">Status</th>
                          <th className="px-4 py-2 font-medium">Findings</th>
                          <th className="px-4 py-2 font-medium">Started</th>
                          <th className="px-4 py-2 font-medium">Report</th>
                        </tr>
                      </thead>
                      <tbody>
                        {audits.map(audit => (
                          <tr key={audit.audit_id} className="border-b border-border-low last:border-0">
                            <td className="px-4 py-3 font-mono text-xs text-txt-secondary">
                              {audit.audit_id.slice(0, 8)}
                            </td>
                            <td className="px-4 py-3 text-txt-primary">
                              {audit.environment ?? "—"}
                            </td>
                            <td className={`px-4 py-3 font-medium ${statusTone(audit)}`}>
                              {audit.status ?? "—"}
                            </td>
                            <td className="px-4 py-3 text-txt-secondary">
                              {audit.alarm_findings ?? 0}
                              {" alarm / "}
                              {audit.total_findings ?? 0}
                              {" total"}
                            </td>
                            <td className="px-4 py-3 text-txt-tertiary">
                              {audit.started_at ? new Date(audit.started_at).toLocaleString() : "—"}
                            </td>
                            <td className="px-4 py-3">
                              {audit.report_available
                                ? (
                                    <a
                                      className="text-txt-primary underline underline-offset-2"
                                      href={`${API}/audits/${audit.audit_id}/report`}
                                      target="_blank"
                                      rel="noreferrer"
                                    >
                                      Download
                                    </a>
                                  )
                                : (
                                    <span className="text-txt-tertiary">—</span>
                                  )}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
      </Card>
    </div>
  );
}
