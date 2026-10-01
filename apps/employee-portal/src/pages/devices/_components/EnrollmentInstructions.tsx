// Copyright (c) 2026 TrustReady <hello@probo.com>.
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

import { Callout } from "@trustready/ui/src/v2/Callout/Callout";
import { Anchor } from "@trustready/ui/src/v2/Link/Anchor";
import { Tabs } from "@trustready/ui/src/v2/Tabs/Tabs";
import { TabsIndicator } from "@trustready/ui/src/v2/Tabs/TabsIndicator";
import { TabsList } from "@trustready/ui/src/v2/Tabs/TabsList";
import { TabsTab } from "@trustready/ui/src/v2/Tabs/TabsTab";
import { Text } from "@trustready/ui/src/v2/typography/Text";
import { useState } from "react";
import { Trans, useTranslation } from "react-i18next";

import { CopyableCodeBlock } from "./CopyableCodeBlock";
import { enrollmentInstructions } from "./variants";

const AGENT_RELEASES_URL
  = "https://github.com/getprobo/probo/releases?q=trustready-agent";

const UNIX_DOWNLOAD_COMMAND = `curl -fsSL "https://github.com/getprobo/probo/releases/download/trustready-agent/vX.Y.Z/trustready-agent_OS_ARCH.tar.gz" -o /tmp/trustready-agent.tar.gz
tar -xzf /tmp/trustready-agent.tar.gz -C /tmp
sudo install -m 0755 /tmp/trustready-agent_OS_ARCH/trustready-agent /usr/local/bin/trustready-agent
rm -rf /tmp/trustready-agent.tar.gz /tmp/trustready-agent_OS_ARCH`;

const WINDOWS_DOWNLOAD_COMMAND = `$zip = "$env:TEMP\\trustready-agent.zip"
$dst = "$env:ProgramFiles\\Probo"
Invoke-WebRequest -Uri "https://github.com/getprobo/probo/releases/download/trustready-agent/vX.Y.Z/trustready-agent_Windows_ARCH.zip" -OutFile $zip
Expand-Archive -Path $zip -DestinationPath $env:TEMP -Force
New-Item -ItemType Directory -Force -Path $dst | Out-Null
Move-Item -Force "$env:TEMP\\trustready-agent_Windows_ARCH\\trustready-agent.exe" "$dst\\trustready-agent.exe"
Remove-Item -Recurse -Force $zip, "$env:TEMP\\trustready-agent_Windows_ARCH"`;

type InstallOs = "unix" | "windows";

export interface EnrollmentInstructionsProps {
  enrollmentToken: string;
  serverUrl: string;
}

export function EnrollmentInstructions({
  enrollmentToken,
  serverUrl,
}: EnrollmentInstructionsProps) {
  const { t } = useTranslation("devices");
  const slots = enrollmentInstructions();
  const [os, setOs] = useState<InstallOs>("unix");
  const downloadComment = `# ${t("addManually.token.downloadStep")}`;
  const enrollComment = `# ${t("addManually.token.enrollStep")}`;
  const unixDownloadCommand = `${downloadComment}
${UNIX_DOWNLOAD_COMMAND}`;
  const unixInstallCommand = `${enrollComment}
sudo /usr/local/bin/trustready-agent install \\
  --server ${serverUrl} \\
  --enrollment-token '${enrollmentToken}'`;
  const windowsDownloadCommand = `${downloadComment}
${WINDOWS_DOWNLOAD_COMMAND}`;
  const windowsInstallCommand = `${enrollComment}
& "$env:ProgramFiles\\Probo\\trustready-agent.exe" install \`
  --server ${serverUrl} \`
  --enrollment-token '${enrollmentToken}'`;

  function handleOsChange(value: string | number | null) {
    if (value === "unix" || value === "windows") {
      setOs(value);
    }
  }

  return (
    <div className={slots.root()}>
      <Callout size={2} variant="surface" color="sky" highContrast>
        <div className={slots.token()}>
          <Text size={2} weight="medium" color="current">
            {t("addManually.token.title")}
          </Text>
          <Text size={2} color="current">
            {t("addManually.token.description")}
          </Text>
          <CopyableCodeBlock code={enrollmentToken} />
        </div>
      </Callout>
      <div className={slots.install()}>
        <Text size={2} color="neutral">
          <Trans
            ns="devices"
            i18nKey="addManually.token.releaseListHint"
            components={{
              link: (
                <Anchor
                  href={AGENT_RELEASES_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  size={2}
                />
              ),
            }}
          />
        </Text>
        <Tabs value={os} onValueChange={handleOsChange}>
          <TabsList>
            <TabsTab value="unix">{t("addManually.token.tabUnix")}</TabsTab>
            <TabsTab value="windows">{t("addManually.token.tabWindows")}</TabsTab>
            <TabsIndicator />
          </TabsList>
        </Tabs>
        {os === "unix"
          ? (
              <div className={slots.group()}>
                <Text size={2} color="neutral">
                  {t("addManually.token.installUnix")}
                </Text>
                <CopyableCodeBlock code={unixDownloadCommand} />
                <CopyableCodeBlock code={unixInstallCommand} />
              </div>
            )
          : (
              <div className={slots.group()}>
                <Text size={2} color="neutral">
                  {t("addManually.token.installWindows")}
                </Text>
                <CopyableCodeBlock code={windowsDownloadCommand} />
                <CopyableCodeBlock code={windowsInstallCommand} />
              </div>
            )}
        <Text size={1} color="neutral">
          {t("addManually.token.securityNotice")}
        </Text>
      </div>
    </div>
  );
}
