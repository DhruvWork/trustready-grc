# Release `trustreadyctl`

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `trustreadyctl/v*`
- **Version source**: `cmd/trustreadyctl/VERSION` (single `X.Y.Z` line)
- **Version bump**: Edit `cmd/trustreadyctl/VERSION` directly
- **Changelog**: `cmd/trustreadyctl/CHANGELOG.md`
- **Files to stage**: `cmd/trustreadyctl/VERSION`, `cmd/trustreadyctl/CHANGELOG.md`
- **Path filter**: `cmd/trustreadyctl pkg/trustreadyctl`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='trustreadyctl/v*')..HEAD --oneline \
  -- cmd/trustreadyctl pkg/trustreadyctl
```

If empty or non-user-facing only, do not release this track.
