# Release Helm chart (`trustready`)

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `helm/v*`
- **Version source**: `contrib/helm/charts/trustready/Chart.yaml` (`version` field)
- **Version bump**: Edit `version` in `contrib/helm/charts/trustready/Chart.yaml`
- **Changelog**: `contrib/helm/charts/trustready/CHANGELOG.md`
- **Files to stage**: `contrib/helm/charts/trustready/Chart.yaml`,
  `contrib/helm/charts/trustready/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-helm.yaml`
- **Path filter**: `contrib/helm`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='helm/v*')..HEAD --oneline \
  -- contrib/helm
```

If empty or non-user-facing only, do not release this track.

## Notes

The chart has its own SemVer (`version`). `appVersion` in `Chart.yaml` is
the default trustreadyd application version the chart deploys (image tag
`v<appVersion>`). Bump `appVersion` when the chart should default
to a newer trustreadyd release.

CI packages the chart and pushes it to
`oci://artifact.trustready.io/trustready/trustready`, then publishes a GitHub Release.
