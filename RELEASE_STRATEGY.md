# Release Strategy

This project uses a pull-request-driven release flow that verifies every change before it is merged, tagged, and published as release artifacts.

## Release Flow

```text
PR
  ↓
Tests
Lint
Build
  ↓
Merge
  ↓
Tag

v0.1.0-alpha.1
  ↓

GitHub Actions
  ↓

Release Artifacts
```

## Process

1. **Open a pull request**
   - All release candidates start as a PR.
   - The PR should describe the change, expected impact, and any release notes needed for users.

2. **Run validation checks**
   - Automated checks must pass before merge:
     - Tests
     - Lint
     - Build
   - Failed checks block the release until fixed.

3. **Merge the PR**
   - Merge only after review and successful validation.
   - The merged commit becomes the source for the release tag.

4. **Create a version tag**
   - Tag the merge commit using semantic versioning.
   - Pre-release builds use an alpha suffix, for example:

     ```text
     v0.1.0-alpha.1
     ```

5. **Publish with GitHub Actions**
   - Pushing the tag triggers the release workflow.
   - GitHub Actions builds and packages the release outputs.

6. **Attach release artifacts**
   - Generated artifacts are attached to the GitHub release.
   - Artifacts should be traceable to the tag that produced them.

## Versioning

Use semantic versioning for release tags:

```text
vMAJOR.MINOR.PATCH[-PRERELEASE.NUMBER]
```

Examples:

- `v0.1.0-alpha.1` for the first alpha release.
- `v0.1.0-alpha.2` for the next alpha release.
- `v0.1.0` for the stable release after alpha validation.

## Release Checklist

Before creating a tag, confirm that:

- [ ] The PR has been reviewed and approved.
- [ ] Tests passed.
- [ ] Lint passed.
- [ ] Build passed.
- [ ] Release notes are ready.
- [ ] The merge commit is the intended release source.

After pushing the tag, confirm that:

- [ ] GitHub Actions completed successfully.
- [ ] Release artifacts were generated.
- [ ] Release artifacts were attached to the GitHub release.
- [ ] The release version matches the tag.
