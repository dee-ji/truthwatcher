# Worker and Queue Execution Model

Truthwatcher now includes a simulation-first worker framework for deployment execution.

## Components

- `tw-server`:
  - accepts deployment plans
  - creates a queued deployment run
  - enqueues a `deploy` job to Redis-backed queue abstraction
- `tw-worker`:
  - dequeues compile/deploy/reconcile jobs
  - executes deploy jobs in deterministic simulation mode only
  - updates run status transitions and target results
  - emits audit events for queued/started/finished/stopped

## Queue abstraction

`internal/queue` provides a small interface:

- `Enqueue`
- `Dequeue`
- `Depth`

Implementations:

- `RedisQueue`: production-oriented list queue (`LPUSH`/`BRPOP`)
- `InMemoryQueue`: deterministic unit-test fallback

## Deployment run lifecycle

Statuses:

- `queued`
- `running`
- `succeeded`
- `failed`
- `stopped`

Worker simulation:

- deterministic target result from hash(run_id, target)
- no SSH/NETCONF/device-side execution
- result text persisted on each target (`simulated success` / `simulated failure`)

## Stop-condition hooks

`StopConditionEvaluator` allows future policy checks.

Current implementation:

- supports error-rate threshold comparisons (for example `>50%`)
- can stop the run and mark status `stopped` with a reason

## Metrics

In-process metrics are collected for:

- queued job count
- run durations

This is intentionally simple scaffolding to support future Prometheus/OpenTelemetry exporters.

## Future adapter readiness

The framework intentionally keeps execution mode as simulation-only.

Future work can add:

- real vendor adapters
- richer queue semantics (visibility timeout, retries, dead-letter)
- durable execution repositories shared between API and worker processes

without changing the queue/job model or run state machine.
