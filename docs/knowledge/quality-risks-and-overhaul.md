# Quality, Risks, And Overhaul

## Verification baseline

Go tests cover workspace config/paths, project CRUD and nginx templates, installer utilities and legacy PHP patches, selector viewport behavior, Podman command state, doctor behavior, and panel backup/logging helpers. The frontend has Vitest configured, but there is no meaningful broad UI test inventory visible in the repository.

Run the current checks with:

```bash
go test ./...
go build -o build/chauf ./cmd/chauf
cd internal/panel-apps && npm run typecheck && npm run build
```

There are no end-to-end tests for a real `link → start → HTTP request` path, real nginx/FPM lifecycle, dnsmasq, systemd, or Podman.

## High-risk findings

### Credential exposure

Podman passwords are persisted in plain text and returned by the panel detail endpoint. Passwords should be omitted by default, stored with restrictive permissions, and revealed only through an explicit local action if the product truly requires it.

### Unauthenticated panel

The panel assumes localhost trust. That assumption fails if binding, DNS, browser behavior, or future remote access changes. Bind to loopback by default, reject remote binding unless explicitly enabled, and use a local bearer token or equivalent protection.

### Shell-based auto-fix

`doctor --auto-fix` executes generated command strings through `sh -c`. Replace arbitrary command strings with structured, allow-listed operations or require an explicit confirmation for each privileged/system mutation.

### Path handling

Backup and container names are used to derive filesystem paths. Reject separators, `..`, symlinks, unexpected extensions, and names that do not match a strict identifier grammar.

### Workspace split-brain

Podman path resolution does not share the central workspace-root implementation. Fix this before supporting alternate workspace locations or migrations.

## Correctness and trust gaps

- Custom YAML parsing can silently fall back to defaults on malformed configuration.
- Help/docs/implementation disagree on commands and flags.
- README references absent guide files.
- Panel health reports a hard-coded version.
- Dashboard metrics are hard-coded.
- Panel backup behavior is synchronous despite planned job semantics.
- Panel “live” logs are only a single snapshot.
- Generated artifacts such as `build/chauf` and `app_js.o` need an explicit source-control policy.

## Overhaul phases

### Phase 1: Truthful contract

- Build a single command/flag inventory from code.
- Update `--help`, `commands`, docs, README links, and `.agent` plans together.
- Mark `env` and `migrate` as planned or implement them; do not expose dead commands as working.
- Decide whether short legacy commands or grouped commands are canonical.

### Phase 2: Reliable state model

- Centralize workspace root resolution for every package.
- Replace silent config fallback with parse errors that identify file and key.
- Add atomic writes, file permissions, and schema validation.
- Define explicit workspace, project, service, container, backup, and job states.

### Phase 3: Onboarding and diagnosis

- Make `doctor` the first-run preparation step.
- Show dependency plan, disk/time expectations, DNS/SSL readiness, and port choices.
- Make `link` summarize detected project type, PHP choice, domain, SSL, and FPM strategy before mutation.
- Return problem/impact/next-command errors.

### Phase 4: Security and operational safety

- Remove default credential responses.
- Protect the panel and remove wildcard CORS.
- Validate all resource names and paths.
- Make destructive operations visibly confirm scope and backup state.
- Convert auto-fix to structured operations.

### Phase 5: Panel decision

- Either narrow the panel to databases or implement Chauffeur-wide data APIs.
- Replace hard-coded dashboard values with API data.
- Remove placeholder navigation.
- Implement asynchronous backup jobs and true log streaming only if the panel remains a strategic surface.

### Phase 6: Integration confidence

- Add a hermetic workspace test fixture.
- Test config/project/nginx generation together.
- Add mocked process and port lifecycle tests.
- Add an integration path for nginx/FPM request routing where host dependencies permit.
- Add API contract tests shared by Go response types and TypeScript consumers.

## Recommended backlog order

1. Contract and docs reconciliation.
2. Workspace-root consistency.
3. Secret handling and panel boundary.
4. Config parser/error model.
5. Onboarding and `doctor` output.
6. Project-centric status/diagnose output.
7. Panel scope decision and implementation.
8. Integration tests.
9. Environment management, only if it remains a validated user need.
10. Future plugins/monitoring/templates/multi-workspace work.

## Decision rule for new features

Before adding a feature, answer:

- Does it shorten the path to a working project?
- Does it improve observability or safety of an existing operation?
- Can the state be represented consistently in the workspace model?
- Can the CLI, panel, docs, and tests expose the same contract?
- Does it preserve explicit host mutation and user-space isolation?

If the answer is no, defer it until the core experience is trustworthy.
