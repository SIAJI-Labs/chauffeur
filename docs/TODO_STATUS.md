# Chauffeur Development Tracker

_A living status board for features, debt, and priorities. Keep this in sync with README.md and AGENTS.md._

## 1. Snapshot
- **Maintainer**: @siegg (solo, learning Go; heavy AI assistance)
- **Stability**: Experimental – validated mostly on one Arch-based workstation
- **CI**: `go test ./...` currently red; needs attention before public release
- **Work focus**: Logging overhaul, service orchestration polish, test coverage

## 2. Completed ✅
### Workspace & Bootstrap
- Smart installer (`install.sh`) with Go prerequisite checks
- `chauf init` scaffolds `~/.chauffeur` with default config, templates, PATH guidance
- Workspace layout contracts documented in AGENTS.md

### Project Registration
- `chauf link` / `links` / `unlink` end-to-end
- Project type detection (Laravel, WordPress, general)
- Nginx template generation + symlinks in `sites-available/enabled`
- Table-formatted `chauf links`

### PHP & Composer Fundamentals
- `chauf php install/use/isolate`
- Project-aware PHP shim
- Composer installer + shim that reuses Chauffeur PHP

### Self-update
- `chauf self-update` fetches latest git release
- `chauf self-update --dev` rebuilds from current repo when run inside it

## 3. In Progress 🚧
1. **Structured logging compliance** – many commands still use `fmt.Printf`; need to migrate to `lib.Logger` helpers
2. **Service orchestration stability** – refine `chauf start/stop/status` to handle mixed project states, better error surfacing
3. **dnsmasq/NetworkManager automation** – ensure instructions are clear, reversible, and logged
4. **Test suite repair** – fix current failures (e.g., tests importing `cli/internal/...`) and restore passing `go test ./...`
5. **Documentation sync discipline** – ongoing effort to keep README, AGENTS, and this tracker aligned

## 4. Planned 📋
| Priority | Item | Notes |
|----------|------|-------|
| P1 | Replace all direct `fmt.Printf` output with logger calls | Blocking for polished release |
| P1 | Rework tests to avoid importing internal packages | Required for Go module hygiene |
| P1 | Stabilize `chauf start/stop` interaction with dnsmasq & port forwarding | Needs integration tests |
| P2 | Expand PHP installer matrix (8.2/8.1/7.4 legacy deps) | Ensure workspace fallback works |
| P2 | Improve `chauf status --detail` output (tables, health info) | Align with logging spec |
| P3 | Add onboarding docs for contributors (Go basics + AI workflow) | Help new maintainers |
| P3 | Publish release checklist (binary build, docs sync, testing) | Needed for first public release |

## 5. Known Issues / Tech Debt
- `go test ./...` fails due to tests importing `cli/internal/config` (Go disallows external packages from accessing internal)
- `chauf status` and `links` still use raw prints; violates AGENTS logging policy
- Repo currently contains built binaries (`chauf`, `main`) checked in by mistake – remove and add to `.gitignore`
- DNS setup path requires clearer rollback instructions when NetworkManager/systemd-resolved changes are applied manually

## 6. Testing & QA
- Target: `go test ./...` green on Go 1.22+
- Integration tests should stub HOME to temp directories, avoiding host mutation
- Add smoke test workflow (start/link/status) once logging and dnsmasq flow stabilize

## 7. Release Readiness Checklist
- [ ] `go test ./...` passes locally
- [ ] README.md / docs/TODO_STATUS.md / AGENTS.md agree on feature status
- [ ] No compiled binaries or caches tracked in git
- [ ] Logging contract enforced across commands
- [ ] dnsmasq/NetworkManager instructions reviewed and verified

## 8. How to Help
- Contribute logging refactors (replace `fmt.Printf` with `lib.Logger` calls)
- Improve tests around `chauf link`/`links` to avoid double-run conflicts
- Document real-world setups (distro, Go version, dnsmasq config) in issues to broaden coverage
- Review AGENTS.md and propose clarifications before building new features

_Last updated: 2025-11-09T16:05:48Z_
