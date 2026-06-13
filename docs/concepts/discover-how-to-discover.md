# Discover How To Discover

Truthwatcher is built for environments where the hard problem is not just finding devices. The hard problem is learning the safe path to better knowledge.

That is "discover how to discover."

## The Problem

Large provider networks may have:

- partial source-of-truth data
- inconsistent naming
- multiple vendors
- multiple EMS or controller systems
- incomplete credential paths
- unknown route reflectors
- stale monitoring or IPAM records
- fragmented regional ownership

A tool that assumes one perfect inventory source will fail in these environments.

## The Truthwatcher Approach

Truthwatcher starts from limited seed input and asks conservative questions:

- What is already known?
- What is unknown?
- What evidence supports the current model?
- Which safe read-only task could improve the model?
- Which target is in scope?
- Which profile and method are appropriate?
- What evidence should the step produce?
- Does a human need to approve execution?

Discovery planning is separate from discovery execution.

## Planning Is Not Execution

Current discovery plans are suggestions. They include:

- target
- method
- profile
- task
- reason
- expected evidence
- risk level

Plans require human approval and do not auto-execute. They must not suggest credential guessing, broad scanning, unsafe command execution, or scope expansion.

## Fixture Discovery

The fake collector is the safe local workflow for development and demos:

```sh
./bin/truthwatcher discover fake --target fixture://junos-mx
```

It reads command outputs from `examples/fixtures`, validates the requested tasks against policy, and stores raw evidence in PostgreSQL. It does not connect to a network.

## Real Discovery Boundary

Optional SSH collection exists behind the same collector interface, but it is not the main workflow. Any real collector must:

- use explicit target configuration
- use read-only commands from a profile
- pass policy checks before command execution
- avoid credential guessing and brute force
- store raw evidence before facts are created
- audit target, command, profile, user/context, and timestamps

## Current Limitation

Truthwatcher can collect fixture-backed evidence and expose graph APIs. Automatic parser output persistence into assets, facts, and relationships is intentionally separate work. Until that wiring exists, graph inspection depends on model data already persisted through available APIs, imports, tests, or future parser persistence work.
