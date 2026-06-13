# Truthwatcher BYO Script Examples

These examples demonstrate the local BYO script contract.

Scripts are not enabled by default. A caller must explicitly construct the script runner with `Enabled: true` and an exact allowlist entry for the script path.

Rules:

- Scripts read one JSON object from stdin.
- Scripts write one JSON object to stdout.
- Scripts must not mutate network state.
- Scripts must not guess credentials.
- Scripts must return raw evidence or normalized candidates.
- Scripts must exit `0` only when the JSON response is valid.
- Scripts should use non-zero exit codes for failures.

The sample script emits static fixture-style evidence only. It performs no network access.
