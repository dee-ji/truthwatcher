# Releases

Truthwatcher uses Git tags as the single source of truth for released versions.

## Version source of truth

- Maintainers intentionally create semantic version tags, such as `v0.1.0`, `v0.1.0-alpha.1`, `v0.1.0-beta.1`, or `v0.1.0-rc.1`.
- Release automation never invents, increments, or guesses the next version.
- Code should not manually hardcode release versions. The repository's development default is `dev`.
- Local and development builds that do not pass release ldflags report:
  - `Version = "dev"`
  - `Commit = "unknown"`
  - `BuildDate = "unknown"`
- Release builds inject version metadata from the Git tag, commit SHA, and UTC build time.

## Release builds

Release builds set the internal version package with Go linker flags:

```sh
go build \
  -ldflags "-X 'truthwatcher/internal/version.Version=${VERSION}' \
            -X 'truthwatcher/internal/version.Commit=${COMMIT}' \
            -X 'truthwatcher/internal/version.BuildDate=${BUILD_DATE}'" \
  -o dist/truthwatcher ./cmd/truthwatcher
```

`VERSION` must be the Git tag that triggered the release. `COMMIT` must be the Git SHA being built. `BUILD_DATE` must be a UTC timestamp.

## Stable and prerelease tags

Only semantic version tags trigger releases:

- `v0.1.0-alpha.1` creates a GitHub prerelease.
- `v0.1.0-beta.1` creates a GitHub prerelease.
- `v0.1.0-rc.1` creates a GitHub prerelease.
- `v0.1.0` creates a stable GitHub release.

Tags containing `-alpha`, `-beta`, or `-rc` are prereleases. Tags without a prerelease suffix are stable releases.
