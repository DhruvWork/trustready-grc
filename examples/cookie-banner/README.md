# Cookie banner examples

Three Vite apps that each run a single banner instance. Shared configuration,
debug, and event UI live in `@trustready/example-cookie-banner-shared`.

| App | Workspace | URL |
| --- | --- | --- |
| Themed | `@trustready/example-cookie-banner-themed` | http://localhost:5180 |
| Themed TCF | `@trustready/example-cookie-banner-themed-tcf` | http://localhost:5181 |
| Headless | `@trustready/example-cookie-banner-headless` | http://localhost:5182 |

The themed TCF app is the IAB CMP validator path: it installs the `__tcfapi`
stub before React boots, then calls `startTCF()` and `registerCookieBanner()`.
The themed app never loads `@trustready/cookie-banner-tcf`. The headless app only
registers headless components.

Banner ID, base URL, and GCM persist in each app's `localStorage` under
`probo-example-config` when the playground is configurable. The apps run
on different ports, so each keeps its own copy.

When both `PUBLIC_COOKIE_BANNER_ID` and
`PUBLIC_COOKIE_BANNER_API_BASE_URL` are non-empty at build time, Vite
inlines them and the app compiles without the configuration form or
`localStorage` config. Leave the ID blank for the local playground.

The IAB-facing static app is the themed TCF build:

```bash
npm -w @trustready/example-cookie-banner-themed-tcf run build
```

## GitHub Pages

`.github/workflows/pages-cookie-banner-themed-tcf.yaml` publishes that
build to this repository's GitHub Pages site. Set Pages source to
GitHub Actions, then add these variables on the
`cookie-banner-example-pages` environment:

- `PUBLIC_COOKIE_BANNER_ID`
- `PUBLIC_COOKIE_BANNER_API_BASE_URL`

Both must be non-empty so Vite bakes the hosted banner in. An empty
id compiles the playground form, which this workflow rejects.

## Run

1. Copy `.env.example` to `.env` in this directory and fill in values.
2. From the repo root:

```bash
npm -w @trustready/example-cookie-banner-themed run dev
npm -w @trustready/example-cookie-banner-themed-tcf run dev
npm -w @trustready/example-cookie-banner-headless run dev
```
