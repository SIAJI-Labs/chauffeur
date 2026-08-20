# Feature And Interface Matrix

| Capability | Chauffeur | Lerd | Overhaul implication |
|---|---|---|---|
| PHP version management | Source-built PHP 7.4–8.4, global/project selection | PHP 8.1–8.5 plus legacy 7.4/8.0, fetch/rebuild/extensions | Keep Chauffeur’s core; improve install expectations and project visibility. |
| nginx and local domains | Direct workspace nginx, `.test`, aliases | Container nginx, `.test`, `.localhost`, groups, sharing | Add honest fallback mode and project-centric URL state. |
| SSL | mkcert secure/unsecure per project | Managed HTTPS with browser trust and transitions | Preserve explicit SSL; surface certificate health and recovery. |
| FPM isolation | Shared or dedicated pools | Shared version containers, FrankenPHP/custom runtimes | Keep the simpler two-strategy model until demand proves broader runtimes. |
| Databases | Optional Podman engines and backup/restore | First-class services, presets, snapshots, migration, admin UIs | Improve operation safety; avoid expanding engines before core UX is reliable. |
| Project setup | Explicit link; config external to project | Link/init/setup wizard, committed `.lerd.yaml`, framework detection | Add optional portable intent and guided linking. |
| Environment files | Planned/broken `env` command | Full env editing, restore, overrides, checks | Implement only after secret/storage model is defined. |
| Workers | Dedicated FPM focus; no comparable broad worker system | Queue, schedule, Horizon, Reverb, generic workers, self-healing | Defer broad workers; expose project process health first. |
| Worktrees | Not first-class | Branch domains, overrides, isolated databases | Candidate later feature after project identity is stable. |
| Diagnostics | Environment doctor, DNS/network/dependency checks | Environment doctor, site doctor, DNS check, worker health | Add project/application diagnosis and structured check results. |
| Logs | CLI file tail/follow; panel snapshot SSE | CLI/UI live logs across runtime and app sources | Make panel semantics honest; implement real streaming only if needed. |
| Observability | PID, uptime, memory, paths; basic status | Request stats, profiling, debug bridge, slow routes, notifications | Add the smallest useful project request signal before advanced tooling. |
| Web UI | Partial Podman panel | Full dashboard with sites/services/system/docs | Narrow or complete Chauffeur panel; remove placeholders. |
| TUI | Bubble Tea selectors only | Full operational TUI for SSH/tmux | Consider only after shared action/state model exists. |
| AI/MCP | No current MCP surface | MCP tools for site/service/db/env/runtime/diagnostics | Potential differentiator, but not before stable machine-readable APIs. |
| Sharing | No comparable built-in flow | LAN and public tunnel integrations | Defer; sharing increases security and support scope. |
| Platform | Linux-first | Linux/macOS, WSL2 beta | Preserve Linux focus unless platform expansion is strategic. |
| Recovery | Podman backup/restore and update flows | Snapshots, rollback, migration safety | Make destructive actions consistently reversible. |
| Documentation | CLI docs plus AI-oriented `.agent` docs, some drift | Embedded docs, extensive guides, command reference | Generate/reconcile docs from a truthful command contract. |
| Testing | Focused Go unit tests, limited integration coverage | Broad Go/frontend tests, Bats installer tests, VM release plan | Adopt one real HTTP success gate and a release matrix. |

## Highest-value gaps exposed by comparison

1. Chauffeur lacks a single shared state model across CLI and panel.
2. Chauffeur’s project onboarding is explicit but not guided.
3. Chauffeur has environment diagnosis but not project/application diagnosis.
4. Chauffeur’s panel is broad in navigation but incomplete in data and behavior.
5. Chauffeur’s configuration is less portable and its `env` contract is unfinished.
6. Chauffeur’s recovery guarantees are not applied consistently to destructive operations.
7. Chauffeur’s verification does not yet enforce a real site HTTP success criterion.
