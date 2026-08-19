# Experience Comparison

## First success

### Chauffeur today

The documented path is:

```bash
chauf init
chauf install nginx php 8.3 composer
chauf link --secure
chauf start
```

This is explicit and aligned with Chauffeur’s zero-silent-mutation philosophy, but it asks users to understand dependencies, service installation, DNS, SSL, and startup as separate concepts.

### Lerd reference

Lerd centralizes setup through `lerd install`, supports installer-driven prerequisite handling, and then uses `lerd link`/`lerd open`. Interactive linking can ask about PHP, HTTPS, databases, services, workers, and framework settings. Noninteractive paths remain available for CI and scripted use.

### Transferable lesson

Create a guided first-run path without hiding the operations:

```text
doctor → dependency plan → install choice → project detection → URL/SSL/FPM summary → start → open
```

Show exactly what will change, but do not require users to discover the correct command sequence from separate docs.

## Project configuration

### Chauffeur today

Project configuration lives under the Chauffeur workspace. This keeps project directories clean and follows the “no project mutation” principle, but it makes project setup less portable and less visible when projects are shared or cloned.

### Lerd reference

Lerd stores portable intent in `.lerd.yaml`, including domains, PHP/Node versions, framework, services, workers, runtime, and project settings. Machine-specific runtime state remains in XDG data directories. Service references can be logical names, presets, or inline definitions.

### Transferable lesson

Consider a minimal optional project manifest for portable intent, not a wholesale move of all state into the repository. A useful first version could describe:

- PHP version.
- HTTPS preference.
- Primary domain/aliases.
- FPM strategy.
- Required managed databases/services.

The manifest must not contain passwords, absolute workspace paths, host-specific ports, or generated runtime state.

## Discovery and navigation

### Chauffeur today

The CLI is the primary interface. The embedded panel has container pages, but dashboard values and several navigation links are placeholders.

### Lerd reference

Lerd presents the same operational model through CLI, web dashboard, TUI, tray, and MCP. The web UI has Dashboard, Sites, Services, System, docs, command palette, detail panes, notifications, and contextual actions. The TUI is designed for SSH and tmux.

### Transferable lesson

Do not add multiple shells until the underlying state/actions are shared. The immediate Chauffeur opportunity is a project-centric status/diagnose model that could later feed both CLI and panel.

## Progressive disclosure

Lerd exposes advanced features through site/service detail views, modals, command palette actions, and contextual controls rather than putting every option on the first screen.

Chauffeur’s CLI is already explicit, but the panel currently exposes broad navigation with unfinished destinations. Prefer a smaller complete panel over a broad incomplete one.

## Failure visibility

Lerd distinguishes environment health from application health:

- `doctor` checks host/runtime prerequisites.
- `site:doctor` checks project/framework state.
- `dns:check` gives layered DNS evidence.
- Worker-health UI surfaces failed workers.
- Logs, request stats, profiling, and debug capture explain runtime behavior.

Chauffeur has a strong environment `doctor` foundation, but it lacks an equivalent project/application doctor and a unified diagnostic summary.

Recommended Chauffeur model:

```text
chauf doctor             host/workspace health
chauf project diagnose   linked project/runtime health
chauf status             concise current state
chauf logs               evidence for the selected failure
```

## Reversibility

Lerd snapshots databases before destructive service actions and supports rollback/recovery for service updates and migrations. Chauffeur has backup/restore functionality for Podman databases, but the operation model is not yet uniformly transactional or job-oriented.

Adopt the rule: any destructive action should state scope, create or require a recoverable snapshot where relevant, and report a durable result.


Lerd’s `.localhost` mode gives users a lower-privilege path when DNS or HTTPS integration is undesirable. Its embedded docs and offline PWA behavior also help when parts of the environment are unavailable.

Chauffeur should consider:

- A documented `.localhost` or explicit `127.0.0.1` fallback.
- Clear HTTP-only behavior when mkcert is unavailable.
- Local embedded command help/docs for offline troubleshooting.
