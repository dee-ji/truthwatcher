# Topology Foundation Walkthrough

This walkthrough demonstrates the first safe topology foundation for Truthwatcher using relational persistence and in-memory graph assembly.

## 1. Import fixture data

```bash
tw-topology import --file=examples/topology/fabric-small.yaml
```

The import payload includes:

- vendors
- platforms
- sites
- devices
- interfaces
- links

## 2. Query topology devices

```bash
curl 'http://localhost:8080/api/v1/topology/devices?site=dc1&vendor=eos'
```

## 3. Query links

```bash
curl 'http://localhost:8080/api/v1/topology/links?platform=7050'
```

## 4. Query device detail

```bash
curl 'http://localhost:8080/api/v1/topology/devices/44444444-4444-4444-4444-444444444441'
```

This endpoint includes simple adjacency output for neighboring devices.

## 5. Query adjacency from CLI

```bash
tw-topology query adjacency --device-id=44444444-4444-4444-4444-444444444441
```

## 6. Export snapshot

```bash
tw-topology export --out=/tmp/topology-export.json
```

## Extension points

The topology service includes placeholders for future advanced graph operations:

- blast-radius analysis
- path analysis

These are intentionally not implemented in this foundational step.
