#!/bin/sh
# Safe local example: reads JSON from stdin, emits static fixture-style evidence.
# It performs no network access and does not mutate any external system.

cat >/dev/null
cat <<'JSON'
{
  "evidence": [
    {
      "target": "fixture://junos-mx",
      "method": "script",
      "command_or_api": "show version",
      "raw_output": "Hostname: fixture-junos-mx\nModel: mx480\nJunos: 22.4R1",
      "metadata": {
        "script": "emit_static_version.sh",
        "network_access": false
      }
    }
  ],
  "warnings": [
    "static example only; no network access was performed"
  ]
}
JSON
