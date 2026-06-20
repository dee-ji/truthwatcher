# Security Policy

Truthwatcher is an evidence-first network cartography project. Security reports are handled conservatively because discovery code may interact with real network infrastructure.

## Supported Versions

Truthwatcher is pre-1.0. Security fixes are provided on the default branch until the project publishes versioned release branches.

## Reporting a Vulnerability

Please do not open a public issue for suspected vulnerabilities.

Report security concerns by opening a private security advisory on GitHub, or by contacting the project maintainers through the repository owner profile if advisories are not available.

Include as much detail as possible:

- Affected version, commit, or branch.
- Impact and affected component.
- Steps to reproduce.
- Relevant logs, payloads, or command output.
- Suggested mitigation, if known.

Maintainers should acknowledge reports within 7 days when possible and keep reporters informed about validation and remediation.

## Security Expectations

- Discovery must remain read-only.
- Arbitrary device commands must not be accepted.
- Credentials must not be committed to the repository.
- Sensitive values must be redacted from logs, audit records, and examples.
- Raw evidence should be handled as potentially sensitive network data.
