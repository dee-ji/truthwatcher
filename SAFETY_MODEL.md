# TruthWatcher Safety Model

## Safety Goal

TruthWatcher must be safe to run in production service-provider networks.

The initial product must be read-only and must not make configuration changes.

## Core Safety Rules

- No arbitrary commands.
- No configuration mode.
- No commit/write memory.
- No reload/reboot.
- No clear commands.
- No delete commands.
- No copy commands.
- No file transfer by default.
- No brute force credential attempts.
- No credential guessing.
- No scanning outside explicitly allowed targets.
- No write API methods in v0.1.
- No SNMP set.
- No NETCONF edit-config.
- No REST PATCH/POST/DELETE.

## Command Policy

All CLI commands must come from a discovery profile allowlist.

The agent or user may request a task such as:

- identify_device.
- get_inventory.
- get_interfaces.
- get_neighbors.
- get_bgp_summary.

The collector maps tasks to approved commands.

The agent should not send raw commands directly to a device.

## Example Allowed Commands

Cisco-style:

- show version
- show inventory
- show interfaces
- show interfaces brief
- show lldp neighbors
- show cdp neighbors
- show arp
- show ipv6 neighbors
- show bgp summary
- show route summary

Juniper-style:

- show version
- show chassis hardware
- show interfaces terse
- show lldp neighbors
- show arp no-resolve
- show ipv6 neighbors
- show bgp summary
- show route summary

## Example Denied Patterns

Reject commands containing:

- configure
- conf t
- edit
- commit
- write
- reload
- reboot
- request system reboot
- clear
- delete
- remove
- copy
- scp
- ftp
- tftp
- erase
- format
- install
- upgrade
- set
- no

Important: Deny patterns must be vendor-aware. Some vendors use `show configuration` as read-only. The policy engine should distinguish safe show commands from config-changing commands.

## Credential Handling

TruthWatcher should not store raw credentials in the database in early versions.

Use credential references:

- Environment variable reference.
- Local config reference.
- External secret manager reference.
- Development-only static reference.

Never expose credentials to agents or logs.

## Agent Safety

Agents may:

- Read discovered facts.
- Read evidence summaries.
- Propose discovery tasks.
- Explain unknowns.
- Recommend next steps.

Agents may not:

- Receive raw secrets.
- Execute arbitrary commands.
- Bypass policy.
- Expand scope without approval.
- Modify network state.

## Audit Requirements

Every discovery action must log:

- Who/what initiated it.
- Target.
- Discovery run ID.
- Method.
- Task.
- Command/request.
- Timestamp.
- Result status.
- Evidence ID.

## Human Approval Points

Require approval for:

- Expanding to a new subnet/domain.
- Trying a new credential reference.
- Using a new EMS/API integration.
- Increasing concurrency.
- Running discovery against production targets in strict mode.
