# Release `trustreadyd` (server group)

This track ships `trustreadyd`, `@trustready/console`, `@trustready/compliance-portal`,
`@trustready/employee-portal`, and `@trustready/ui` together as the Docker image and
accompanying binary archive. They share the same version.

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `trustreadyd/v*`
- **Version source**: `cmd/trustreadyd/VERSION` (single `X.Y.Z` line)
- **Version bump**: Edit `cmd/trustreadyd/VERSION` directly
- **Changelog**: `cmd/trustreadyd/CHANGELOG.md` (covers every bundled component)
- **Files to stage**: `cmd/trustreadyd/VERSION`, `cmd/trustreadyd/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-trustreadyd.yaml`
- **Path filter**: `cmd/trustreadyd apps/console apps/compliance-portal apps/employee-portal packages/ui pkg`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='trustreadyd/v*')..HEAD --oneline \
  -- cmd/trustreadyd apps/console apps/compliance-portal apps/employee-portal packages/ui pkg
```

If empty or non-user-facing only, do not release this track.

## Notes

The changelog covers changes across every bundled component (`trustreadyd`,
`@trustready/console`, `@trustready/compliance-portal`, `@trustready/employee-portal`,
`@trustready/ui`).

CI builds the frontends and Go binaries, builds and pushes the
multi-arch image to `artifact.probo.inc/probo/probo:v<version>` (and
`:latest`), runs Trivy + cosign + attestations, and publishes the GitHub
Release.
