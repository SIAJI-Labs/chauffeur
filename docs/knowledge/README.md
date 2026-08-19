# Chauffeur Knowledge Base

This directory is the implementation-grounded baseline for an experience overhaul. It describes what Chauffeur is intended to be, what the repository actually implements, and which gaps must be resolved before adding more surface area.

## Reading order

1. [Product and strategy](./product-and-strategy.md)
2. [Architecture and runtime](./architecture-and-runtime.md)
3. [Feature inventory](./feature-inventory.md)
4. [CLI and user journeys](./cli-and-user-journeys.md)
5. [Panel and integrations](./panel-and-integrations.md)
6. [Quality, risks, and overhaul plan](./quality-risks-and-overhaul.md)

## Source of truth

The code is authoritative for current behavior. Product intent and planned behavior are recorded in `.agent/PRD.md`, `.agent/SPECS.md`, `.agent/specs/`, and `.agent/IMPLEMENTATION_PLAN.md`. Where they disagree, this knowledge base calls out the discrepancy instead of silently choosing one.

## Snapshot

- Product: Linux-first local PHP development environment inspired by Valet/Herd.
- Primary workflow: install services, link a project, start services, visit a local `.test` domain.
- Core runtime: user-space nginx, PHP-FPM, Composer, optional mkcert and dnsmasq integration.
- Optional database workflow: Podman-managed MySQL, PostgreSQL, MariaDB, MongoDB, and Redis.
- Admin surface: embedded Go HTTP server with a partially complete React panel.
- Current repository maturity: core CLI workflow is substantially implemented; configuration, environment management, panel completeness, and documentation truthfulness need overhaul.
