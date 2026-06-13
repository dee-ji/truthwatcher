# Local Install

Truthwatcher is packaged as one Go binary plus PostgreSQL. The binary embeds database migrations and the static UI, so there is no container, Helm chart, external web server, or separate migration artifact required.

## Prerequisites

- Go matching `go.mod`.
- PostgreSQL running locally or reachable from the workstation.
- Standard PostgreSQL client tools such as `createdb` and `psql`.

## Build The Binary

```sh
make build
```

The development binary is written to:

```text
bin/truthwatcher
```

For a local release-style build:

```sh
make release-local
```

The release binary is written to:

```text
dist/truthwatcher-<goos>-<goarch>/truthwatcher
```

## Configure PostgreSQL

Create a local database using your preferred PostgreSQL user and permissions model. One simple local setup is:

```sh
createdb truthwatcher
```

Then export the database URL:

```sh
export TRUTHWATCHER_DATABASE_URL='postgres://localhost/truthwatcher?sslmode=disable'
```

Other supported environment variables:

```sh
export TRUTHWATCHER_ADDR='127.0.0.1:8080'
export TRUTHWATCHER_LOG_LEVEL='info'
export TRUTHWATCHER_DEV_MODE='false'
```

## Run Migrations

Migrations are embedded in the binary.

```sh
./bin/truthwatcher migrate status
./bin/truthwatcher migrate up
```

For a release-local binary, replace `./bin/truthwatcher` with the binary under `dist/`.

## Start The Server

```sh
./bin/truthwatcher server
```

The server listens on `TRUTHWATCHER_ADDR`, defaulting to `127.0.0.1:8080`. The embedded UI is served by the same binary.

Useful endpoints:

```text
GET /healthz
GET /readyz
GET /api/v1/version
```

## Run Fake Discovery

Fake discovery uses local fixtures and does not touch a network.

```sh
./bin/truthwatcher discover fake --target fixture://junos-mx
```

Optional flags:

```text
--profile   Built-in discovery profile. Inferred from target when omitted.
--tasks     Comma-separated safe discovery tasks.
--fixtures  Fixture root directory. Defaults to examples/fixtures.
```

## CLI Help

```sh
./bin/truthwatcher --help
./bin/truthwatcher server --help
./bin/truthwatcher migrate --help
./bin/truthwatcher discover fake --help
./bin/truthwatcher version --help
```

## Packaging Boundary

Truthwatcher local packaging intentionally avoids Docker, Kubernetes, Helm, message brokers, and cloud dependencies. PostgreSQL is the only runtime service required by the packaged binary.
