# Lessons And Adoption Plan

## Adopt now

### 1. One truthful operational model

Define shared representations for workspace, project, domain, runtime, FPM pool, service, database, backup, diagnostic, and operation. Render them in CLI output first; use the same data in the panel.

### 2. Guided but explicit onboarding

Add a guided mode to `chauf doctor`, `chauf init`, or `chauf link` that summarizes:

- Detected project type.
- Required and installed PHP versions.
- Domain and aliases.
- DNS mode.
- HTTPS readiness.
- Shared/dedicated FPM recommendation.
- Services that will start.
- Disk/time expectations for source builds.

Require confirmation before mutations and keep noninteractive flags for automation.

### 3. Project-level diagnostics

Implement `chauf project diagnose` or equivalent checks for:

- Project path and document root.
- PHP version availability.
- nginx config and active site.
- FPM socket/process.
- DNS and HTTPS reachability.
- Environment file presence and safe validation.
- Database/service availability where configured.

Return structured checks with severity, evidence, and next command.

### 4. Reversible operations

Before removing or replacing service/database state, snapshot or backup where applicable. Return an operation summary and recovery location. Extend rollback semantics beyond update code.

### 5. Real-world release gate

Adopt a release test rule inspired by Lerd’s “200 rule”: a test phase is not complete until a real linked sample site answers a real HTTP request successfully. The exact status code can allow redirects, but the criterion must validate the user outcome rather than only process state.

### 6. Documentation embedded in the product contract

Use one command metadata source to drive help, command listings, and reference docs where practical. Keep the existing knowledge base as architecture/strategy documentation, not as a second manually drifting command reference.

## Adopt later

### Portable project intent

Add an optional project manifest only after precedence, migration, secret handling, and machine-specific state are defined. Start with PHP, HTTPS, domains, FPM mode, and logical databases/services.

### Worktree awareness

Support branch-derived domains and per-worktree overrides after project identity and configuration are stable.

### Live event stream

Add WebSocket/SSE events only if the panel needs continuous status/log updates. Define event types and reconnection semantics first; do not retrofit a one-shot endpoint with “live” wording.

### Machine-readable APIs and MCP

Expose stable JSON status/diagnostic/action results before adding MCP. MCP should consume the same public operations, not create an alternate behavior layer.

## Do not copy yet

- Full multi-platform support.
- Large service preset marketplace.
- Framework-definition store.
- Public tunnel ecosystem.
- Full TUI and system tray.
- Request profiling/debug bridge.
- Broad worker orchestration.
- Remote dashboard control.

These are valid platform features, but each expands security, support, state, and release complexity. They do not solve Chauffeur’s current primary problems of contract drift, onboarding friction, panel incompleteness, and weak integration verification.

## Proposed Chauffeur overhaul sequence

### Phase A: Contract and safety

- Reconcile CLI docs, flags, README, and `.agent` plans.
- Fix the shared workspace-root abstraction, including Podman.
- Remove password fields from default panel responses.
- Bind panel to loopback and add local authentication/token protection.
- Validate resource names and backup paths.
- Remove wildcard CORS.

### Phase B: Shared state and diagnostics

- Define state/action/result types.
- Add structured JSON output for status and diagnostics.
- Add project diagnosis.
- Make `status` project-centric and include URL, PHP, SSL, FPM, and service evidence.

### Phase C: Onboarding

- Add guided `link`/setup mode.
- Add explicit DNS modes, including a no-system-DNS fallback if compatible with the product boundary.
- Show source-build progress, disk use, and expected duration.
- Make first successful request the completion signal.

### Phase D: Panel decision

- Choose focused database panel or full Chauffeur control plane.
- If focused, remove misleading navigation.
- If full, implement real project/service APIs, data-driven dashboard, accessible empty/error states, and true operation feedback.

### Phase E: Verification

- Add hermetic config/project fixtures.
- Add CLI/API contract tests.
- Add installer tests.
- Add process/service integration tests.
- Add VM or containerized release checks for DNS, SSL, nginx, FPM, and a real HTTP response.

## Decision framework

For every proposed Lerd-inspired feature, ask:

1. Does it reduce time to a working project or improve recovery?
2. Can it use the existing direct-process/workspace model without a new subsystem?
3. Can CLI, panel, docs, and tests share one contract?
4. Does it preserve explicit host mutation and local security?
5. Can it be removed or deferred without trapping state in a new format?

If not, defer it. The comparison’s main recommendation is to copy Lerd’s cohesion and verification discipline, not its entire feature count.
