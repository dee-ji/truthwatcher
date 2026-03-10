# AGENTS.md

## Overview

This repository contains Truthwatcher, an open-source intent-driven network management platform.

Truthwatcher allows engineers to:

- Declare high-level network intent
- Translate intent into vendor-specific configurations
- Deploy changes safely in phased rollouts
- Continuously monitor the network for configuration drift
- Maintain a source of truth for network state

The system architecture is inspired by large-scale infrastructure control planes used at hyperscalers and is implemented primarily in Go.

---

# Architecture Concepts

Truthwatcher is composed of several core subsystems.

These names are part of the project's architectural vocabulary and should remain consistent across the codebase.

## Radiant
Primary control plane and orchestration service.

Responsibilities:

- Intent lifecycle management
- Workflow coordination
- Interaction between subsystems
- Control plane state management

Binary:

cmd/radiant

---

## Spanreed
External interface layer.

Responsibilities:

- REST / gRPC APIs
- authentication and authorization
- Web UI backend
- CLI integration

Binary:

cmd/spanreed

---

## Archive
The network source of truth.

Responsibilities:

- device inventory
- configuration state
- topology information
- intent storage

Likely backed by PostgreSQL.

Package:

internal/archive

---

## Ideals
Intent modeling and validation system.

Responsibilities:

- intent schema definitions
- validation of intent objects
- policy enforcement
- intent compilation inputs

Package:

internal/ideals

---

## Elsecall
Intent translation engine.

Responsibilities:

- convert intent into vendor-specific configuration
- support multiple vendor targets
- generate configuration artifacts

Example targets:

- Junos
- IOS-XR
- EOS
- SR-OS

Package:

internal/elsecall

---

## Shadesmar
Network topology and dependency graph engine.

Responsibilities:

- network adjacency modeling
- dependency graph generation
- blast radius analysis
- topology queries

Package:

internal/shadesmar

---

## Oathgate
Pre-deployment safety and simulation engine.

Responsibilities:

- change previews
- configuration diff generation
- topology impact analysis
- simulation of rollout plans

Package:

internal/oathgate

---

## Highstorm
Deployment engine.

Responsibilities:

- phased change rollouts
- canary deployments
- batch scheduling
- rollback management

Binary:

cmd/highstorm

---

## Stormlight
Drift detection and reconciliation engine.

Responsibilities:

- monitoring deployed configuration
- detecting drift from intent
- compliance validation
- reconciliation requests

Binary:

cmd/stormlight

---

## Seekers
Network discovery and topology ingestion system.

Responsibilities:

- device discovery
- LLDP / CDP ingestion
- BGP topology mapping
- inventory synchronization

Binary:

cmd/seekers

---

## Squire
Distributed execution worker.

Responsibilities:

- execute network operations
- apply configuration
- gather device telemetry
- run jobs dispatched by Highstorm or Radiant

Binary:

cmd/squire

---

# Repository Layout

The repository follows a Go monorepo layout.

cmd/
Entry points for deployable services.

Each service has its own binary.

Example:

cmd/radiant  
cmd/spanreed  
cmd/highstorm  
cmd/stormlight  
cmd/seekers  
cmd/squire  
cmd/twctl

---

internal/
Private application logic not intended for external consumption.

Modules include:

internal/archive  
internal/ideals  
internal/elsecall  
internal/shadesmar  
internal/oathgate  
internal/highstorm  
internal/stormlight  
internal/seekers  
internal/squire  
internal/platform  

---

pkg/
Public APIs and reusable client libraries.

Examples:

pkg/api  
pkg/client  
pkg/version  

Only place code here that is intended to be used by external projects.

---

configs/
Example configuration files for development, testing, and production deployments.

---

deployments/
Infrastructure deployment manifests.

May include:

- docker-compose
- Kubernetes
- Helm charts
- systemd services

---

docs/
Architecture documentation and developer guides.

Expected subdirectories:

docs/architecture  
docs/operators  
docs/developers  
docs/tutorials  
docs/adr  

---

examples/
Example intent definitions and configuration scenarios.

---

test/
Integration, end-to-end, and performance testing.

---

# Development Guidelines

## Language

Primary language: Go

Guidelines:

- Follow idiomatic Go
- Prefer composition over inheritance
- Avoid premature abstraction
- Keep packages cohesive and small
- Write code that compiles cleanly even if functionality is incomplete

---

## Package Design

Packages should represent clear domain concepts.

Avoid:

- circular dependencies
- cross-package leakage
- overly generic utility packages

Prefer domain-driven boundaries.

---

## Internal vs Public Code

Use:

internal/

for application logic.

Only place code in:

pkg/

if it is intended for external reuse.

---

## Placeholder Implementations

During early development:

- Use TODO comments
- Provide stub implementations
- Keep compilation working

Do not guess complex implementation logic prematurely.

---

# Logging and Observability

Truthwatcher should support:

- structured logging
- metrics
- tracing

Preferred approach:

- structured logs
- minimal log noise
- meaningful error context

---

# Testing Expectations

The project should include:

- unit tests
- integration tests
- end-to-end tests

Tests should prioritize behavioral correctness over implementation detail.

---

# Security Considerations

Truthwatcher interacts with network infrastructure and must follow strict safety practices.

Guidelines:

- never log credentials
- avoid storing secrets in configuration files
- validate all external inputs
- implement least-privilege authentication

---

# Scope Guardrails

Early development should focus on core platform structure, not full functionality.

Avoid implementing:

- full network protocol stacks
- vendor-specific configuration complexity
- heavy plugin frameworks

Focus first on:

- repository structure
- clean interfaces
- architecture clarity
- compilation-safe scaffolding

---

# Contribution Guidelines

When contributing:

1. Maintain architectural consistency.
2. Add documentation for new components.
3. Avoid introducing unnecessary dependencies.
4. Keep pull requests focused and small.
5. Update tests when behavior changes.

---

# Project Philosophy

Truthwatcher is designed around the principle that the network should continuously converge toward declared truth.

Key ideas:

- Intent is the source of authority.
- Deployments must be safe and reversible.
- The system must continuously validate the real network.
- Drift should be detected and reconciled automatically.

Truthwatcher exists to ensure that the deployed network always reflects the intended network.

---

## Security-Sensitive Contribution Guidance

Changes in authentication, authorization, identity mapping, request middleware, and deployment safety checks are security-sensitive.

When modifying these paths:

1. Prefer explicit allow/deny checks near endpoint handlers rather than implicit magic.
2. Keep local-dev bypass behavior loudly visible in logs and docs.
3. Add or update tests for both allowed and denied authorization outcomes.
4. Document new permissions, default roles, and migration impacts.
5. Avoid introducing hidden default credentials or production-insecure fallbacks.
