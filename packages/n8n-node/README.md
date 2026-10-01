# @trustready/n8n-nodes-trustready

n8n community node package for the [TrustReady](https://www.trustready.io) compliance platform. Automate compliance workflows — manage controls, documents, risks, vendors, cookie banners, and more — over the TrustReady GraphQL API.

This package provides two nodes:

| Node | Type | Description |
|------|------|-------------|
| **TrustReady** | Action | Read and write TrustReady resources (tasks, documents, controls, organizations, and 30+ other resources) |
| **TrustReady Trigger** | Trigger | Start a workflow when TrustReady webhook events occur (document published, user created, obligation updated, and more) |

## Requirements

- A self-hosted n8n instance with [community nodes enabled](https://docs.n8n.io/hosting/configuration/configuration-examples/community-nodes/)
- A TrustReady account with an API key
- n8n 1.0+ (uses the community node package format)

## Installation

Install the package from npm on your self-hosted n8n instance. Only users with the **Owner** or **Admin** role can install community nodes.

### GUI installation (recommended)

1. In n8n, go to **Settings → Community Nodes**.
2. Click **Install**.
3. Enter the npm package name:

   ```
   @trustready/n8n-nodes-trustready
   ```

   To pin a specific version, append it (for example `@trustready/n8n-nodes-trustready@0.199.0`).

4. Accept the community node risk notice and click **Install**.
5. Restart n8n if the new nodes do not appear in the node palette immediately.

See the [n8n GUI installation guide](https://docs.n8n.io/integrations/community-nodes/installation-and-management/gui-installation/) for details.

### Manual installation

If you run n8n in Docker or queue mode, you can install the package manually:

```bash
mkdir -p ~/.n8n/nodes
cd ~/.n8n/nodes
npm install @trustready/n8n-nodes-trustready
```

Restart n8n after installation. See the [manual installation guide](https://docs.n8n.io/integrations/community-nodes/installation/manual-install/) for upgrade and downgrade steps.

## Credentials

All TrustReady nodes use the **TrustReady API** credential type. The node sends your API key as a `Bearer` token on every request.

### Configure credentials in n8n

1. Add a **TrustReady** or **TrustReady Trigger** node to a workflow.
2. Open the **Credential** dropdown and select **Create New Credential**.
3. Fill in the fields:

   | Field | Default | Description |
   |-------|---------|-------------|
   | **TrustReady Server** | `https://us.trustready.io` | Base URL of your TrustReady instance. Use `https://eu.trustready.io` for the EU region, or your own URL when self-hosting. |
   | **API Key** | — | A TrustReady API key with access to the organizations you automate against. |

4. Click **Test** to verify connectivity. n8n calls the TrustReady GraphQL API and checks that the key is valid.
5. Click **Save**. The credential is shared across all TrustReady nodes in your instance.

### Get an API key

1. Sign in to the TrustReady console.
2. Go to **Settings → API Keys**.
3. Click **Create API Key**.
4. Copy the key immediately — it is shown only once.

For self-hosted TrustReady, set **TrustReady Server** to your instance URL (for example `https://trustready.example.com`). The node talks to `/api/console/v1/graphql` on that host.

More detail: [TrustReady n8n authentication docs](https://www.trustready.io/docs/api/n8n/authentication).

## Workflow example: notify Slack when a document is published

This workflow listens for TrustReady document events and posts a message to Slack.

```
TrustReady Trigger  →  Slack
(document      (post message
 published)     with document name)
```

### Steps

1. **Create credentials** as described above and save them as `TrustReady API`.

2. **Add a TrustReady Trigger node**
   - **Credential:** TrustReady API
   - **Organization ID:** your TrustReady organization GID (for example `gid://trustready/Organization/…`)
   - **Events:** `Document Version Published`
   - **Verify Signature:** enabled (recommended)

3. **Add a Slack node** (or any notification node) connected to the trigger output.
   - Map fields from the webhook payload, for example:
     - **Text:** `A document was published: {{ $json.data.documentVersion.document.name }}`

4. **Activate the workflow.** n8n registers a webhook subscription in TrustReady. When a document version is published, TrustReady delivers the event and the Slack message is sent.

### Update events carry the previous state

For `*:updated` events, the payload includes an `updatedFrom` object next to `data`, holding a full snapshot of the entity as it was before the update. This lets a workflow react to what actually changed — for example, only notify when a user's role changes:

- **Condition:** `{{ $json.data.membership.role !== $json.updatedFrom.membership.role }}`
- **Text:** `Role changed from {{ $json.updatedFrom.membership.role }} to {{ $json.data.membership.role }}`

`updatedFrom` is present only on update events; it is absent for created, deleted, and other lifecycle events.

### Alternative: list open tasks on a schedule

Use the **TrustReady** action node without a trigger:

1. Add a **Schedule Trigger** node (for example, every weekday at 9:00).
2. Add a **TrustReady** node:
   - **Resource:** Task
   - **Operation:** Get Many
   - **Organization ID:** your organization GID
   - **Return All:** enabled (or set a **Limit**)
3. Add a downstream node (Slack, email, or spreadsheet) to process `$json` task records.

This pattern works well for daily compliance standups or overdue-task digests.

## Resources

The **TrustReady** node exposes operations across the platform, including:

Access Review, Asset, Audit, Audit Log, Control, Cookie Banner, Cookie Category, Cookie Consent Record, Data, Device, Document, DPIA, Evidence, Finding, Framework, Measure, Obligation, Organization, Processing Activity, Risk, Task, Third Party, Trust Center, User, Vendor, and more.

Use the **Execute** resource to run custom GraphQL queries or mutations when a dedicated operation is not available.

## Links

- [TrustReady documentation](https://www.trustready.io/docs)
- [TrustReady n8n authentication](https://www.trustready.io/docs/api/n8n/authentication)
- [Package on npm](https://www.npmjs.com/package/@trustready/n8n-nodes-trustready)
- [Source code](https://github.com/getprobo/trustready/tree/main/packages/n8n-node)
- [Report an issue](https://github.com/getprobo/trustready/issues)

## License

MIT
