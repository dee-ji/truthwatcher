# Local Knowledge Providers

Truthwatcher can describe optional external knowledge providers for local agentic development. These providers are development-time context only; they are not required by the runtime server, discovery workflow, parser, API, database migrations, or embedded UI.

## Local Layout

For Codex and GoLand development, keep Truthwatcher and Mistspren as sibling repositories:

```sh
export GOLAND_PROJECTS="$HOME/GolandProjects"
export TRUTHWATCHER_HOME="$GOLAND_PROJECTS/truthwatcher"
export MISTSPREN_HOME="$GOLAND_PROJECTS/mistspren"
```

Expected layout:

```text
~/GolandProjects/
  truthwatcher/
  mistspren/
```

`TRUTHWATCHER_HOME` points to the local Truthwatcher repository root.
`MISTSPREN_HOME` points to the local Mistspren repository root.

## Development Config

The local development config lives at [truthwatcher.yaml](../truthwatcher.yaml). It records project metadata and optional knowledge providers:

```yaml
project:
  name: truthwatcher
  repo: github.com/dee-ji/truthwatcher
  local_path: ${TRUTHWATCHER_HOME}

knowledge:
  providers:
    - name: mistspren-local
      type: filesystem
      enabled: true
      root: ${MISTSPREN_HOME}
      purpose:
        - memory
        - adr
        - architecture-decisions
        - agent-loop-context

    - name: mistspren-github
      type: github
      enabled: false
      repo: github.com/dee-ji/mistspren
      branch: main
      purpose:
        - memory
        - adr
        - architecture-decisions
        - agent-loop-context
```

Filesystem providers are preferred for local Codex and GoLand workflows because they use the checked-out sibling repo directly. GitHub provider references are allowed for future remote agent workflows, but they stay disabled by default.

## Validation

Build and check local knowledge provider resolution:

```sh
make build
./bin/truthwatcher dev check-knowledge
```

The command expands environment variables such as `${MISTSPREN_HOME}` and reports each provider with:

- provider name
- provider type
- enabled or disabled
- resolved path or repo
- status: `available`, `missing`, `disabled`, or `misconfigured`

If `MISTSPREN_HOME` is set and points to an existing directory, `mistspren-local` reports `available`. If `MISTSPREN_HOME` is missing or the directory does not exist, it reports `missing` without failing application startup.

## Agent Usage

Codex and other local development agents should run the knowledge check before consulting Mistspren or any other optional external provider:

```sh
truthwatcher dev check-knowledge
```

If the binary is not built, agents may run the same check through Go:

```sh
go run ./cmd/truthwatcher dev check-knowledge
```

Agents must follow the reported provider status:

- `available`: the provider may be consulted for development-time memory, ADR, architecture decision, and agent loop context.
- `missing`: continue with Truthwatcher-local code and docs only.
- `disabled`: do not consult the provider.
- `misconfigured`: treat the provider as unavailable until the local config is fixed.

Agents must not infer provider availability from repository names, sibling directory conventions, or GitHub URLs without running the check. A missing or unavailable provider is not an error for normal development work.

## Boundary Rules

Mistspren is optional development-time memory, ADR, architecture decision, and agent loop context. Production Truthwatcher must not import Mistspren code, shell out to Mistspren, vendor Mistspren content, or require the Mistspren repository to exist.

The runtime binary remains independent from configured knowledge providers. The `server`, `migrate`, `discover`, `parse`, `export`, and `import` commands do not load `truthwatcher.yaml` for Mistspren.
