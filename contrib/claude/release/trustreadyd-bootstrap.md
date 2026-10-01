# Release `trustreadyd-bootstrap`

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `trustreadyd-bootstrap/v*`
- **Version source**: `cmd/trustreadyd-bootstrap/VERSION` (single `X.Y.Z` line)
- **Version bump**: Edit `cmd/trustreadyd-bootstrap/VERSION` directly
- **Changelog**: `cmd/trustreadyd-bootstrap/CHANGELOG.md`
- **Files to stage**: `cmd/trustreadyd-bootstrap/VERSION`,
  `cmd/trustreadyd-bootstrap/CHANGELOG.md`
- **Workflow**: `.github/workflows/release-trustreadyd-bootstrap.yaml`
- **Path filter**: `cmd/trustreadyd-bootstrap`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='trustreadyd-bootstrap/v*')..HEAD --oneline \
  -- cmd/trustreadyd-bootstrap
```

If empty or non-user-facing only, do not release this track.

## Notes

CI builds binaries for 9 OS/arch targets and publishes a GitHub Release.
The same binary, built from the tagged ref, is also bundled into the
trustreadyd Docker image when `trustreadyd/v*` runs.
