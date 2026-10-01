# Release `@trustready/skills`

After confirming commits below, follow the
[common steps](./README.md#3-common-steps-every-track).

## Track facts

- **Tag pattern**: `@trustready/skills/v*`
- **Version source**: `packages/skills/package.json`
- **Version bump**: `npm --workspace @trustready/skills version <X.Y.Z> --no-git-tag-version`,
  then set the same version in `packages/skills/plugin.json`,
  `packages/skills/.claude-plugin/plugin.json`, and
  `packages/skills/.codex-plugin/plugin.json`
- **Validate**: `npm --workspace @trustready/skills run validate`
- **Changelog**: `packages/skills/CHANGELOG.md`
- **Files to stage**: `packages/skills/package.json`,
  `packages/skills/plugin.json`,
  `packages/skills/.claude-plugin/plugin.json`,
  `packages/skills/.codex-plugin/plugin.json`,
  `packages/skills/CHANGELOG.md`, `package-lock.json`
- **Workflow**: `.github/workflows/release-npm-skills.yaml`
- **Path filter**: `packages/skills`

## Detect commits

```shell
git log $(git describe --tags --abbrev=0 --match='@trustready/skills/v*')..HEAD --oneline \
  -- packages/skills
```

If empty or non-user-facing only, do not release this track.

## Notes

There is no build step. Run `validate` after the version bump to catch manifest
or structural errors before tagging — it fails when the plugin manifests drift
from `package.json`, and when the Agent Plugins or Agent Skills contracts are
violated. CI runs the same validation, publishes to npm with provenance + SBOM,
and creates a GitHub Release.
